package query

import "testing"

func TestLogicalNot(t *testing.T) {
	tests := []struct {
		query  Query
		result bool
	}{
		{Not(False), true},
		{Not(True), false},
	}

	for _, test := range tests {
		t.Logf("%s", test.query)
		if test.query.Eval(nil) != test.result {
			t.Error("unexpected result")
		}
	}
}
