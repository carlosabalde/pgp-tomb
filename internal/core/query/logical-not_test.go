package query

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogicalNot(t *testing.T) {
	tests := []struct {
		query  Query
		result bool
		string string
	}{
		{Not(False), true, "!false"},
		{Not(True), false, "!true"},
	}

	for _, test := range tests {
		assert.Equal(t, test.query.Eval(nil), test.result)
		assert.Equal(t, test.query.(fmt.Stringer).String(), test.string)
	}
}
