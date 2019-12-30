package query

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEqual(t *testing.T) {
	context1 := Map{
		"foo": "42",
		"bar": "3.14",
		"baz": "",
	}

	tests := []struct {
		query   Query
		context Context
		result  bool
		string  string
	}{
		{Equal("foo", "42"), context1, true, "(foo == '42')"},
		{Equal("foo", "3.14"), context1, false, "(foo == '3.14')"},
		{Equal("foo", ""), context1, false, "(foo == '')"},
		{Equal("baz", ""), context1, true, "(baz == '')"},
		{Equal("baz", " "), context1, false, "(baz == ' ')"},
		{Equal("quz", ""), context1, true, "(quz == '')"},
	}

	for _, test := range tests {
		assert.Equal(t, test.query.Eval(test.context), test.result)
		assert.Equal(t, test.query.(fmt.Stringer).String(), test.string)
	}
}
