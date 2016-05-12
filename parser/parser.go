package parser

import (
	"bytes"
	"fmt"
	"github.com/witheve/evingo/util/color"
	"io/ioutil"
	"unicode/utf8"
  "unicode"
)

type TokenType string

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

type Token struct {
	tokenType TokenType
	value     string
	line      int
	char      int
}

func (t Token) String() string {
	return fmt.Sprintf("{%v %v line %v ch %v}", t.tokenType, t.value, t.line, t.char)
}

func isWhiteSpace(ch rune) bool {
	return ch == ' ' || ch == '\n' || ch == '\t'
}

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

func isSpecialChar(ch rune) bool {
	_, found := specials[ch]
	return found
}

func isIdentifierChar(ch rune) bool {
	return !isWhiteSpace(ch) && !isSpecialChar(ch)
}

var keywords = map[string]TokenType{
	"union":  UNION,
	"choose": CHOOSE,
	"end":    END,
	"and":    AND,
	"or":     OR,
}

func isKeyword(str string) bool {
	_, found := keywords[str]
	return found
}

func isStringChar(ch rune) bool {
	return ch == '"'
}

func stringEater(ch rune, state *scanState) bool {
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

type Scanner struct {
	str        string
	line       int
	offset     int
	byteOffset int
}

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

type ScanPredicate func(rune) bool

func (s *Scanner) eatWhile(pred ScanPredicate) string {
	var curString bytes.Buffer
	for char, ok := s.peek(); ok && pred(char); char, ok = s.peek() {
		char, ok = s.read()
		curString.WriteRune(char)
	}
	return curString.String()
}

type scanState struct {
	state string
}
type StatefulScanPredicate func(rune, *scanState) bool

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
			str := scanner.eatWhileState(stringEater)
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

func ParseTokens(tokens []Token) {
	var token Token
	tokenLen := len(tokens)
	for ix := 0; ix < tokenLen; ix++ {
		startIx := ix
		token = tokens[ix]
		line := token.line
		for ix < tokenLen-1 && tokens[ix+1].line == line {
			ix++
			token = tokens[ix]
		}
		fmt.Printf("line %v goes from %v to %v\n", line, startIx, ix)
		fmt.Printf("%v\n\n", tokens[startIx:ix+1])
	}
}

func ParseString(code string) {
		tokens := Lex(code)
		ParseTokens(tokens)
}

func ParseFile(path string) {
	content, err := ioutil.ReadFile(path)
	if err == nil {
		code := string(content)
    ParseString(code)
	} else {
		fmt.Printf(color.Error("Couldn't read file: %v"), path)
	}
}
