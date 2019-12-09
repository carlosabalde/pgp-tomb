package query

import "testing"

func TestLogicalNot(t *testing.T) {
	cases := []struct {
		query  Query
		result bool
	}{
		{Not(False), true},
		{Not(True), false},
	}
	for _, test := range cases {
		if test.query.Eval(nil) != test.result {
			t.Error("unexpected result")
		}
	}
}
