package maps

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeysSliceBasics(t *testing.T) {
	result, err := KeysSlice(map[string]bool{"foo": true, "bar": false})
	require.NoError(t, err)
	expected := []string{"foo", "bar"}
	require.IsType(t, result.Interface(), expected)
	require.ElementsMatch(
		t,
		result.Interface().([]string),
		expected)
}

func TestKeysSliceEmptyMap(t *testing.T) {
	result, err := KeysSlice(map[int]bool{})
	require.NoError(t, err)
	expected := []int{}
	require.IsType(t, result.Interface(), expected)
	require.ElementsMatch(
		t,
		result.Interface().([]int),
		expected)
}

func TestKeysSliceNotAMap(t *testing.T) {
	_, err := KeysSlice("foo")
	require.Error(t, err)
}
