package cloudlogging

import (
	"context"
	"fmt"
	stdlog "log"

	stackdriver "cloud.google.com/go/logging"
	"google.golang.org/api/option"
)

var (
	levelToStackdriverSeverityMap map[Level]stackdriver.Severity
)

// createStackdriverLogger creates a new Stackdriver logging client and a logger
func createStackdriverLogger(opts options) (*stackdriver.Client,
	*stackdriver.Logger, error) {

	ctx := context.Background()
	o := []option.ClientOption{}

	if opts.credentialsFilePath != "" {
		//TODO use WriteScope here too
		o = append(o, option.WithCredentialsFile(opts.credentialsFilePath))
	}

	// See: https://godoc.org/cloud.google.com/go/logging#NewClient
	parent := fmt.Sprintf("projects/%v", opts.gcpProjectID)
	client, err := stackdriver.NewClient(ctx, parent, o...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create stackdriver client: %v", err)
	}

	// Install an error handler
	client.OnError = func(err error) {
		stdlog.Printf("Stackdriver error: %v", err)
	}

	loggeropts := []stackdriver.LoggerOption{}
	if opts.stackDriverMonitoredResource != nil {
		loggeropts = append(loggeropts,
			stackdriver.CommonResource(opts.stackDriverMonitoredResource))
	}

	logger := client.Logger(opts.stackdriverLogID, loggeropts...)

	// Emit a log entry for testing
	logger.Log(stackdriver.Entry{
		Payload:  "Stackdriver logger created",
		Severity: stackdriver.Info,
	})

	return client, logger, nil
}

func init() {
	levelToStackdriverSeverityMap = map[Level]stackdriver.Severity{
		Debug:   stackdriver.Debug,
		Info:    stackdriver.Info,
		Warning: stackdriver.Warning,
		Error:   stackdriver.Error,
		Fatal:   stackdriver.Critical,
	}
}
