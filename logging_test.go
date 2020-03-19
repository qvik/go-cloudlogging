package cloudlogging

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

// To run the tests, run:
// go test -v .

// captureStdout captures the stdout output of a function.
func captureStdout(f func()) string {
	// We will capture log output via pipe
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the function expected to generate output
	f()

	// Restore old stdout
	os.Stdout = old

	// Close & read the pipe
	w.Close()
	out, _ := ioutil.ReadAll(r)

	return strings.Trim(string(out), "\n ")
}

// captureStderr captures the stderr output of a function.
func captureStderr(f func()) string {
	// We will capture the output via pipe
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Run the function expected to generate output
	f()

	// Restore old stderr
	os.Stderr = old

	// Close & read the pipe
	w.Close()
	out, _ := ioutil.ReadAll(r)

	return strings.Trim(string(out), "\n ")
}

func TestLocalLogger(t *testing.T) {
	logOutput := captureStdout(func() {
		opts := []LogOption{
			WithLocal(),
		}
		log, err := NewLogger(opts...)
		if err != nil {
			t.Errorf("failed to create logger: %v", err)
		}

		log.SetLogLevel(Debug)

		log.Debugf("Test A=%v,B=%v", 1, 2)

		// Note: we cannot call Close() or Flush() b/c we have captured stdout
		// and zap's Sync() will crash otherwise
	})

	if !strings.HasSuffix(logOutput, "Test A=1,B=2") {
		t.Errorf("Invalid log output: %v", logOutput)
	}
}

func TestSetDefaultKeysAndValues(t *testing.T) {
	v := []interface{}{"key1", "value1", "key2", "value2"}
	log := MustNewLocalOnlyLogger()
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

	baseLog := MustNewLocalOnlyLogger()
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
	log, _ := NewLocalOnlyLogger()
	log.Debug("Debug log message", "label1", 1, "label2", 2)
}

func ExampleLogger_Debugf() {
	log, _ := NewLocalOnlyLogger()
	myMsg := "message"
	log.Debugf("Debug with msg: %v", myMsg)
}

func ExampleLogger_WithAdditionalKeysAndValues_output() {
	baseLog := MustNewLocalOnlyLogger()
	baseLog.SetDefaultKeysAndValues("key1", "value1")

	// Create a "sub" logger that inherits the default keys and values from
	// its defined baselogger (baseLog)
	subLog := baseLog.WithAdditionalKeysAndValues("key2", "value2")
	subLog.Debug("Sublog debug message", "label", "value")

	// Output: Sublog debug message, {key1: value1, key2: value2, label: value}
}
