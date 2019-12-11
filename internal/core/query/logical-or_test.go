package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogicalOr(t *testing.T) {
	tests := []struct {
		query  Query
		result bool
	}{
		{Or(False, False), false},
		{Or(False, True), true},
		{Or(True, False), true},
		{Or(True, True), true},
	}

	for _, test := range tests {
		assert.Equal(t, test.query.Eval(nil), test.result)
	}
}
