package parser

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type token struct {
	Type   tokenType
	Value  string
	line   int
	column int
}

func (self token) String() string {
	return fmt.Sprintf("%s:%q", self.Type, self.Value)
}

type tokenType int

const (
	T_ERR tokenType = iota
	T_EOF

	T_IDENTIFIER

	T_STRING
	T_BOOLEAN

	T_LOGICAL_AND
	T_LOGICAL_OR
	T_LOGICAL_NOT

	T_LEFT_PAREN
	T_RIGHT_PAREN

	T_IS_EQUAL
	T_IS_NOT_EQUAL
	T_MATCHES
	T_NOT_MATCHES
)

var tokenName = map[tokenType]string{
	T_ERR:          "T_ERR",
	T_EOF:          "T_EOF",
	T_IDENTIFIER:   "T_IDENTIFIER",
	T_STRING:       "T_STRING",
	T_BOOLEAN:      "T_BOOLEAN",
	T_LOGICAL_AND:  "T_LOGICAL_AND",
	T_LOGICAL_OR:   "T_LOGICAL_OR",
	T_LOGICAL_NOT:  "T_LOGICAL_NOT",
	T_LEFT_PAREN:   "T_LEFT_PAREN",
	T_RIGHT_PAREN:  "T_RIGHT_PAREN",
	T_IS_EQUAL:     "T_IS_EQUAL",
	T_IS_NOT_EQUAL: "T_IS_NOT_EQUAL",
	T_MATCHES:      "T_MATCHES",
	T_NOT_MATCHES:  "T_NOT_MATCHES",
}

func (self tokenType) String() string {
	result := tokenName[self]
	if result == "" {
		return fmt.Sprintf("T_UNKNOWN_%d", int(self))
	}
	return result
}

const eof = -1

// Represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// Holds the state of the scanner.
type lexer struct {
	name   string     // The name of the input; used only for error reports.
	input  string     // The string being scanned.
	state  stateFn    // The next lexing function to enter.
	pos    int        // Current position in the input.
	start  int        // Start position of this token.
	width  int        // Width of last rune read from input.
	tokens chan token // Channel of scanned tokens.
}

// Returns the next rune in the input.
func (self *lexer) next() (result rune) {
	if self.pos >= len(self.input) {
		self.width = 0
		return eof
	}
	result, self.width = utf8.DecodeRuneInString(self.input[self.pos:])
	self.pos += self.width
	return result
}

// Steps back one rune. Can only be called once per call of next.
func (self *lexer) backup() {
	self.pos -= self.width
}

// Returns the string consumed by the lexer after the last emit.
func (self *lexer) buffer() string {
	return self.input[self.start:self.pos]
}

// Passes an token back to the client.
func (self *lexer) emit(t tokenType) {
	self.tokens <- token{
		t,
		self.buffer(),
		self.lineNum(),
		self.columnNum(),
	}
	self.start = self.pos
}

// Skips over the pending input before this point.
func (self *lexer) ignore() {
	self.start = self.pos
}

// Reports which line we're on. Doing it this way means we don't have to worry
// about peek double counting.
func (self *lexer) lineNum() int {
	return 1 + strings.Count(self.input[:self.pos], "\n")
}

// Reports the character of the current line we're on.
func (self *lexer) columnNum() int {
	if lf := strings.LastIndex(self.input[:self.pos], "\n"); lf != -1 {
		return len(self.input[lf+1 : self.pos])
	}
	return len(self.input[:self.pos])
}

// Returns an error token and terminates the scan by passing back a nil pointer
// that will be the next state, terminating l.token.
func (self *lexer) errorf(format string, args ...interface{}) stateFn {
	self.tokens <- token{
		T_ERR,
		fmt.Sprintf(format, args...),
		self.lineNum(),
		self.columnNum(),
	}
	return nil
}

// Returns the next token from the input.
func (self *lexer) token() token {
	for {
		select {
		case t := <-self.tokens:
			return t
		default:
			self.state = self.state(self)
		}
	}
}

// Creates a new scanner for the input string.
func newLexer(input string) *lexer {
	return &lexer{
		input:  input,
		tokens: make(chan token, 8),
		state:  stateInit,
	}
}

////////////////////////////////////////////////////////////////////////////////
// STATE FUNCTIONS
////////////////////////////////////////////////////////////////////////////////

// The initial state of the lexer.
func stateInit(l *lexer) stateFn {
	switch r := l.next(); {
	case isWhitespace(r):
		l.ignore()
		return stateInit
	case isAlphanumeric(r):
		return stateIdentifier
	case isOperator(r):
		return stateOperator
	case r == '\'':
		return stateSingleQuote
	case r == '"':
		return stateDoubleQuote
	case r == '(':
		l.emit(T_LEFT_PAREN)
		return stateInit
	case r == ')':
		l.emit(T_RIGHT_PAREN)
		return stateInit
	}
	return stateEnd
}

// The final state of the lexer. After this state is entered no more tokens can
// be requested as it will result in a nil pointer dereference.
func stateEnd(l *lexer) stateFn {
	// Always end with EOF token. The parser will keep asking for tokens until
	// an T_EOF or T_ERR token are encountered.
	l.emit(T_EOF)

	return nil
}

// Scans an identifier from the input stream.
func stateIdentifier(l *lexer) stateFn {
loop:
	for {
		switch r := l.next(); {
		case isAlphanumeric(r):
		default:
			break loop
		}
	}

	l.backup()

	switch l.buffer() {
	case "true", "false":
		l.emit(T_BOOLEAN)
	default:
		l.emit(T_IDENTIFIER)
	}

	return stateInit
}

// Scans an operator from the input stream.
func stateOperator(l *lexer) stateFn {
	r := l.next()
	for isOperator(r) {
		r = l.next()
	}

	l.backup()

	switch l.buffer() {
	case "!":
		l.emit(T_LOGICAL_NOT)
	case "&&":
		l.emit(T_LOGICAL_AND)
	case "||":
		l.emit(T_LOGICAL_OR)
	case "==":
		l.emit(T_IS_EQUAL)
	case "!=":
		l.emit(T_IS_NOT_EQUAL)
	case "~":
		l.emit(T_MATCHES)
	case "!~":
		l.emit(T_NOT_MATCHES)
	}

	return stateInit
}

// Scans an identifier enclosed in single quotes from the input stream.
func stateSingleQuote(l *lexer) stateFn {
	return stateQuote(l, '\'')
}

// Scans a string enclosed in double quotes from the input stream.
func stateDoubleQuote(l *lexer) stateFn {
	return stateQuote(l, '"')
}

////////////////////////////////////////////////////////////////////////////////
// HELPERS
////////////////////////////////////////////////////////////////////////////////

func stateQuote(l *lexer, quote rune) stateFn {
	l.ignore()
loop:
	for {
		switch l.next() {
		case quote:
			l.backup()
			l.emit(T_STRING)
			l.next()
			l.ignore()
			break loop
		case eof:
			return l.errorf("unexpected EOF")
		}
	}

	return stateInit
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

func isAlphanumeric(r rune) bool {
	return r == '_' || r == '.' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isOperator(r rune) bool {
	return r == '=' || r == '~' || r == '!' || r == '&' || r == '|'
}
