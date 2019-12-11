package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogicalAnd(t *testing.T) {
	tests := []struct {
		query  Query
		result bool
	}{
		{And(False, False), false},
		{And(False, True), false},
		{And(True, False), false},
		{And(True, True), true},
	}

	for _, test := range tests {
		assert.Equal(t, test.query.Eval(nil), test.result)
	}
}
