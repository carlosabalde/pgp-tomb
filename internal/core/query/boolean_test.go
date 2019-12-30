package query

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoolean(t *testing.T) {
	tests := []struct {
		query  Query
		result bool
		string string
	}{
		{True, true, "true"},
		{False, false, "false"},
	}

	for _, test := range tests {
		assert.Equal(t, test.query.Eval(nil), test.result)
		assert.Equal(t, test.query.(fmt.Stringer).String(), test.string)
	}
}
