package parser

import (
	"fmt"

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

func (p *parser) nextToken() token {
	return p.lexer.token()
}

func (p *parser) formatError(format string, v ...interface{}) error {
	return errors.Errorf(
		"syntax error in %d:%d: %s",
		p.token.line, p.token.col, fmt.Sprintf(format, v...))
}

func (p *parser) parse() (*tree, error) {
	p.tree = newTree()

	if err := p.parseExpression(); err != nil {
		return nil, err
	}

	if p.token.Type != T_EOF {
		return nil, p.formatError("unexpected '%s'", p.token.Value)
	}

	return p.tree, nil
}

func (p *parser) parseExpression() error {
	if err := p.parseTerm(); err != nil {
		return err
	}

	for p.token.Type == T_LOGICAL_OR {
		orNode := newTree()
		orNode.value = p.token
		orNode.left = p.tree
		if err := p.parseTerm(); err != nil {
			return err
		}
		orNode.right = p.tree
		p.tree = orNode
	}

	return nil
}

func (p *parser) parseTerm() error {
	if err := p.parseFactor(); err != nil {
		return err
	}

	for p.token.Type == T_LOGICAL_AND {
		andNode := newTree()
		andNode.value = p.token
		andNode.left = p.tree
		if err := p.parseFactor(); err != nil {
			return err
		}
		andNode.right = p.tree
		p.tree = andNode
	}

	return nil
}

func (p *parser) parseFactor() error {
	p.token = p.nextToken()

	switch p.token.Type {
	case T_BOOLEAN:
		p.tree = newTree()
		p.tree.value = p.token
		p.token = p.nextToken()

	case T_IDENTIFIER:
		p.tree = newTree()
		identifierToken := p.token
		p.token = p.nextToken()
		switch p.token.Type {
		case T_IS_EQUAL, T_IS_NOT_EQUAL, T_MATCHES, T_NOT_MATCHES:
			p.tree.value = p.token
			p.tree.left = newTree()
			p.tree.left.value = identifierToken
			p.token = p.nextToken()
			if p.token.Type == T_STRING {
				p.tree.right = newTree()
				p.tree.right.value = p.token
				p.token = p.nextToken()
			} else {
				return p.formatError("string value expected")
			}
		default:
			return p.formatError("comparison operator expected")
		}

	case T_LOGICAL_NOT:
		notNode := newTree()
		notNode.value = p.token
		if err := p.parseFactor(); err != nil {
			return err
		}
		notNode.left = p.tree
		p.tree = notNode

	case T_LEFT_PAREN:
		if err := p.parseExpression(); err != nil {
			return err
		}
		if p.token.Type == T_RIGHT_PAREN {
			p.token = p.nextToken()
		} else {
			return p.formatError("missing right parenthesis")
		}

	case T_RIGHT_PAREN:
		return p.formatError("unexpected right parenthesis")

	case T_EOF:
		return p.formatError("unexpected EOF")

	default:
		return p.formatError("unexpected '%s'", p.token.Value)
	}

	return nil
}

func newParser(l *lexer) *parser {
	return &parser{
		lexer: l,
	}
}

func Parse(s string) (Tree, error) {
	return newParser(newLexer(s)).parse()
}
