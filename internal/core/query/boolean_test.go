package query

import "testing"

func TestBoolean(t *testing.T) {
	cases := []struct {
		query  Query
		result bool
	}{
		{True, true},
		{False, false},
	}
	for _, test := range cases {
		if test.query.Eval(nil) != test.result {
			t.Error("unexpected result")
		}
	}
}
