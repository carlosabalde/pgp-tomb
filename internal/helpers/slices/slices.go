package slices

import (
	"errors"
	"reflect"
)

func Union(arr1, arr2 interface{}) (reflect.Value, error) {
	// Create a map to hold contents of both slices.
	items := make(map[interface{}]bool)

	// Put values of both slices into the map, checking that types are
	// consistent.
	var kind reflect.Kind
	unknownKind := true
	for _, arr := range [2]interface{}{arr1, arr2} {
		tmp, err := distinct(arr)
		if err != nil {
			return reflect.Value{}, err
		}

		if unknownKind {
			if tmp.Len() > 0 {
				unknownKind = false
				kind = tmp.Index(0).Kind()
			}
		} else {
			if tmp.Len() > 0 && tmp.Index(0).Kind() != kind {
				return reflect.Value{}, errors.New("incompatible slice types")
			}
		}

		n := tmp.Len()
		for i := 0; i < n; i++ {
			items[tmp.Index(i).Interface()] = true
		}
	}

	// Create the result slice.
	return done(arr1, items), nil
}

func Difference(arr1, arr2 interface{}) (reflect.Value, error) {
	// Create a map to hold contents of the first slice.
	items := make(map[interface{}]bool)

	// Put values of the first slice into the map.
	tmpArr1, errArr1 := distinct(arr1)
	if errArr1 != nil {
		return reflect.Value{}, errArr1
	}
	nArr1 := tmpArr1.Len()
	for i := 0; i < nArr1; i++ {
		items[tmpArr1.Index(i).Interface()] = true
	}

	// Remove values of the second slice from the map, checking that types are
	// consistent.
	tmpArr2, errArr2 := distinct(arr2)
	if errArr2 != nil {
		return reflect.Value{}, errArr2
	}
	if tmpArr1.Len() > 0 && tmpArr2.Len() > 0 && tmpArr1.Index(0).Kind() != tmpArr2.Index(0).Kind() {
		return reflect.Value{}, errors.New("incompatible slice types")
	}
	nArr2 := tmpArr2.Len()
	for i := 0; i < nArr2; i++ {
		delete(items, tmpArr2.Index(i).Interface())
	}

	// Create the result slice.
	return done(arr1, items), nil
}

func distinct(arr interface{}) (reflect.Value, error) {
	// Create a slice from the input interface.
	slice := reflect.ValueOf(arr)
	if slice.Kind() != reflect.Slice {
		return reflect.Value{}, errors.New("not a slice")
	}

	// Create a map to hold contents of the slice.
	items := make(map[interface{}]bool)

	// Put values of the slice into a map.
	n := slice.Len()
	for i := 0; i < n; i++ {
		items[slice.Index(i).Interface()] = true
	}

	// Create the result slice.
	return done(arr, items), nil
}

func done(arr interface{}, items map[interface{}]bool) reflect.Value {
	result := reflect.MakeSlice(reflect.TypeOf(arr), len(items), len(items))

	i := 0
	for item := range items {
		result.Index(i).Set(reflect.ValueOf(item))
		i++
	}

	return result
}
