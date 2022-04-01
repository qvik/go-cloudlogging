package cloudlogging

import (
	"fmt"
	"testing"
)

const (
	sampleLabel = "this_is_a_sample_label"
	sampleValue = "This is a sample value for a logging label"
)

func BenchmarkPrimeLabelConversion(b *testing.B) {
	myMap := make(map[string]string)
	var key interface{} = sampleLabel
	var value interface{} = sampleValue

	for i := 0; i < b.N; i++ {
		myMap[fmt.Sprint(key)] = fmt.Sprint(value)
	}
}

func BenchmarkPrimeLabelKeyCast(b *testing.B) {
	myMap := make(map[string]string)
	var key interface{} = sampleLabel
	var value interface{} = sampleValue

	for i := 0; i < b.N; i++ {
		if stringKey, ok := key.(string); ok {
			myMap[stringKey] = fmt.Sprint(value)
		} else {
			myMap[fmt.Sprint(key)] = fmt.Sprint(value)
		}
	}
}

func BenchmarkPrimeLabelKeyValueCast(b *testing.B) {
	myMap := make(map[string]string)
	var key interface{} = sampleLabel
	var value interface{} = sampleValue

	for i := 0; i < b.N; i++ {
		if stringKey, ok := key.(string); ok {
			if stringValue, ok := value.(string); ok {
				myMap[stringKey] = stringValue
			} else {
				myMap[stringKey] = fmt.Sprint(value)
			}
		} else {
			myMap[fmt.Sprint(key)] = fmt.Sprint(value)
		}
	}
}

func compareListValuesToMap(list []interface{},
	theMap map[interface{}]interface{}) bool {

	for i := 0; i < len(list); i += 2 {
		k := list[i]
		v := list[i+1]

		if theMap[k] != v {
			return false
		}
	}

	return true
}

func TestWithCommonKeysAndValues(t *testing.T) {
	v := []interface{}{"key1", "value1", "key2", "value2"}
	log, err := NewLogger(WithCommonKeysAndValues(v...))
	if err != nil {
		t.Fatalf("failed to create logger")
	}

	if (len(v) / 2) != len(log.commonKeysAndValues) {
		t.Errorf("mismatching param array lengths: %v vs %v: %+v",
			len(v), len(log.commonKeysAndValues),
			log.commonKeysAndValues)
	}

	if !compareListValuesToMap(v, log.commonKeysAndValues) {
		t.Errorf("list values dont match those in the map")
	}
}

func TestWithAdditionalKeysAndValues(t *testing.T) {
	v1 := []interface{}{"key1", "value1", "key2", false}
	v2 := []interface{}{"key3", 123}

	baseLog, err := NewLogger(WithCommonKeysAndValues(v1...))
	if err != nil {
		t.Fatalf("failed to create logger")
	}

	log := baseLog.WithAdditionalKeysAndValues(v2...)

	if log == baseLog {
		t.Error("indistinctive logger instances")
	}

	if log.logLevel != baseLog.logLevel {
		t.Error("distinctive log levels")
	}

	// Check that base logger has not been affected
	if (len(v1) / 2) != len(baseLog.commonKeysAndValues) {
		t.Errorf("mismatching param array lengths: %v vs %v: %+v",
			len(v1), len(baseLog.commonKeysAndValues),
			baseLog.commonKeysAndValues)
	}

	if !compareListValuesToMap(v1, baseLog.commonKeysAndValues) {
		t.Errorf("list values dont match those in the map")
	}

	v := append(v1, v2...)
	if (len(v) / 2) != len(log.commonKeysAndValues) {
		t.Errorf("mismatching param array lengths: %v vs %v: %+v",
			len(v), len(log.commonKeysAndValues),
			log.commonKeysAndValues)
	}

	if !compareListValuesToMap(v, log.commonKeysAndValues) {
		t.Errorf("list values dont match those in the map")
	}
}

// GODOC EXAMPLES

func ExampleLogger_Debug() {
	log, _ := NewLogger() // Allocates a null logger
	log.Debug("Debug log message", "label1", 1, "label2", 2)
}

func ExampleLogger_Debugf() {
	log, _ := NewLogger() // Allocates a null logger
	myMsg := "message"
	log.Debugf("Debug with msg: %v", myMsg)
}

func ExampleLogger_WithAdditionalKeysAndValues() {
	// Allocates a null logger with no actual backends
	baseLog, _ := NewLogger(WithCommonKeysAndValues("key1", "value1"))

	// Create a "sub" logger that inherits the default keys and values from
	// its defined baselogger (baseLog)
	subLog := baseLog.WithAdditionalKeysAndValues("key2", "value2")
	subLog.Debug("Sublog debug message", "label", "value")
}
