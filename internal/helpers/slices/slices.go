package slices

import (
	"reflect"
)

func Union(arr1, arr2 interface{}) reflect.Value {
	// Create a map to hold contents of both slices.
	items := make(map[interface{}]bool)

	// Put values of both slices into the map, checking that types are
	// consistent.
	var kind reflect.Kind
	unknownKind := true
	for _, arr := range [2]interface{}{arr1, arr2} {
		tmp, ok := distinct(arr)
		if !ok {
			return reflect.Value{}
		}

		if unknownKind {
			if tmp.Len() > 0 {
				unknownKind = false
				kind = tmp.Index(0).Kind()
			}
		} else {
			if tmp.Len() > 0 && tmp.Index(0).Kind() != kind {
				return reflect.Value{}
			}
		}

		n := tmp.Len()
		for i := 0; i < n; i++ {
			items[tmp.Index(i).Interface()] = true
		}
	}

	// Create the result slice.
	return done(arr1, items)
}

func Difference(arr1, arr2 interface{}) reflect.Value {
	// Create a map to hold contents of the first slice.
	items := make(map[interface{}]bool)

	// Put values of the first slice into the map.
	tmpArr1, okArr1 := distinct(arr1)
	if !okArr1 {
		return reflect.Value{}
	}
	nArr1 := tmpArr1.Len()
	for i := 0; i < nArr1; i++ {
		items[tmpArr1.Index(i).Interface()] = true
	}

	// Remove values of the second slice from the map, checking that types are
	// consistent.
	tmpArr2, okArr2 := distinct(arr2)
	if !okArr2 {
		return reflect.Value{}
	}
	if tmpArr1.Len() > 0 && tmpArr2.Len() > 0 && tmpArr1.Index(0).Kind() != tmpArr2.Index(0).Kind() {
		return reflect.Value{}
	}
	nArr2 := tmpArr2.Len()
	for i := 0; i < nArr2; i++ {
		delete(items, tmpArr2.Index(i).Interface())
	}

	// Create the result slice.
	return done(arr1, items)
}

func distinct(arr interface{}) (reflect.Value, bool) {
	// Create a slice from the input interface.
	slice := reflect.ValueOf(arr)
	if slice.Kind() != reflect.Slice {
		return reflect.Value{}, false
	}

	// Create a map to hold contents of the slice.
	items := make(map[interface{}]bool)

	// Put values of the slice into a map.
	n := slice.Len()
	for i := 0; i < n; i++ {
		items[slice.Index(i).Interface()] = true
	}

	// Create the result slice.
	return done(arr, items), true
}

func done(arr interface{}, items map[interface{}]bool) reflect.Value {
	result := reflect.MakeSlice(reflect.TypeOf(arr), len(items), len(items))

	i := 0
	for item := range items {
		iValue := reflect.ValueOf(item)
		oValue := result.Index(i)
		oValue.Set(iValue)
		i++
	}

	return result
}
