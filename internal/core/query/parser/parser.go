package parser

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// Grammar:
//   <expression> ::= <term>{<or><term>}
//   <term>       ::= <factor>{<and><factor>}
//   <factor>     ::= <boolean>|<comparison>|<not><factor>|(<expression>)
//   <comparison> ::= <identifier><operator><string>
//   <boolean>    ::= 'false'|'true'
//   <or>         ::= '||'
//   <and>        ::= '&&'
//   <not>        ::= '!'
//   <operator>   ::= '=='|'!='|'~'|'!~'

type parser struct {
	lexer        *lexer
	currentNode  *tree
	currentToken token
}

func (self *parser) nextToken() token {
	return self.lexer.nextToken()
}

func (self *parser) formatError(format string, v ...interface{}) error {
	return errors.Errorf(
		"syntax error in %d:%d: %s",
		self.currentToken.line, self.currentToken.column, fmt.Sprintf(format, v...))
}

func (self *parser) parse() (*tree, error) {
	self.currentNode = newTree()

	if err := self.parseExpression(); err != nil {
		return nil, err
	}

	if self.currentToken.Type != T_EOF {
		return nil, self.formatError("unexpected '%s'", self.currentToken.Value)
	}

	return self.currentNode, nil
}

func (self *parser) parseExpression() error {
	if err := self.parseTerm(); err != nil {
		return err
	}

	for self.currentToken.Type == T_LOGICAL_OR {
		orNode := newTree()
		orNode.value = self.currentToken
		orNode.left = self.currentNode
		if err := self.parseTerm(); err != nil {
			return err
		}
		orNode.right = self.currentNode
		self.currentNode = orNode
	}

	return nil
}

func (self *parser) parseTerm() error {
	if err := self.parseFactor(); err != nil {
		return err
	}

	for self.currentToken.Type == T_LOGICAL_AND {
		andNode := newTree()
		andNode.value = self.currentToken
		andNode.left = self.currentNode
		if err := self.parseFactor(); err != nil {
			return err
		}
		andNode.right = self.currentNode
		self.currentNode = andNode
	}

	return nil
}

func (self *parser) parseFactor() error {
	self.currentToken = self.nextToken()

	switch self.currentToken.Type {
	case T_BOOLEAN:
		self.currentNode = newTree()
		self.currentNode.value = self.currentToken
		self.currentToken = self.nextToken()

	case T_IDENTIFIER:
		if self.currentToken.Value != "uri" &&
			!strings.HasPrefix(self.currentToken.Value, "tags.") {
			return self.formatError("invalid identifier '%s'", self.currentToken.Value)
		}
		self.currentNode = newTree()
		identifierToken := self.currentToken
		self.currentToken = self.nextToken()
		switch self.currentToken.Type {
		case T_IS_EQUAL, T_IS_NOT_EQUAL, T_MATCHES, T_NOT_MATCHES:
			self.currentNode.value = self.currentToken
			self.currentNode.left = newTree()
			self.currentNode.left.value = identifierToken
			self.currentToken = self.nextToken()
			if self.currentToken.Type == T_STRING {
				self.currentNode.right = newTree()
				self.currentNode.right.value = self.currentToken
				self.currentToken = self.nextToken()
			} else {
				return self.formatError("string value expected")
			}
		default:
			return self.formatError("comparison operator expected")
		}

	case T_LOGICAL_NOT:
		notNode := newTree()
		notNode.value = self.currentToken
		if err := self.parseFactor(); err != nil {
			return err
		}
		notNode.left = self.currentNode
		self.currentNode = notNode

	case T_LEFT_PARENTHESES:
		if err := self.parseExpression(); err != nil {
			return err
		}
		if self.currentToken.Type == T_RIGHT_PARENTHESES {
			self.currentToken = self.nextToken()
		} else {
			return self.formatError("missing right parenthesis")
		}

	case T_RIGHT_PARENTHESES:
		return self.formatError("unexpected right parenthesis")

	case T_EOF:
		return self.formatError("unexpected EOF")

	default:
		return self.formatError("unexpected '%s'", self.currentToken.Value)
	}

	return nil
}

func newParser(l *lexer) *parser {
	return &parser{
		lexer: l,
	}
}

func Parse(str string) (Tree, error) {
	return newParser(newLexer(str)).parse()
}
