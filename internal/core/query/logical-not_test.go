package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogicalNot(t *testing.T) {
	tests := []struct {
		query  Query
		result bool
	}{
		{Not(False), true},
		{Not(True), false},
	}

	for _, test := range tests {
		assert.Equal(t, test.query.Eval(nil), test.result)
	}
}
