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
	lexer *lexer
	tree  *tree
	token token
}

func (self *parser) nextToken() token {
	return self.lexer.token()
}

func (self *parser) formatError(format string, v ...interface{}) error {
	return errors.Errorf(
		"syntax error in %d:%d: %s",
		self.token.line, self.token.column, fmt.Sprintf(format, v...))
}

func (self *parser) parse() (*tree, error) {
	self.tree = newTree()

	if err := self.parseExpression(); err != nil {
		return nil, err
	}

	if self.token.Type != T_EOF {
		return nil, self.formatError("unexpected '%s'", self.token.Value)
	}

	return self.tree, nil
}

func (self *parser) parseExpression() error {
	if err := self.parseTerm(); err != nil {
		return err
	}

	for self.token.Type == T_LOGICAL_OR {
		orNode := newTree()
		orNode.value = self.token
		orNode.left = self.tree
		if err := self.parseTerm(); err != nil {
			return err
		}
		orNode.right = self.tree
		self.tree = orNode
	}

	return nil
}

func (self *parser) parseTerm() error {
	if err := self.parseFactor(); err != nil {
		return err
	}

	for self.token.Type == T_LOGICAL_AND {
		andNode := newTree()
		andNode.value = self.token
		andNode.left = self.tree
		if err := self.parseFactor(); err != nil {
			return err
		}
		andNode.right = self.tree
		self.tree = andNode
	}

	return nil
}

func (self *parser) parseFactor() error {
	self.token = self.nextToken()

	switch self.token.Type {
	case T_BOOLEAN:
		self.tree = newTree()
		self.tree.value = self.token
		self.token = self.nextToken()

	case T_IDENTIFIER:
		if self.token.Value != "uri" &&
			!strings.HasPrefix(self.token.Value, "tags.") {
			return self.formatError("invalid identifier '%s'", self.token.Value)
		}
		self.tree = newTree()
		identifierToken := self.token
		self.token = self.nextToken()
		switch self.token.Type {
		case T_IS_EQUAL, T_IS_NOT_EQUAL, T_MATCHES, T_NOT_MATCHES:
			self.tree.value = self.token
			self.tree.left = newTree()
			self.tree.left.value = identifierToken
			self.token = self.nextToken()
			if self.token.Type == T_STRING {
				self.tree.right = newTree()
				self.tree.right.value = self.token
				self.token = self.nextToken()
			} else {
				return self.formatError("string value expected")
			}
		default:
			return self.formatError("comparison operator expected")
		}

	case T_LOGICAL_NOT:
		notNode := newTree()
		notNode.value = self.token
		if err := self.parseFactor(); err != nil {
			return err
		}
		notNode.left = self.tree
		self.tree = notNode

	case T_LEFT_PAREN:
		if err := self.parseExpression(); err != nil {
			return err
		}
		if self.token.Type == T_RIGHT_PAREN {
			self.token = self.nextToken()
		} else {
			return self.formatError("missing right parenthesis")
		}

	case T_RIGHT_PAREN:
		return self.formatError("unexpected right parenthesis")

	case T_EOF:
		return self.formatError("unexpected EOF")

	default:
		return self.formatError("unexpected '%s'", self.token.Value)
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
