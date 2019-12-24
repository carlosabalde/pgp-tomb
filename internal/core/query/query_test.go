package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	context1 := Map{
		"uri":      "foo/bar/baz.txt",
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
		if assert.NoError(t, err) {
			assert.True(t, query.Eval(context1))
		}
	}
}
