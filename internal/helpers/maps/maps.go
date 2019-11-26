package maps

import (
	"reflect"
)

func StringKeysSlice(i interface{}) []string {
	keys := reflect.ValueOf(i).MapKeys()
	result := make([]string, len(keys))
	for i := 0; i < len(keys); i++ {
		result[i] = keys[i].String()
	}
	return result
}
