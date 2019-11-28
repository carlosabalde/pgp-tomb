package slices

import (
	"reflect"

	"github.com/pkg/errors"
)

func Union(arr1, arr2 interface{}) (reflect.Value, error) {
	// Make sure inputs are slices.
	if reflect.TypeOf(arr1).Kind() != reflect.Slice ||
		reflect.TypeOf(arr2).Kind() != reflect.Slice {
		return reflect.Value{}, errors.New("not a slice")
	}

	// Make sure both slices have compatible types.
	if reflect.TypeOf(arr1) != reflect.TypeOf(arr2) {
		return reflect.Value{}, errors.New("incompatible types")
	}

	// Put values of both slices as keys into a map to avoid repetitions.
	items := make(map[interface{}]bool)
	for _, arr := range [2]interface{}{arr1, arr2} {
		slice := reflect.ValueOf(arr)
		for i, n := 0, slice.Len(); i < n; i++ {
			items[slice.Index(i).Interface()] = true
		}
	}

	// Turn map keys into a slice and return it.
	return sliceFromMap(items, arr1), nil
}

func Difference(arr1, arr2 interface{}) (reflect.Value, error) {
	// Make sure inputs are slices.
	if reflect.TypeOf(arr1).Kind() != reflect.Slice ||
		reflect.TypeOf(arr2).Kind() != reflect.Slice {
		return reflect.Value{}, errors.New("not a slice")
	}

	// Make sure both slices have compatible types.
	if reflect.TypeOf(arr1) != reflect.TypeOf(arr2) {
		return reflect.Value{}, errors.New("incompatible types")
	}

	// Put values of the first slice as keys into a map and
	// then remove values of the second slice.
	items := make(map[interface{}]bool)
	slice1 := reflect.ValueOf(arr1)
	for i, n := 0, slice1.Len(); i < n; i++ {
		items[slice1.Index(i).Interface()] = true
	}
	slice2 := reflect.ValueOf(arr2)
	for i, n := 0, slice2.Len(); i < n; i++ {
		delete(items, slice2.Index(i).Interface())
	}

	// Turn map keys into a slice and return it.
	return sliceFromMap(items, arr1), nil
}

func sliceFromMap(items map[interface{}]bool, arr interface{}) reflect.Value {
	result := reflect.MakeSlice(reflect.TypeOf(arr), len(items), len(items))
	i := 0
	for item := range items {
		result.Index(i).Set(reflect.ValueOf(item))
		i++
	}
	return result
}
