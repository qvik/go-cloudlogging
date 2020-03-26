package cloudlogging

import "testing"

func TestSetDefaultKeysAndValues(t *testing.T) {
	v := []interface{}{"key1", "value1", "key2", "value2"}
	log, err := NewLogger()
	if err != nil {
		t.Fatalf("failed to create logger")
	}

	log.SetDefaultKeysAndValues(v...)

	for i, x := range log.defaultKeysAndValues {
		if x != v[i] {
			t.Errorf("unexpected key/value: %v", x)
		}
	}
}

func TestWithAdditionalKeysAndValues(t *testing.T) {
	v1 := []interface{}{"key1", "value1", "key2", false}
	v2 := []interface{}{"key3", 123}

	baseLog, err := NewLogger()
	if err != nil {
		t.Fatalf("failed to create logger")
	}
	baseLog.SetDefaultKeysAndValues(v1...)

	log := baseLog.WithAdditionalKeysAndValues(v2...)

	if log.baseLog != baseLog {
		t.Errorf("invalid baseLog")
	}

	// Check that base logger has not been affected
	if len(v1) != len(baseLog.defaultKeysAndValues) {
		t.Errorf("mismatching param array lengths: %v vs %v",
			len(v1), len(baseLog.defaultKeysAndValues))
	}
	for i, x := range baseLog.defaultKeysAndValues {
		if x != v1[i] {
			t.Errorf("unexpected key/value: %v", x)
		}
	}

	v := append(v1, v2...)
	if len(v) != len(log.defaultKeysAndValues) {
		t.Errorf("mismatching param array lengths: %v vs %v",
			len(v), len(log.defaultKeysAndValues))
	}

	for i, x := range log.defaultKeysAndValues {
		if x != v[i] {
			t.Errorf("unexpected key/value: %v", x)
		}
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
	baseLog, _ := NewLogger() // Allocates a null logger
	baseLog.SetDefaultKeysAndValues("key1", "value1")

	// Create a "sub" logger that inherits the default keys and values from
	// its defined baselogger (baseLog)
	subLog := baseLog.WithAdditionalKeysAndValues("key2", "value2")
	subLog.Debug("Sublog debug message", "label", "value")
}
