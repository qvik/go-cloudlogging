package cloudlogging

import (
	"context"
	"fmt"
	stdlog "log"

	"cloud.google.com/go/logging"
	stackdriver "cloud.google.com/go/logging"
	googleoauth2 "golang.org/x/oauth2/google"
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
	} else {
		//TODO is this a good place to put this?
		tokenSource, err := googleoauth2.DefaultTokenSource(context.Background(), logging.WriteScope)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create default token source: %v", err)
		}

		o = append(o, option.WithTokenSource(tokenSource))
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
