package maps

import (
	"reflect"

	"github.com/pkg/errors"
)

func KeysSlice(m interface{}) (reflect.Value, error) {
	// Make sure input is a map.
	if reflect.TypeOf(m).Kind() != reflect.Map {
		return reflect.Value{}, errors.New("not a map")
	}

	// Build a slice to hold the keys.
	keys := reflect.ValueOf(m).MapKeys()
	result := reflect.MakeSlice(reflect.SliceOf(
		reflect.TypeOf(m).Key()), len(keys), len(keys))
	for i, key := range keys {
		result.Index(i).Set(key)
	}

	// Return the slice.
	return result, nil
}
