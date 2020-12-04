package internal

import (
	stdlog "log"
)

// ApplyKeysAndValues writes a list of keys and values in the argument array
// into the supplied map. The format for the the keysAndValues parameter is
// key1, value1, key2, value2, .. and its length must be a multiple of two
// or the function panics.
func MustApplyKeysAndValues(keysAndValues []interface{},
	toMap map[interface{}]interface{}) {

	if len(keysAndValues) == 0 {
		return
	}

	if len(keysAndValues)%2 != 0 {
		stdlog.Panicf("must pass even number of keysAndValues")
	}

	count := 0
	for count < len(keysAndValues) {
		key := keysAndValues[count]
		value := keysAndValues[count+1]

		toMap[key] = value

		count += 2
	}
}

// MapToKeysAndValuesList creates a list of keys and values out of a
// map (order matching the enumaration order of the map's keys). The
// format is key1, value1, key2, value2, ..
func MapToKeysAndValuesList(theMap map[interface{}]interface{}) []interface{} {
	list := make([]interface{}, len(theMap)*2)
	count := 0
	for k, v := range theMap {
		list[count] = k
		list[count+1] = v
		count += 2
	}

	return list
}

// GetArg returns the n:th value in args, if defined, otherwise returns false
// in the boolean return value. n must be a positive value.
//
// Use it as such:
//  if value, ok := getArg(1, args); ok {
//    // use 'value'
//  }
func GetArg(n int, args ...string) (string, bool) {
	if n < 0 {
		return "", false
	}

	if len(args) > n {
		return args[n], true
	}

	return "", false
}
