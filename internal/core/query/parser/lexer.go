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
	T_ERROR tokenType = iota
	T_EOF

	T_IDENTIFIER

	T_STRING
	T_BOOLEAN

	T_LOGICAL_AND
	T_LOGICAL_OR
	T_LOGICAL_NOT

	T_LEFT_PARENTHESES
	T_RIGHT_PARENTHESES

	T_IS_EQUAL
	T_IS_NOT_EQUAL
	T_MATCHES
	T_NOT_MATCHES
)

var tokenNames = map[tokenType]string{
	T_ERROR:             "T_ERROR",
	T_EOF:               "T_EOF",
	T_IDENTIFIER:        "T_IDENTIFIER",
	T_STRING:            "T_STRING",
	T_BOOLEAN:           "T_BOOLEAN",
	T_LOGICAL_AND:       "T_LOGICAL_AND",
	T_LOGICAL_OR:        "T_LOGICAL_OR",
	T_LOGICAL_NOT:       "T_LOGICAL_NOT",
	T_LEFT_PARENTHESES:  "T_LEFT_PARENTHESES",
	T_RIGHT_PARENTHESES: "T_RIGHT_PARENTHESES",
	T_IS_EQUAL:          "T_IS_EQUAL",
	T_IS_NOT_EQUAL:      "T_IS_NOT_EQUAL",
	T_MATCHES:           "T_MATCHES",
	T_NOT_MATCHES:       "T_NOT_MATCHES",
}

func (self tokenType) String() string {
	result := tokenNames[self]
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
	input         string     // The string being scanned.
	state         stateFn    // The next lexing function to enter.
	position      int        // Current position in the input.
	tokenStart    int        // Start position of this token.
	lastRuneWidth int        // Width of last rune read from input.
	tokens        chan token // Channel of scanned tokens.
}

// Returns the next rune in the input.
func (self *lexer) nextRune() (result rune) {
	if self.position >= len(self.input) {
		self.lastRuneWidth = 0
		return eof
	}
	result, self.lastRuneWidth = utf8.DecodeRuneInString(self.input[self.position:])
	self.position += self.lastRuneWidth
	return result
}

// Steps back one rune. Can only be called once per call of next!
func (self *lexer) unreadLastRune() {
	self.position -= self.lastRuneWidth
}

// Returns the string consumed by the lexer after the last emit.
func (self *lexer) bufferSinceLastEmit() string {
	return self.input[self.tokenStart:self.position]
}

// Passes an token back to the client.
func (self *lexer) emitToken(t tokenType) {
	self.tokens <- token{
		t,
		self.bufferSinceLastEmit(),
		self.lineNummber(),
		self.columnNumber(),
	}
	self.tokenStart = self.position
}

// Skips over the pending input before this point.
func (self *lexer) ignore() {
	self.tokenStart = self.position
}

// Reports which line we're on. Doing it this way means we don't have to worry
// about peek double counting.
func (self *lexer) lineNummber() int {
	return 1 + strings.Count(self.input[:self.position], "\n")
}

// Reports the character of the current line we're on.
func (self *lexer) columnNumber() int {
	if lf := strings.LastIndex(self.input[:self.position], "\n"); lf != -1 {
		return len(self.input[lf+1 : self.position])
	}
	return len(self.input[:self.position])
}

// Returns an error token and terminates the scan by passing back a nil pointer
// that will be the next state.
func (self *lexer) emitErrorToken(format string, args ...interface{}) stateFn {
	self.tokens <- token{
		T_ERROR,
		fmt.Sprintf(format, args...),
		self.lineNummber(),
		self.columnNumber(),
	}
	return nil
}

// Returns the next token from the input.
func (self *lexer) nextToken() token {
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
	switch r := l.nextRune(); {
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
		l.emitToken(T_LEFT_PARENTHESES)
		return stateInit
	case r == ')':
		l.emitToken(T_RIGHT_PARENTHESES)
		return stateInit
	}
	return stateEnd
}

// The final state of the lexer. After this state is entered no more tokens can
// be requested as it will result in a nil pointer dereference.
func stateEnd(l *lexer) stateFn {
	// Always end with EOF token. The parser will keep asking for tokens until
	// an T_EOF or T_ERROR token are encountered.
	l.emitToken(T_EOF)

	return nil
}

// Scans an identifier from the input stream.
func stateIdentifier(l *lexer) stateFn {
loop:
	for {
		switch r := l.nextRune(); {
		case isAlphanumeric(r):
		default:
			break loop
		}
	}

	l.unreadLastRune()

	switch l.bufferSinceLastEmit() {
	case "true", "false":
		l.emitToken(T_BOOLEAN)
	default:
		l.emitToken(T_IDENTIFIER)
	}

	return stateInit
}

// Scans an operator from the input stream.
func stateOperator(l *lexer) stateFn {
	r := l.nextRune()
	for isOperator(r) {
		r = l.nextRune()
	}

	l.unreadLastRune()

	switch l.bufferSinceLastEmit() {
	case "!":
		l.emitToken(T_LOGICAL_NOT)
	case "&&":
		l.emitToken(T_LOGICAL_AND)
	case "||":
		l.emitToken(T_LOGICAL_OR)
	case "==":
		l.emitToken(T_IS_EQUAL)
	case "!=":
		l.emitToken(T_IS_NOT_EQUAL)
	case "~":
		l.emitToken(T_MATCHES)
	case "!~":
		l.emitToken(T_NOT_MATCHES)
	default:
		return l.emitErrorToken("unknown operator")
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
		switch l.nextRune() {
		case quote:
			l.unreadLastRune()
			l.emitToken(T_STRING)
			l.nextRune()
			l.ignore()
			break loop
		case eof:
			return l.emitErrorToken("unexpected EOF")
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
