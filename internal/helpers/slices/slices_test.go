package slices

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnion(t *testing.T) {
	if result, err := Union([]int{1, 2, 3}, []int{2, 5}); assert.NoError(t, err) {
		expected := []int{1, 2, 3, 5}
		if assert.IsType(t, result.Interface(), expected) {
			assert.ElementsMatch(
				t,
				result.Interface().([]int),
				expected)
		}
	}

	if _, err := Union([]int{1}, []string{"foo"}); assert.Error(t, err) {
	}

	if _, err := Union("foo", "bar"); assert.Error(t, err) {
	}
}

func TestDifference(t *testing.T) {
	if result, err := Difference([]int{1, 2, 3}, []int{2, 5}); assert.NoError(t, err) {
		expected := []int{1, 3}
		if assert.IsType(t, result.Interface(), expected) {
			assert.ElementsMatch(
				t,
				result.Interface().([]int),
				expected)
		}
	}

	if result, err := Difference([]int{2, 5}, []int{1, 2, 3}); assert.NoError(t, err) {
		expected := []int{5}
		if assert.IsType(t, result.Interface(), expected) {
			assert.ElementsMatch(
				t,
				result.Interface().([]int),
				expected)
		}
	}

	if _, err := Difference([]int{1}, []string{"foo"}); assert.Error(t, err) {
	}

	if _, err := Difference("foo", "bar"); assert.Error(t, err) {
	}
}
