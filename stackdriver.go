package cloudlogging

import (
	"context"
	stdlog "log"

	stackdriver "cloud.google.com/go/logging"
	"google.golang.org/api/option"
)

var (
	levelToStackdriverSeverityMap map[Level]stackdriver.Severity
)

func createStackdriverLogger(opts options) (*stackdriver.Client,
	*stackdriver.Logger, error) {

	ctx := context.Background()
	o := []option.ClientOption{}

	if opts.credentialsFilePath != "" {
		o = append(o, option.WithCredentialsFile(opts.credentialsFilePath))
	}

	client, err := stackdriver.NewClient(ctx, opts.gcpProjectID, o...)
	if err != nil {
		return nil, nil, err
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

	if err := logger.LogSync(context.Background(), stackdriver.Entry{
		Payload:  "Stackdriver logger created",
		Severity: stackdriver.Info,
	}); err != nil {
		stdlog.Printf("failed to initialize Stackdriver logging: %v", err)
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
