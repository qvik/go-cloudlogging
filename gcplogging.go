package cloudlogging

import (
	"fmt"
	stdlog "log"
	"os"

	"github.com/qvik/go-cloudlogging/internal"

	"google.golang.org/genproto/googleapis/api/monitoredres"
)

// NewComputeEngineLogger returns a Logger suitable for use in Google Compute Engine
// instance as well as Google Kubernetes Engine. The returned logger only logs via Stackdriver.
func NewComputeEngineLogger(projectID, logID string) (*Logger, error) {
	// See about using https://godoc.org/cloud.google.com/go/logging#CommonResource
	// with values from:
	//https://cloud.google.com/logging/docs/api/v2/resource-list#resource-types

	opts := []LogOption{}

	// On GCE we can omit supplying a MonitoredResource - it will be
	// autodetected:
	// https://godoc.org/cloud.google.com/go/logging#CommonResource
	opts = append(opts, WithStackdriver(projectID, "", logID, nil))

	return NewLogger(opts...)
}

// MustNewComputeEngineLogger returns a Logger suitable for use in Google Compute Engine
// instance as well as Google Kubernetes Engine. The returned logger only logs via Stackdriver.
// Panics on errors.
func MustNewComputeEngineLogger(projectID, logID string) *Logger {
	log, err := NewComputeEngineLogger(projectID, logID)
	if err != nil {
		stdlog.Panicf("failed to create logger: %v", err)
	}

	return log
}

// NewCloudFunctionLogger returns a Logger suitable for use in Google
// Cloud Functions. It will emit the logs using the Stackdriver API.
// The first value of args is the logID. If omitted or empty string is given,
// the default value of "cloudfunctions.googleapis.com/cloud-functions" is used.
func NewCloudFunctionLogger(args ...string) (*Logger, error) {
	// See about using https://godoc.org/cloud.google.com/go/logging#CommonResource
	// with values from:
	//https://cloud.google.com/logging/docs/api/v2/resource-list#resource-types

	logID := "cloudfunctions.googleapis.com/cloud-functions"
	if arg0, ok := internal.GetArg(0, args...); ok && arg0 != "" {
		logID = arg0
	}

	projectID := os.Getenv("GCP_PROJECT")
	if projectID == "" {
		return nil, fmt.Errorf("env var GCP_PROJECT missing")
	}

	functionName := os.Getenv("FUNCTION_NAME")
	if functionName == "" {
		return nil, fmt.Errorf("env var FUNCTION_NAME missing")
	}

	functionRegion := os.Getenv("FUNCTION_REGION")
	if functionName == "" {
		return nil, fmt.Errorf("env var FUNCTION_REGION missing")
	}

	opts := []LogOption{}

	// Create a monitored resource descriptor that will target Cloud Functions
	monitoredRes := &monitoredres.MonitoredResource{
		Type: "cloud_function",
		Labels: map[string]string{
			"project_id":    projectID,
			"function_name": functionName,
			"region":        functionRegion,
		},
	}

	opts = append(opts,
		WithStackdriver(projectID, "", logID, monitoredRes))

	return NewLogger(opts...)
}

// MustNewCloudFunctionLogger returns a Logger suitable for use in Google
// Cloud Functions. It will emit the logs using the Stackdriver API.
// The first value of args is the logID. If omitted or empty string is given,
// the default value of "cloudfunctions.googleapis.com/cloud-functions" is used.
// Panics on errors.
func MustNewCloudFunctionLogger(args ...string) *Logger {
	log, err := NewCloudFunctionLogger(args...)
	if err != nil {
		stdlog.Panicf("failed to create logger: %v", err)
	}

	return log
}

// NewAppEngineLogger returns a Logger suitable for use in AppEngine.
// On local dev server it uses the local Zap logger and in the cloud it
// uses the Stackdriver logger.
// The first value of args is the logID. If omitted or empty string is given,
// the default value of "appengine.googleapis.com/request_log" is used.
func NewAppEngineLogger(args ...string) (*Logger, error) {
	opts := []LogOption{}

	logID := "appengine.googleapis.com/request_log"
	if arg0, ok := internal.GetArg(0, args...); ok && arg0 != "" {
		logID = arg0
	}

	if os.Getenv("NODE_ENV") == "production" {
		projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
		if projectID == "" {
			return nil, fmt.Errorf("env var GOOGLE_CLOUD_PROJECT missing")
		}

		serviceID := os.Getenv("GAE_SERVICE")
		if serviceID == "" {
			return nil, fmt.Errorf("env var GAE_SERVICE missing")
		}

		versionID := os.Getenv("GAE_VERSION")
		if versionID == "" {
			return nil, fmt.Errorf("env var GAE_VERSION missing")
		}

		// Create a monitored resource descriptor that will target GAE
		monitoredRes := &monitoredres.MonitoredResource{
			Type: "gae_app",
			Labels: map[string]string{
				"project_id": projectID,
				"module_id":  serviceID,
				"version_id": versionID,
			},
		}

		opts = append(opts, WithStackdriver(projectID,
			"", logID, monitoredRes))
	} else {
		opts = append(opts, WithZap())
	}

	return NewLogger(opts...)
}

// MustNewAppEngineLogger returns a Logger suitable for use in AppEngine.
// On local dev server it uses the local stdout -logger and in the cloud it
// uses the Stackdriver logger.
// The first value of args is the logID. If omitted or empty string is given,
// the default value of "appengine.googleapis.com/request_log" is used.
// Panics on errors.
func MustNewAppEngineLogger(args ...string) *Logger {
	log, err := NewAppEngineLogger(args...)
	if err != nil {
		stdlog.Panicf("failed to create logger: %v", err)
	}

	return log
}
