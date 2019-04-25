package cloudlogging

import (
	"encoding/json"
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

		log.SetLogLevel(Trace)

		log.Debugf("Test")
		if err := log.Flush(); err != nil {
			t.Errorf("failed to flush logger: %v", err)
		}
	})

	if !strings.HasSuffix(logOutput, "level=debug msg=Test") {
		t.Errorf("Invalid log output: %v", logOutput)
	}
}

func TestLocalFluentDLogger(t *testing.T) {
	logOutput := captureStdout(func() {
		opts := []LogOption{
			WithLocalFluentD(),
		}
		log, err := NewLogger(opts...)
		if err != nil {
			t.Errorf("failed to create logger: %v", err)
		}

		log.SetLogLevel(Trace)

		log.Debugf("Test")
		if err := log.Flush(); err != nil {
			t.Errorf("failed to flush logger: %v", err)
		}
	})

	type fluentdmsg struct {
		Message   string `json:"message"`
		Timestamp string `json:"timestamp"`
		Severity  string `json:"severity"`
	}

	msg := new(fluentdmsg)
	if err := json.Unmarshal([]byte(logOutput), msg); err != nil {
		t.Errorf("failed to read fluentd output: %v", err)
		return
	}

	if msg.Message == "" {
		t.Errorf("missing message in fluentd log entry")
	}

	if msg.Severity == "" {
		t.Errorf("missing severity in fluentd log entry")
	}

	if msg.Timestamp == "" {
		t.Errorf("missing timestamp in fluentd log entry")
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

	log.SetLogLevel(Trace)

	log.Infof("Hello, world")

	if err := log.Flush(); err == nil {
		t.Errorf("flush must fail here!")
	}
}
