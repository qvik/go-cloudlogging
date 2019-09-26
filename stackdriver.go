package cloudlogging

import (
	"context"
	"fmt"
	"os"

	stackdriver "cloud.google.com/go/logging"
	"google.golang.org/api/option"
)

var (
	levelToStackdriverSeverityMap map[Level]stackdriver.Severity
)

func createStackdriverLogger(opts options) (*stackdriver.Client,
	*stackdriver.Logger, error) {

	o := []option.ClientOption{}

	if opts.credentialsFilePath != "" {
		o = append(o, option.WithCredentialsFile(opts.credentialsFilePath))
	}

	client, err := stackdriver.NewClient(context.Background(),
		opts.gcpProjectID, o...)
	if err != nil {
		return nil, nil, err
	}

	// Install an error handler
	client.OnError = func(err error) {
		fmt.Fprintf(os.Stderr, "Stackdriver error: %v", err)
	}

	loggeropts := []stackdriver.LoggerOption{}
	if opts.stackDriverMonitoredResource != nil {
		loggeropts = append(loggeropts,
			stackdriver.CommonResource(opts.stackDriverMonitoredResource))
	}

	logger := client.Logger(opts.stackdriverLogID, loggeropts...)

	if err := logger.LogSync(context.Background(), stackdriver.Entry{
		Payload:  "Stackdriver logger created",
		Severity: stackdriver.Info,
	}); err != nil {
		fmt.Fprintf(os.Stderr,
			"failed to initialize Stackdriver logging: %v", err)
	}

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
