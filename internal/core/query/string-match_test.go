package query

import (
	"fmt"
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
		string     string
	}{
		{"foo", "^42$", true, "(foo ~ '^42$')"},
		{"foo", "^3.14$", false, "(foo ~ '^3.14$')"},
		{"foo", ".*", true, "(foo ~ '.*')"},
		{"baz", "^$", true, "(baz ~ '^$')"},
		{"baz", "^ $", false, "(baz ~ '^ $')"},
		{"quz", "^$", true, "(quz ~ '^$')"},
	}

	for _, test := range tests {
		query, err := Match(test.identifier, test.regexp)
		if assert.NoError(t, err) {
			assert.Equal(t, query.Eval(context1), test.result)
			assert.Equal(t, query.(fmt.Stringer).String(), test.string)
		}
	}
}
