package query

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/carlosabalde/pgp-tomb/internal/core/query/parser"
)

// This interface represents a tree node.
type Query interface {
	Eval(Params) bool
}

// Defines the interface needed by 'Query' in order to be able to evaluate
// query expressions.
type Params interface {
	Get(string) string
}

// Simple implementation of 'Params' using a map of strings.
type Map map[string]string

func (m Map) Get(key string) string {
	return m[key]
}

func Parse(query string) (Query, error) {
	tree, err := parser.Parse(query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse query")
	}
	return visit(tree)
}

func visit(t parser.Tree) (Query, error) {
	token := t.Value()

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

	case parser.T_LOGICAL_AND:
		l, err := visit(t.Left())
		if err != nil {
			return nil, err
		}
		r, err := visit(t.Right())
		if err != nil {
			return nil, err
		}
		return And(l, r), nil

	case parser.T_LOGICAL_OR:
		l, err := visit(t.Left())
		if err != nil {
			return nil, err
		}
		r, err := visit(t.Right())
		if err != nil {
			return nil, err
		}
		return Or(l, r), nil

	case parser.T_LOGICAL_NOT:
		l, err := visit(t.Left())
		if err != nil {
			return nil, err
		}
		return Not(l), nil

	case parser.T_IS_EQUAL, parser.T_IS_NOT_EQUAL:
		if t.Left().Value().Type != parser.T_IDENTIFIER ||
			t.Right().Value().Type != parser.T_STRING {
			return nil, errors.New("invalid comparison")
		}
		identifier := t.Left().Value().Value
		str := t.Right().Value().Value
		query := Equal(identifier, str)
		if token.Type == parser.T_IS_EQUAL {
			return query, nil
		}
		return Not(query), nil

	case parser.T_MATCHES, parser.T_NOT_MATCHES:
		if t.Left().Value().Type != parser.T_IDENTIFIER ||
			t.Right().Value().Type != parser.T_STRING {
			return nil, errors.New("invalid comparison")
		}
		identifier := t.Left().Value().Value
		regexp := t.Right().Value().Value
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
