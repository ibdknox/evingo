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
	"unicode/utf8"
  "unicode"
)

//-----------------------------------------------------
// Tokens
//-----------------------------------------------------

type TokenType string

type Token struct {
	tokenType TokenType
	value     string
	line      int
	offset      int
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
	END                     = "END"
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
	"end":    END,
	"and":    AND,
	"or":     OR,
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

func Lex(str string) []Token {
	scanner := NewScanner(str)
	var tokens []Token
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
			tokens = append(tokens, Token{STRING, str, line, offset + 1})
      scanner.read()
			// eat the string
		case isSpecialChar(char):
			scanner.read()
			tokens = append(tokens, Token{specials[char], string(char), line, offset})
		case isDigitChar(char):
			str := scanner.eatWhile(isDigitChar)
			tokens = append(tokens, Token{NUMBER, string(str), line, offset})
    case char == '-':
      scanner.read()
      next, nextOk := scanner.peek()
      if nextOk && isDigitChar(next) {
        str := "-" + scanner.eatWhile(isDigitChar)
        tokens = append(tokens, Token{NUMBER, string(str), line, offset})
      } else {
        str := "-" + scanner.eatWhile(isIdentifierChar)
        curType = IDENTIFIER
        tokens = append(tokens, Token{curType, str, line, offset})
      }
		case isIdentifierChar(char):
			str := scanner.eatWhile(isIdentifierChar)
			curType = IDENTIFIER
			if isKeyword(str) {
				curType, _ = keywords[str]
			}
			tokens = append(tokens, Token{curType, str, line, offset})
		}
	}
	return tokens
}

type tokenIterator struct {
  pos int
  tokens []Token
}

func (iter *tokenIterator) peek() Token {
  return iter.tokens[iter.pos + 1]
}

func (iter *tokenIterator) read() Token {
  iter.pos++
  return iter.tokens[iter.pos]
}

func (iter *tokenIterator) unread() Token {
  iter.pos--
  return iter.tokens[iter.pos]
}

//-----------------------------------------------------
// Parsing
//-----------------------------------------------------

type nodeType string

const (
  CODE_CONTEXT_NODE nodeType = "CODE_CONTEXT"
  QUERY_NODE nodeType = "QUERY"
  SCAN_NODE = "SCAN"
  ADD_NODE = "ADD"
  REMOVE_NODE = "REMOVE"
  EXPRESSION_NODE = "EXPRESSION"
  CHOOSE_NODE = "CHOOSE"
  UNION_NODE = "UNION"
  UNKNOWN_NODE = "UNKNOWN"
)

type node struct {
  nodeType nodeType
  info map[string]string
  line int
  offset int
}

func newNode(nodeType nodeType, line int, offset int) node {
  var info = make(map[string]string)
  return node{nodeType, info, line, offset}
}

type line struct {
  parent *line
  children []*line
  tokens []Token
  rootNode node
  line int
  offset int
}

func (t line) String() string {
  indent := ""
  // for i := 0; i < t.offset + 1; i++ {
  //   indent += " "
  // }
  childLines := ""
  for _, child := range t.children {
    childLines += child.String()
  }
	return fmt.Sprintf("\n%s %v %v%s", indent, t.line, tokensToString(t.tokens), childLines)
}

func newLine(parent *line, offset int, tokens []Token) line {
  lineNum := -1
  if len(tokens) > 0 {
    lineNum = tokens[0].line
  }
  rootNode := newNode(UNKNOWN_NODE, lineNum, offset)
  return line{parent, make([]*line,0), tokens, rootNode, lineNum, offset}
}

func tokensToString(tokens []Token) string {
  var bytes bytes.Buffer
  prevEnd := 0
  for _, token := range tokens {
    for i := 0; i < token.offset - prevEnd; i++ {
      bytes.WriteRune(' ')
    }
    bytes.WriteString(token.value)
    prevEnd = token.offset + len(token.value)
  }
  return bytes.String()
}

func latestContext(context []node) node {
  var latestContext node
  if len(context) > 0 {
    latestContext = context[len(context)-1]
  }
  return latestContext
}

func parseRootLevel(tokens []Token, queries []node, context[]node) []node {
  latestContext := latestContext(context)
  var query node
  line := tokensToString(tokens)
  fmt.Println("header:", line)
  var newContext []node
  if latestContext.nodeType != QUERY_NODE {
    query = newNode(QUERY_NODE, tokens[0].line, tokens[0].offset)
    query.info["name"] = line;
    queries = append(queries, query)
    newContext = append(newContext, query)
  } else if latestContext.nodeType == QUERY_NODE {
    // if we're already in a query node and at the top level, then we must be
    // adding to the name
    query = latestContext
    query.info["name"] += "\n" + line
    newContext = context
  }
  fmt.Println("Query: %v", query)
  return newContext
}

func parseLine(tokens []Token, queries []node, context []node) []node {
  latestContext := latestContext(context)
  currentToken := tokens[0]
  if currentToken.offset == 0 {
    return parseRootLevel(tokens, queries, context)
  }
  if latestContext.nodeType == "" {
    fmt.Printf(color.Error("Programs should start without leading indentation, line %v should be unindented.\n"), tokens[0].line)
  }
  return context
}

func ParseTokens(tokens []Token, info map[string]string) {
	var token Token
  // var queries []node
  // var context []node
  var codeContext = newLine(nil, -1, make([]Token,0))
  codeContext.rootNode.nodeType = CODE_CONTEXT_NODE
  codeContext.rootNode.info = info
  parentLine := &codeContext
	tokenLen := len(tokens)
	for ix := 0; ix < tokenLen; ix++ {
		startIx := ix
		token = tokens[ix]
		line := token.line
		for ix < tokenLen-1 && tokens[ix+1].line == line {
			ix++
			token = tokens[ix]
		}
    lineTokens := tokens[startIx:ix+1]
    indent := lineTokens[0].offset
    for parentLine != &codeContext && parentLine.offset >= indent {
      parentLine = parentLine.parent
    }
    currentLine := newLine(parentLine, indent, lineTokens)
    fmt.Println("Parent", parentLine)
    fmt.Println("Child", currentLine)
    parentLine.children = append(parentLine.children, &currentLine)
    parentLine = &currentLine
    // context = parseLine(, queries, context)
    fmt.Printf("Line tree: %v\n", codeContext)
		fmt.Printf("%v\n\n", tokens[startIx:ix+1])
	}
  fmt.Printf("Line tree: %v", codeContext)
}

func ParseString(code string) {
		tokens := Lex(code)
    info := make(map[string]string)
    info["sourceType"] = "string"
		ParseTokens(tokens, info)
}

func ParseFile(path string) {
	content, err := ioutil.ReadFile(path)
	if err == nil {
		code := string(content)
		tokens := Lex(code)
    info := make(map[string]string)
    info["sourceType"] = "file"
    info["file"] = path
		ParseTokens(tokens, info)
	} else {
		fmt.Printf(color.Error("Couldn't read file: %v"), path)
	}
}
