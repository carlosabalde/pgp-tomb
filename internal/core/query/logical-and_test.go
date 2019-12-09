package query

import "testing"

func TestLogicalAnd(t *testing.T) {
	cases := []struct {
		query  Query
		result bool
	}{
		{And(False, False), false},
		{And(False, True), false},
		{And(True, False), false},
		{And(True, True), true},
	}
	for _, test := range cases {
		if test.query.Eval(nil) != test.result {
			t.Error("unexpected result")
		}
	}
}
