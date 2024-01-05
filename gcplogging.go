package cloudlogging

import (
	"fmt"
	stdlog "log"
	"os"

	"github.com/qvik/go-cloudlogging/internal"

	"google.golang.org/genproto/googleapis/api/monitoredres"
)

// NewComputeEngineLogger returns a Logger suitable for use in Google Compute Engine
// instance as well as Google Kubernetes Engine. The returned logger
// only logs via Google Cloud Logging.
func NewComputeEngineLogger(projectID, logID string) (*Logger, error) {
	// See about using https://godoc.org/cloud.google.com/go/logging#CommonResource
	// with values from:
	//https://cloud.google.com/logging/docs/api/v2/resource-list#resource-types

	opts := []LogOption{}

	// On GCE we can omit supplying a MonitoredResource - it will be
	// autodetected:
	// https://godoc.org/cloud.google.com/go/logging#CommonResource
	opts = append(opts, WithGoogleCloudLogging(projectID, "", logID, nil))

	return NewLogger(opts...)
}

// MustNewComputeEngineLogger returns a Logger suitable for use in Google Compute Engine
// instance as well as Google Kubernetes Engine. The returned logger
// only logs via Google Cloud Logging.
// Panics on errors.
func MustNewComputeEngineLogger(projectID, logID string) *Logger {
	log, err := NewComputeEngineLogger(projectID, logID)
	if err != nil {
		stdlog.Panicf("failed to create logger: %v", err)
	}

	return log
}

// NewCloudFunctionLogger returns a Logger suitable for use in Google
// Cloud Functions. It will emit the logs using the Google Cloud Logging API.
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
		WithGoogleCloudLogging(projectID, "", logID, monitoredRes))

	return NewLogger(opts...)
}

// MustNewCloudFunctionLogger returns a Logger suitable for use in Google
// Cloud Functions. It will emit the logs using the Google Cloud Logging API.
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
// uses the Google Cloud Logging logger.
// The first value of args is the logID. If omitted or empty string is given,
// the default value of "appengine.googleapis.com/request_log" is used.
func NewAppEngineLogger(args ...string) (*Logger, error) {
	opts := []LogOption{}

	logID := "appengine.googleapis.com/request_log"
	if arg0, ok := internal.GetArg(0, args...); ok && arg0 != "" {
		logID = arg0
	}

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	serviceID := os.Getenv("GAE_SERVICE")
	versionID := os.Getenv("GAE_VERSION")

	if projectID != "" && serviceID != "" && versionID != "" {
		// Create a monitored resource descriptor that will target GAE
		monitoredRes := &monitoredres.MonitoredResource{
			Type: "gae_app",
			Labels: map[string]string{
				"project_id": projectID,
				"module_id":  serviceID,
				"version_id": versionID,
			},
		}

		opts = append(opts, WithGoogleCloudLogging(projectID,
			"", logID, monitoredRes))
	} else {
		// Not apparently running on Google App Engine, use local Zap logging
		opts = append(opts, WithZap())
	}

	return NewLogger(opts...)
}

// MustNewAppEngineLogger returns a Logger suitable for use in AppEngine.
// On local dev server it uses the local stdout -logger and in the cloud it
// uses the Google Cloud Logging logger.
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

// NewCloudRunLogger returns a Logger suitable for use in Cloud Run.
// On local dev server it uses the local Zap logger and in the cloud it
// uses the Google Cloud Logging logger.
// The first value of args is the logID. If omitted or empty string is given,
// the default value of "run.googleapis.com/request_log" is used.
func NewCloudRunLogger(projectID string, args ...string) (*Logger, error) {
	opts := []LogOption{}

	logID := "run.googleapis.com/request_log"
	if arg0, ok := internal.GetArg(0, args...); ok && arg0 != "" {
		logID = arg0
	}

	service := os.Getenv("K_SERVICE")
	revision := os.Getenv("K_REVISION")
	configuration := os.Getenv("K_CONFIGURATION")

	if service != "" && revision != "" && configuration != "" {
		// Create a monitored resource descriptor that will target GAE
		monitoredRes := &monitoredres.MonitoredResource{
			Type: "cloud_run_revision",
			Labels: map[string]string{
				"project_id":         projectID,
				"service_name":       service,
				"revision_name":      revision,
				"configuration_name": configuration,
			},
		}

		opts = append(opts, WithGoogleCloudLogging(projectID,
			"", logID, monitoredRes))
	} else {
		// Not apparently running on Google App Engine, use local Zap logging
		opts = append(opts, WithZap())
	}

	return NewLogger(opts...)
}
