package query

import "testing"

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
		t.Logf("%s", test.query)
		if test.query.Eval(nil) != test.result {
			t.Error("unexpected result")
		}
	}
}
