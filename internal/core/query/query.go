package query

import (
	"github.com/pkg/errors"

	"github.com/carlosabalde/pgp-tomb/internal/core/query/parser"
)

// This interface represents a tree node. There are several implementations of
// the interface in this package, but one may define custom Query's as long as
// they implement the 'Eval' function.
type Query interface {
	Eval(Params) bool
}

// Defines the interface needed by 'Query' in order to be able to evaluate
// expressions.
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

func visit(tree parser.Tree) (Query, error) {
	return nil, nil
}
