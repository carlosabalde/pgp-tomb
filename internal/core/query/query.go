package query

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/carlosabalde/pgp-tomb/internal/core/query/parser"
)

// Representation of a tree node modeling a boolean query expression.
type Query interface {
	Eval(Context) bool
}

// Defines the interface needed by 'Query' in order to be able to evaluate
// query expressions.
type Context interface {
	GetIdentifier(string) string
}

func Parse(query string) (Query, error) {
	tree, err := parser.Parse(query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse query")
	}
	return visit(tree)
}

func visit(tree parser.Tree) (Query, error) {
	token := tree.Value()

	switch token.Type {
	case parser.T_BOOLEAN:
		switch token.Value {
		case "true":
			return True, nil
		case "false":
			return False, nil
		default:
			return nil, errors.Errorf("unexpected boolean value '%s'", token.Value)
		}

	case parser.T_LOGICAL_AND, parser.T_LOGICAL_OR:
		left, err := visit(tree.Left())
		if err != nil {
			return nil, err
		}
		right, err := visit(tree.Right())
		if err != nil {
			return nil, err
		}
		if token.Type == parser.T_LOGICAL_AND {
			return And(left, right), nil
		}
		return Or(left, right), nil

	case parser.T_LOGICAL_NOT:
		left, err := visit(tree.Left())
		if err != nil {
			return nil, err
		}
		return Not(left), nil

	case parser.T_IS_EQUAL, parser.T_IS_NOT_EQUAL:
		if tree.Left().Value().Type != parser.T_IDENTIFIER ||
			tree.Right().Value().Type != parser.T_STRING {
			return nil, errors.New("invalid comparison")
		}
		identifier := tree.Left().Value().Value
		str := tree.Right().Value().Value
		query := Equal(identifier, str)
		if token.Type == parser.T_IS_EQUAL {
			return query, nil
		}
		return Not(query), nil

	case parser.T_MATCHES, parser.T_NOT_MATCHES:
		if tree.Left().Value().Type != parser.T_IDENTIFIER ||
			tree.Right().Value().Type != parser.T_STRING {
			return nil, errors.New("invalid comparison")
		}
		identifier := tree.Left().Value().Value
		regexp := tree.Right().Value().Value
		query, err := Match(identifier, regexp)
		if err != nil {
			return nil, err
		}
		if token.Type == parser.T_MATCHES {
			return query, nil
		}
		return Not(query), nil

	default:
		return nil, errors.Errorf("unexpected token %s", token.Type)
	}

	return nil, nil
}

func sprintf(format string, v ...interface{}) string {
	return fmt.Sprintf(format, v...)
}

func join(items []Query, sep string) string {
	s := make([]string, len(items))
	for i, item := range items {
		s[i] = sprintf("%s", item)
	}
	return strings.Join(s, sep)
}
