package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeysSlice(t *testing.T) {
	if keys, err := KeysSlice(map[string]bool{"foo": true, "bar": false}); assert.NoError(t, err) {
		expected := []string{"foo", "bar"}
		if assert.IsType(t, keys.Interface(), expected) {
			assert.ElementsMatch(
				t,
				keys.Interface().([]string),
				expected)
		}
	}

	if keys, err := KeysSlice(map[int]bool{}); assert.NoError(t, err) {
		expected := []int{}
		if assert.IsType(t, keys.Interface(), expected) {
			assert.ElementsMatch(
				t,
				keys.Interface().([]int),
				expected)
		}
	}

	if _, err := KeysSlice("foo"); assert.Error(t, err) {
	}
}
