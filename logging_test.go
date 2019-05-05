package cloudlogging

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

// To run the tests, run:  go test -v github.com/qvik/cloudlogging

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

func TestAppEngine(t *testing.T) {
	// Simulate app engine production env
	os.Setenv("NODE_ENV", "production")
	os.Setenv("GOOGLE_CLOUD_PROJECT", "test-does-not-exist")
	os.Setenv("GAE_SERVICE", "default")
	os.Setenv("GAE_VERSION", "test-does-not-exist")

	log, err := NewAppEngineLogger()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	log.SetLogLevel(Debug)

	log.Infof("Hello, world")

	if err := log.Flush(); err == nil {
		t.Errorf("flush must fail here!")
	}
}

// GODOC EXAMPLES

func ExampleDebug() {
	log, _ := NewLocalOnlyLogger()
	log.Debug("Debug log message", "label1", 1, "label2", 2)
}

func ExampleDebugf() {
	log, _ := NewLocalOnlyLogger()
	myMsg := "message"
	log.Debugf("Debug with msg: %v", myMsg)
}
