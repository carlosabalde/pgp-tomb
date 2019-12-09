package query

import "testing"

func TestMatch(t *testing.T) {
	context1 := Map{
		"foo": "42",
		"bar": "3.14",
		"baz": "",
	}

	type testCase struct {
		query   Query
		context Context
		result  bool
	}

	newTestCase := func(identifier, regexp string, context Context, result bool) (test testCase) {
		var err error
		test.query, err = Match(identifier, regexp)
		if err != nil {
			t.Fatal(err)
		}
		test.context = context
		test.result = result
		return
	}

	tests := []testCase{
		newTestCase("foo", "^42$", context1, true),
		newTestCase("foo", "^3.14$", context1, false),
		newTestCase("foo", ".*", context1, true),
		newTestCase("baz", "^$", context1, true),
		newTestCase("baz", "^ $", context1, false),
		newTestCase("quz", "^$", context1, true),
	}

	for _, test := range tests {
		t.Logf("%s", test.query)
		if test.query.Eval(test.context) != test.result {
			t.Error("unexpected result")
		}
	}
}
