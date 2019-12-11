package query

import (
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
	}{
		{Equal("foo", "42"), context1, true},
		{Equal("foo", "3.14"), context1, false},
		{Equal("foo", ""), context1, false},
		{Equal("baz", ""), context1, true},
		{Equal("baz", " "), context1, false},
		{Equal("quz", ""), context1, true},
	}

	for _, test := range tests {
		assert.Equal(t, test.query.Eval(test.context), test.result)
	}
}
