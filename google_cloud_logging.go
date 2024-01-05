package cloudlogging

import (
	"context"
	"fmt"
	stdlog "log"

	gcloudlog "cloud.google.com/go/logging"
	"google.golang.org/api/option"
)

var (
	levelToGoogleCloudLoggingSeverityMap map[Level]gcloudlog.Severity
)

// createGoogleCloudLoggingLogger creates a new Google Cloud Logging client and a logger
func createGoogleCloudLoggingLogger(opts options) (*gcloudlog.Client,
	*gcloudlog.Logger, error) {

	ctx := context.Background()
	o := []option.ClientOption{}

	if opts.credentialsFilePath != "" {
		o = append(o, option.WithCredentialsFile(opts.credentialsFilePath))
	}

	// See: https://godoc.org/cloud.google.com/go/logging#NewClient
	parent := fmt.Sprintf("projects/%v", opts.gcpProjectID)
	client, err := gcloudlog.NewClient(ctx, parent, o...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create google cloud logging client: %w", err)
	}

	// Install an error handler
	client.OnError = func(err error) {
		stdlog.Printf("google cloud logging error: %v", err)
	}

	loggeropts := []gcloudlog.LoggerOption{}
	if opts.googleCloudLoggingMonitoredResource != nil {
		loggeropts = append(loggeropts,
			gcloudlog.CommonResource(opts.googleCloudLoggingMonitoredResource))
	}

	logger := client.Logger(opts.googleCloudLoggingLogID, loggeropts...)

	// Emit a log entry for testing
	logger.Log(gcloudlog.Entry{
		Payload:  "google cloud logging logger created.",
		Severity: gcloudlog.Info,
	})

	return client, logger, nil
}

func init() {
	levelToGoogleCloudLoggingSeverityMap = map[Level]gcloudlog.Severity{
		Debug:   gcloudlog.Debug,
		Info:    gcloudlog.Info,
		Warning: gcloudlog.Warning,
		Error:   gcloudlog.Error,
		Fatal:   gcloudlog.Critical,
	}
}
