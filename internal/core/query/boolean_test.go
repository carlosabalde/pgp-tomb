package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoolean(t *testing.T) {
	tests := []struct {
		query  Query
		result bool
	}{
		{True, true},
		{False, false},
	}

	for _, test := range tests {
		assert.Equal(t, test.query.Eval(nil), test.result)
	}
}
