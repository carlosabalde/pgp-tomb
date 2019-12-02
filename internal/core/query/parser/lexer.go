package parser

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type token struct {
	Type  tokenType
	Value string
	line  int
	col   int
}

func (i token) String() string {
	return fmt.Sprintf("%s:%q", i.Type, i.Value)
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

func (i tokenType) String() string {
	s := tokenName[i]
	if s == "" {
		return fmt.Sprintf("T_UNKNOWN_%d", int(i))
	}
	return s
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
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// Steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// Returns the string consumed by the lexer after the last emit.
func (l *lexer) buffer() string {
	return l.input[l.start:l.pos]
}

// Passes an token back to the client.
func (l *lexer) emit(t tokenType) {
	l.tokens <- token{
		t,
		l.buffer(),
		l.lineNum(),
		l.columnNum(),
	}
	l.start = l.pos
}

// Skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// Reports which line we're on. Doing it this way means we don't have to worry
// about peek double counting.
func (l *lexer) lineNum() int {
	return 1 + strings.Count(l.input[:l.pos], "\n")
}

// Reports the character of the current line we're on.
func (l *lexer) columnNum() int {
	if lf := strings.LastIndex(l.input[:l.pos], "\n"); lf != -1 {
		return len(l.input[lf+1 : l.pos])
	}
	return len(l.input[:l.pos])
}

// Returns an error token and terminates the scan by passing back a nil pointer
// that will be the next state, terminating l.token.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- token{
		T_ERR,
		fmt.Sprintf(format, args...),
		l.lineNum(),
		l.columnNum(),
	}
	return nil
}

// Returns the next token from the input.
func (l *lexer) token() token {
	for {
		select {
		case t := <-l.tokens:
			return t
		default:
			l.state = l.state(l)
		}
	}
}

// Creates a new scanner for the input string.
func newLexer(input string) *lexer {
	l := &lexer{
		input:  input,
		tokens: make(chan token, 8),
	}
	l.state = stateInit
	return l
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
