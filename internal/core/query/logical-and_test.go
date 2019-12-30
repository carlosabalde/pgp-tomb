package query

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogicalAnd(t *testing.T) {
	tests := []struct {
		query  Query
		result bool
		string string
	}{
		{And(False, False), false, "(false && false)"},
		{And(False, True), false, "(false && true)"},
		{And(True, False), false, "(true && false)"},
		{And(True, True), true, "(true && true)"},
	}

	for _, test := range tests {
		assert.Equal(t, test.query.Eval(nil), test.result)
		assert.Equal(t, test.query.(fmt.Stringer).String(), test.string)
	}
}
