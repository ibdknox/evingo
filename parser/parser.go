//-----------------------------------------------------
// Parser
// This package takes a string or file and turns it
// into an Eve program graph
//-----------------------------------------------------

package parser

import (
	"bytes"
	"fmt"
	"github.com/witheve/evingo/util/color"
	"io/ioutil"
	"strings"
	"unicode"
	"unicode/utf8"
)

//-----------------------------------------------------
// Tokens
//-----------------------------------------------------

type TokenType string

type Token struct {
	tokenType TokenType
	value     string
	line      int
	offset    int
}

func (t Token) String() string {
	return fmt.Sprintf("{%v %v line %v ch %v}", t.tokenType, t.value, t.line, t.offset)
}

const (
	TAG           TokenType = "TAG"
	NAME                    = "NAME"
	DOT                     = "DOT"
	OPEN_PAREN              = "OPEN_PAREN"
	CLOSE_PAREN             = "CLOSE_PAREN"
	OPEN_BRACKET            = "OPEN_BRACKET"
	CLOSE_BRACKET           = "CLOSE_BRACKET"
	OPEN_CURLY              = "OPEN_CURLY"
	CLOSE_CURLY             = "CLOSE_CURLY"
	CHOOSE                  = "CHOOSE"
	UNION                   = "UNION"
	OR                      = "OR"
	AND                     = "AND"
	ADD                     = "ADD"
	REMOVE                  = "REMOVE"
	STRING                  = "STRING"
	NUMBER                  = "NUMBER"
	IDENTIFIER              = "IDENTIFIER"
)

var specials = map[rune]TokenType{
	'#': TAG,
	'@': NAME,
	'.': DOT,
	'[': OPEN_BRACKET,
	']': CLOSE_BRACKET,
	'(': OPEN_PAREN,
	')': CLOSE_PAREN,
	'{': OPEN_CURLY,
	'}': CLOSE_CURLY,
}

var keywords = map[string]TokenType{
	"union":  UNION,
	"choose": CHOOSE,
	"and":    AND,
	"or":     OR,
	"add":    ADD,
	"remove": REMOVE,
}

//-----------------------------------------------------
// Rune predicates
//-----------------------------------------------------

func isWhiteSpace(ch rune) bool {
	return ch == ' ' || ch == '\n' || ch == '\t'
}

func isSpecialChar(ch rune) bool {
	_, found := specials[ch]
	return found
}

func isIdentifierChar(ch rune) bool {
	return !isWhiteSpace(ch) && !isSpecialChar(ch)
}

func isKeyword(str string) bool {
	_, found := keywords[str]
	return found
}

func isStringChar(ch rune) bool {
	return ch == '"'
}

func isStringCharStateful(ch rune, state *scanState) bool {
	curState := state.state
	if ch == '\\' {
		state.state = "escaped"
		return true
	} else {
		state.state = "normal"
		return curState == "escaped" || ch != '"'
	}
}

func isDigitChar(ch rune) bool {
	return unicode.IsDigit(ch) || ch == '.'
}

//-----------------------------------------------------
// Scanner
//-----------------------------------------------------

type Scanner struct {
	str        string
	line       int
	offset     int
	byteOffset int
}

type scanState struct {
	state string
}

type ScanPredicate func(rune) bool
type StatefulScanPredicate func(rune, *scanState) bool

func (s *Scanner) _peek() (rune, bool, int) {
	char, width := utf8.DecodeRuneInString(s.str[s.byteOffset:])
	if char == utf8.RuneError {
		return char, false, width
	}
	return char, true, width
}

func (s *Scanner) peek() (rune, bool) {
	char, ok, _ := s._peek()
	return char, ok
}

func (s *Scanner) read() (rune, bool) {
	char, ok, width := s._peek()
	switch char {
	case utf8.RuneError:
		return char, ok
	case '\n':
		s.line++
		s.offset = 0
		s.byteOffset += width
		return char, ok
	}
	s.offset++
	s.byteOffset += width
	return char, ok
}

func (s *Scanner) eatWhiteSpace() {
	for char, ok := s.peek(); ok && isWhiteSpace(char); char, ok = s.peek() {
		s.read()
	}
}

func (s *Scanner) eatWhile(pred ScanPredicate) string {
	var curString bytes.Buffer
	for char, ok := s.peek(); ok && pred(char); char, ok = s.peek() {
		char, ok = s.read()
		curString.WriteRune(char)
	}
	return curString.String()
}

func (s *Scanner) eatWhileState(pred StatefulScanPredicate) string {
	var curString bytes.Buffer
	var state scanState
	for char, ok := s.peek(); ok && pred(char, &state); char, ok = s.peek() {
		char, ok = s.read()
		curString.WriteRune(char)
	}
	return curString.String()
}

func NewScanner(str string) *Scanner {
	return &Scanner{str, 1, 0, 0}
}

//-----------------------------------------------------
// Lexing
//-----------------------------------------------------

func Lex(str string) []*Token {
	scanner := NewScanner(str)
	var tokens []*Token
	var curType TokenType
	for char, ok := scanner.peek(); ok; char, ok = scanner.peek() {
		line := scanner.line
		offset := scanner.offset
		switch {
		case isWhiteSpace(char):
			scanner.eatWhiteSpace()
		case isStringChar(char):
			scanner.read()
			str := scanner.eatWhileState(isStringCharStateful)
			tokens = append(tokens, &Token{STRING, str, line, offset + 1})
			scanner.read()
			// eat the string
		case isSpecialChar(char):
			scanner.read()
			tokens = append(tokens, &Token{specials[char], string(char), line, offset})
		case isDigitChar(char):
			str := scanner.eatWhile(isDigitChar)
			tokens = append(tokens, &Token{NUMBER, string(str), line, offset})
		case char == '-':
			scanner.read()
			next, nextOk := scanner.peek()
			if nextOk && isDigitChar(next) {
				str := "-" + scanner.eatWhile(isDigitChar)
				tokens = append(tokens, &Token{NUMBER, string(str), line, offset})
			} else {
				str := "-" + scanner.eatWhile(isIdentifierChar)
				curType = IDENTIFIER
				tokens = append(tokens, &Token{curType, str, line, offset})
			}
		case isIdentifierChar(char):
			str := scanner.eatWhile(isIdentifierChar)
			curType = IDENTIFIER
			if isKeyword(str) {
				curType, _ = keywords[str]
			}
			tokens = append(tokens, &Token{curType, str, line, offset})
		}
	}
	return tokens
}

type tokenIterator struct {
	pos    int
	tokens []*Token
}

func (iter *tokenIterator) peek() (*Token, bool) {
	if iter.pos+1 < len(iter.tokens) {
		return nil, false
	}
	return iter.tokens[iter.pos+1], true
}

func (iter *tokenIterator) read() (*Token, bool) {
	if iter.pos+1 < len(iter.tokens) {
		iter.pos += 1
		return iter.tokens[iter.pos], true
	}
	return nil, false
}

func (iter *tokenIterator) unread() (*Token, bool) {
	if iter.pos-1 > -1 {
		iter.pos -= 1
		return iter.tokens[iter.pos], true
	}
	return nil, false
}

func newTokenIterator(tokens []*Token) tokenIterator {
	return tokenIterator{-1, tokens}
}

//-----------------------------------------------------
// Parsing
//-----------------------------------------------------

type nodeType string

const (
	CODE_CONTEXT_NODE nodeType = "CODE_CONTEXT"
	QUERY_NODE        nodeType = "QUERY"
	OBJECT_NODE                = "OBJECT"
	ADD_NODE                   = "ADD"
	REMOVE_NODE                = "REMOVE"
	EXPRESSION_NODE            = "EXPRESSION"
	BINDING_NODE               = "BINDING_NODE"
	VARIABLE_NODE              = "VARIABLE_NODE"
	CHOOSE_NODE                = "CHOOSE"
	UNION_NODE                 = "UNION"
	UNKNOWN_NODE               = "UNKNOWN"
)

type node struct {
	nodeType nodeType
	info     map[string]interface{}
	children []*node
	line     int
	offset   int
}

// CODE_CONTEXT_NODE
//    sourceType
//    source
//
// QUERY_NODE
//    name
//    variables map[string]*VARIABLE_NODE
//
// OBJECT_NODE
//    variable
//    attributes []*BINDING_NODE
//
// VARIABLE_NODE
//    name
//
// BINDING_NODE
//    variable
//    field string
//    source *node

func (curNode *node) String() string {
	childLines := ""
	for _, child := range curNode.children {
		childString := child.String()
		for _, line := range strings.Split(childString, "\n") {
			childLines += "  " + line + "\n"
		}
	}
	infoString := ""
	for key, value := range curNode.info {
		switch key {
		case "source":
			infoString += "    source: cycle\n"
		case "variable":
			infoString += fmt.Sprintf("    variable: %s\n", value.(*node).info["name"])
		case "variables":
			variables := value.(map[string]*node)
			infoString += "    variables: "
			for name, _ := range variables {
				infoString += fmt.Sprintf("%s, ", name)
			}
			infoString += "\n"
		case "name":
			infoString += fmt.Sprintf("    %s: %q\n", key, value)
		default:
			infoString += fmt.Sprintf("    %s: %s\n", key, value)
		}
	}
	return fmt.Sprintf("%v\n%s%s", curNode.nodeType, infoString, childLines)
}

func newNode(nodeType nodeType, line int, offset int) *node {
	var info = make(map[string]interface{})
	return &node{nodeType, info, make([]*node, 0), line, offset}
}

type line struct {
	parent   *line
	children []*line
	tokens   []*Token
	rootNode *node
	line     int
	offset   int
}

func (t *line) String() string {
	childLines := ""
	for _, child := range t.children {
		childLines += child.String()
	}
	return fmt.Sprintf("\n%v %v%s", t.line, tokensToString(t.tokens), childLines)
}

func newLine(parent *line, offset int, tokens []*Token) *line {
	lineNum := -1
	if len(tokens) > 0 {
		lineNum = tokens[0].line
	}
	rootNode := newNode(UNKNOWN_NODE, lineNum, offset)
	return &line{parent, make([]*line, 0), tokens, rootNode, lineNum, offset}
}

func tokensToString(tokens []*Token) string {
	var bytes bytes.Buffer
	prevEnd := 0
	for _, token := range tokens {
		for i := 0; i < token.offset-prevEnd; i++ {
			bytes.WriteRune(' ')
		}
		bytes.WriteString(token.value)
		prevEnd = token.offset + len(token.value)
	}
	return bytes.String()
}

func getParentQuery(line *line) *node {
	cur := line
	for cur.parent != nil {
		if cur.parent.rootNode.nodeType == QUERY_NODE {
			return cur.parent.rootNode
		}
		cur = cur.parent
	}
	return nil
}

func setChildOnParentNode(line *line) {
	parentNode := line.parent.rootNode
	parentNode.children = append(parentNode.children, line.rootNode)
}

func parseQueryLine(line *line) {
	tokens := line.tokens
	lineString := tokensToString(tokens)
	// check if this query line is actually just adding to the name
	// of the previous query, or if it's an entirely new query. We
	// can tell by if the lines are adjacent
	sibling, hasSibling := getSiblingLine(line)
	if !hasSibling || sibling.line+1 != line.line {
		// we are a totally new query
		curNode := line.rootNode
		curNode.nodeType = QUERY_NODE
		curNode.info["name"] = lineString
		curNode.info["variables"] = make(map[string]*node)
		curNode.line = tokens[0].line
		curNode.offset = tokens[0].offset
		setChildOnParentNode(line)
	} else {
		siblingNode := sibling.rootNode
		siblingNode.info["name"] = siblingNode.info["name"].(string) + "\n" + lineString
		line.rootNode = sibling.rootNode
	}
}

func newBinding(token *Token, source *node, field string, variable *node) *node {
	node := newNode(BINDING_NODE, token.line, token.offset)
	node.info["source"] = source
	node.info["field"] = field
	node.info["variable"] = variable
	return node
}

func newConstantBinding(token *Token, source *node, field string, constant interface{}, constantType string) *node {
	node := newNode(BINDING_NODE, token.line, token.offset)
	node.info["source"] = source
	node.info["field"] = field
	node.info["constant"] = constant
	node.info["constantType"] = constantType
	return node
}

func assignVariable(line *line, token *Token, name string) *node {
	// get the closest query to use as the variable cache
	query := getParentQuery(line)
	variables := query.info["variables"].(map[string]*node)
	if existing, ok := variables[name]; ok {
		return existing
	}
	variable := newNode(VARIABLE_NODE, token.line, token.offset)
	variable.info["name"] = name
	variables[name] = variable
	return variable
}

func parseObjectLine(line *line) {
	// query := getParentQuery(line)
	iter := newTokenIterator(line.tokens)
	curNode := line.rootNode
	curNode.nodeType = OBJECT_NODE
	info := curNode.info
	setChildOnParentNode(line)
	mode := "variable"
	var nameToken *Token
	for token, ok := iter.read(); ok; token, ok = iter.read() {
		if mode == "variable" {
			switch token.tokenType {
			case TAG:
				//create a binding for tag
				value, ok := iter.read()
				if !ok {
					fmt.Println(color.Error("Naked # on line %v"), token.line)
					continue
				}
				curNode.children = append(curNode.children, newConstantBinding(value, curNode, "tag", value.value, "string"))
				if nameToken == nil {
					nameToken = value
				}
			case NAME:
				//create a binding for name
				value, ok := iter.read()
				if !ok {
					fmt.Println(color.Error("Naked @ on line %v"), token.line)
					continue
				}
				curNode.children = append(curNode.children, newConstantBinding(value, curNode, "name", value.value, "string"))
				if nameToken == nil {
					nameToken = value
				}
			default:
				// what do we do here?
				// mode = "attributes"
			}
		}
		fmt.Println("token: ", iter.pos, token)
	}
	if nameToken == nil {
		fmt.Println(color.Error("Object query without any naming # or @, %v"), line.line)
	} else {
		variable := assignVariable(line, nameToken, nameToken.value)
		info["variable"] = variable
	}
}

func parseAttributeLine(line *line) {
	fmt.Println("PARSING ATTRIBUTE LINE", line)
	// the first token should be an identifier
	iter := newTokenIterator(line.tokens)
	field, ok := iter.read()
	if !ok {
		fmt.Println(color.Error("Empty attribute on line %v"), line.line)
	}
	// possible cases
	//  attr
	//  attr: constant
	//  attr: var
	//  attr: some-expression
	//  attr = constant
	//  attr = var
	//  attr = some-expression
	op, ok := iter.read()
	fmt.Println("OP: ", op)
	switch {
	case !ok:
		// we're just binding the attribute to its own name
		// we need to look up if there's already a variable
		// and if not, get one
		variable := assignVariable(line, field, field.value)
		line.rootNode = newBinding(field, line.parent.rootNode, field.value, variable)
		fmt.Println("ADDING A BINDING!")
		setChildOnParentNode(line)
	case op.value == ":":
		fallthrough
	case op.value == "=":
		rightSide, ok := iter.read()
		// @TODO it's technically ok to put the right-hand side of the expression on another line,
		// I'm not sure exactly how we should handle that
		if !ok {
			fmt.Println(color.Error("Equality without left-hand side on line %v"), line.line)
		}
		fmt.Println("EQUALITY ATTRIBUTE: ", rightSide)
		// @TODO: write a helper that parses cases where an expression is expected
		if rightSide.tokenType == NUMBER || rightSide.tokenType == STRING {
			constantType := "string"
			if rightSide.tokenType == NUMBER {
				constantType = "number"
			}
			line.rootNode = newConstantBinding(field, line.parent.rootNode, field.value, rightSide.value, constantType)
		} else {
			line.rootNode = newConstantBinding(field, line.parent.rootNode, field.value, "THIS SHOULD BE AN EXPRESSION", "string")
		}
		setChildOnParentNode(line)

	}
}

func parseMutationLine(line *line) {
	iter := newTokenIterator(line.tokens)
	curNode := line.rootNode
	mutator, _ := iter.read()
	switch mutator.tokenType {
	case ADD:
		curNode.nodeType = ADD_NODE
	case REMOVE:
		curNode.nodeType = REMOVE_NODE
	}
	//@TODO check for "forever"
	setChildOnParentNode(line)
}

func getSiblingLine(line *line) (*line, bool) {
	for ix, child := range line.parent.children {
		if child.line == line.line && ix > 0 {
			return line.parent.children[ix-1], true
		}
	}
	return newLine(nil, -1, make([]*Token, 0)), false
}

func parseLine(line *line) {
	parentNode := getLineNode(line.parent)
	parentType := parentNode.nodeType
	fmt.Println("PARSING", line.line, "PARENT", parentType)
	if parentType == CODE_CONTEXT_NODE {
		parseQueryLine(line)
	} else {
		iter := newTokenIterator(line.tokens)
		// otherwise we have to look at the first bits of the line and the parent
		// context to get a sense of what it's doing
		if parentType == OBJECT_NODE {
			//treat this as an attribute
			parseAttributeLine(line)
		}
		firstToken, _ := iter.read()
		if firstToken.tokenType == TAG || firstToken.tokenType == NAME {
			//@TODO: we need to handle #eavs specially
			parseObjectLine(line)
		}
		if firstToken.tokenType == ADD || firstToken.tokenType == REMOVE {
			parseMutationLine(line)
		}

		// if it's just an identifier and we're in the context of a selection

	}
}

func getLineNode(line *line) *node {
	// NOTE: if parseLine calls getLineNode on its children
	// before setting the nodeType it's possible for this
	// to loop infinitely. Before attempting to look at your
	// children, make sure you nodeType is set to something
	// other than UNKNOWN_NODE
	curType := line.rootNode.nodeType
	if curType == UNKNOWN_NODE {
		parseLine(line)
		fmt.Println("Node type: ", line.rootNode.nodeType)
	}
	return line.rootNode
}

func walkLinesAndParse(root *line) {
	getLineNode(root)
	for _, child := range root.children {
		walkLinesAndParse(child)
	}
}

func fullParseTree(root *line) *node {
	walkLinesAndParse(root)
	return root.rootNode
}

func ParseTokens(tokens []*Token, info map[string]interface{}) {
	var token *Token
	// var queries []node
	// var context []node
	var codeContext = newLine(nil, -1, make([]*Token, 0))
	codeContext.rootNode.nodeType = CODE_CONTEXT_NODE
	codeContext.rootNode.info = info
	parentLine := codeContext
	tokenLen := len(tokens)
	for ix := 0; ix < tokenLen; ix++ {
		startIx := ix
		token = tokens[ix]
		line := token.line
		for ix < tokenLen-1 && tokens[ix+1].line == line {
			ix++
			token = tokens[ix]
		}
		lineTokens := tokens[startIx : ix+1]
		indent := lineTokens[0].offset
		for parentLine != codeContext && parentLine.offset >= indent {
			parentLine = parentLine.parent
		}
		currentLine := newLine(parentLine, indent, lineTokens)
		fmt.Println("Parent", parentLine)
		fmt.Println("Child", currentLine)
		parentLine.children = append(parentLine.children, currentLine)
		parentLine = currentLine
		// context = parseLine(, queries, context)
		fmt.Printf("Line tree: %v\n\n\n", codeContext)
	}
	fmt.Printf("Parse nodes:\n\n%v\n\n", fullParseTree(codeContext))
}

func ParseString(code string) {
	tokens := Lex(code)
	info := make(map[string]interface{})
	info["sourceType"] = "string"
	ParseTokens(tokens, info)
}

func ParseFile(path string) {
	content, err := ioutil.ReadFile(path)
	if err == nil {
		code := string(content)
		tokens := Lex(code)
		info := make(map[string]interface{})
		info["sourceType"] = "file"
		info["file"] = path
		ParseTokens(tokens, info)
	} else {
		fmt.Printf(color.Error("Couldn't read file: %v"), path)
	}
}
