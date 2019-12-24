package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatch(t *testing.T) {
	context1 := Map{
		"foo": "42",
		"bar": "3.14",
		"baz": "",
	}

	tests := []struct {
		identifier string
		regexp     string
		result     bool
	}{
		{"foo", "^42$", true},
		{"foo", "^3.14$", false},
		{"foo", ".*", true},
		{"baz", "^$", true},
		{"baz", "^ $", false},
		{"quz", "^$", true},
	}

	for _, test := range tests {
		query, err := Match(test.identifier, test.regexp)
		if assert.NoError(t, err) {
			assert.Equal(t, query.Eval(context1), test.result)
		}
	}
}
