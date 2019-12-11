package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeysSlice(t *testing.T) {
	if result, err := KeysSlice(map[string]bool{"foo": true, "bar": false}); assert.NoError(t, err) {
		expected := []string{"foo", "bar"}
		if assert.IsType(t, result.Interface(), expected) {
			assert.ElementsMatch(
				t,
				result.Interface().([]string),
				expected)
		}
	}

	if result, err := KeysSlice(map[int]bool{}); assert.NoError(t, err) {
		expected := []int{}
		if assert.IsType(t, result.Interface(), expected) {
			assert.ElementsMatch(
				t,
				result.Interface().([]int),
				expected)
		}
	}

	if _, err := KeysSlice("foo"); assert.Error(t, err) {
	}
}
