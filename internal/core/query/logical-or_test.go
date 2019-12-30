package query

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogicalOr(t *testing.T) {
	tests := []struct {
		query  Query
		result bool
		string string
	}{
		{Or(False, False), false, "(false || false)"},
		{Or(False, True), true, "(false || true)"},
		{Or(True, False), true, "(true || false)"},
		{Or(True, True), true, "(true || true)"},
	}

	for _, test := range tests {
		assert.Equal(t, test.query.Eval(nil), test.result)
		assert.Equal(t, test.query.(fmt.Stringer).String(), test.string)
	}
}
