package slices

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnionBasics(t *testing.T) {
	result, err := Union([]int{1, 2, 3}, []int{2, 5})
	require.NoError(t, err)
	expected := []int{1, 2, 3, 5}
	require.IsType(t, result.Interface(), expected)
	require.ElementsMatch(
		t,
		result.Interface().([]int),
		expected)
}

func TestUnionIncompatibleTypes(t *testing.T) {
	_, err := Union([]int{1}, []string{"foo"})
	require.Error(t, err)
}

func TestUnionNotASlice(t *testing.T) {
	_, err := Union("foo", "bar")
	require.Error(t, err)
}

func TestDifferenceBasics(t *testing.T) {
	{
		result, err := Difference([]int{1, 2, 3}, []int{2, 5})
		require.NoError(t, err)
		expected := []int{1, 3}
		require.IsType(t, result.Interface(), expected)
		require.ElementsMatch(
			t,
			result.Interface().([]int),
			expected)
	}

	{
		result, err := Difference([]int{2, 5}, []int{1, 2, 3})
		require.NoError(t, err)
		expected := []int{5}
		require.IsType(t, result.Interface(), expected)
		require.ElementsMatch(
			t,
			result.Interface().([]int),
			expected)
	}
}

func TestDifferenceIncompatibleTypes(t *testing.T) {
	_, err := Difference([]int{1}, []string{"foo"})
	require.Error(t, err)
}

func TestDifferenceNotASlice(t *testing.T) {
	_, err := Difference("foo", "bar")
	require.Error(t, err)
}
