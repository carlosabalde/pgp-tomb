package maps

import (
	"errors"
	"reflect"
)

func StringKeysSlice(m interface{}) (reflect.Value, error) {
	rv := reflect.ValueOf(m)
 	if rv.Kind() != reflect.Map {
	    return reflect.Value{}, errors.New("not a map")
  	}

  	keys := rv.MapKeys()
  	result := reflect.MakeSlice(reflect.SliceOf(rv.Type().Key()), len(keys), len(keys))
	for i, key := range keys {
		result.Index(i).Set(key)
	}

	return result, nil
}
