package zap

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/qvik/go-cloudlogging"
)

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

func TestZapLogger(t *testing.T) {
	logOutput := captureStdout(func() {
		opts := []cloudlogging.LogOption{
			cloudlogging.WithZap(),
		}
		log, err := cloudlogging.NewLogger(opts...)
		if err != nil {
			t.Errorf("failed to create logger: %v", err)
		}

		log.SetLogLevel(cloudlogging.Debug)

		log.Debugf("Test A=%v,B=%v", 1, 2)

		// Note: we cannot call Close() or Flush() b/c we have captured stdout
		// and zap's Sync() will crash otherwise
	})

	if !strings.HasSuffix(logOutput, "Test A=1,B=2") {
		t.Errorf("Invalid log output: %v", logOutput)
	}
}
