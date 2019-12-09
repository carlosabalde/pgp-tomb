package query

import "testing"

func TestBoolean(t *testing.T) {
	tests := []struct {
		query  Query
		result bool
	}{
		{True, true},
		{False, false},
	}

	for _, test := range tests {
		t.Logf("%s", test.query)
		if test.query.Eval(nil) != test.result {
			t.Error("unexpected result")
		}
	}
}
