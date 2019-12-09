package query

import "testing"

func TestParse(t *testing.T) {
	context1 := Map{
		"uri": "foo/bar/baz.txt",
		"tags.foo": "42",
		"tags.bar": "3.14",
		"tags.baz": "",
	}

	for _, s := range []string{
		`uri == "foo/bar/baz.txt"`,
		`uri != "xxx"`,
		`uri ~ "^foo/bar/"`,
		`uri !~ "^xxx"`,
		`uri == "foo/bar/baz.txt" && tags.foo == '42' && tags.bar != "42"`,
		`tags.foo ~ '^xxx' || tags.bar ~ "14$"`,
	} {
		query, err := Parse(s)
		t.Logf("%s", query)
		if err != nil {
			t.Fatal(err)
		}
		if !query.Eval(context1) {
			t.Error("unexpected result")
		}
	}
}
