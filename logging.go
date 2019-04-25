// Package cloudlogging provides logging wrapper for Google Cloud Platform
// environments.
package cloudlogging

import (
	"context"
	"fmt"
	"os"

	stackdriver "cloud.google.com/go/logging"
	joonix "github.com/joonix/log"
	logrus "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"google.golang.org/genproto/googleapis/api/monitoredres"
)

// Log levels
const (
	Trace = iota
	Debug
	Info
	Warning
	Error
	Fatal
)

// Logger function
type logFunc func(string, ...interface{})

// Logger writes logs to the local logger as well as
// the Stackdriver cloud logs.
//
// Logger is thread safe to use; the Set* methods however, are not and
// you should either set them at the start of your program or synchronize
// your access to the Logger instance if setting on them on the fly. The
// reasoning here being that standard Logger calls should not suffer a
// performance hit from locking.
//
// Logger uses the logrus (https://github.com/sirupsen/logrus)
// library for local logging. It can be either
// configured to emit FluentD (with Kubernetes compatibility) or plain logs.
//
// Logger uses Stackdriver Go library for cloud logging.
//
// The API is compatible (and thus a drop-in replacement) for popular
// logging libraries such as:
// - Go's standard "log" interface
// - https://github.com/op/go-logging
// - https://github.com/sirupsen/logrus
//
type Logger struct {
	// Current log level
	logLevel int

	// Local formatting logger
	logrusLogger *logrus.Logger

	// Stackdriver client
	stackdriverClient *stackdriver.Client

	// Stackdriver logger
	stackdriverLogger *stackdriver.Logger
}

func createLogrusLogger(opts options) *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	if opts.localOutput != nil {
		logger.SetOutput(opts.localOutput)
	}

	if opts.useLocalFluentD {
		logger.SetFormatter(&joonix.FluentdFormatter{})
	} else {
		// If in local dev mode, use text output for logging
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:          true,
			DisableLevelTruncation: true,
		})
	}

	return logger
}

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

	loggeropts := []stackdriver.LoggerOption{}
	if opts.stackDriverMonitoredResource != nil {
		loggeropts = append(loggeropts,
			stackdriver.CommonResource(opts.stackDriverMonitoredResource))
	}

	logger := client.Logger(opts.stackdriverLogID, loggeropts...)

	return client, logger, nil
}

// NewLogger creates a new Logger instance using the given options.
func NewLogger(opt ...LogOption) (*Logger, error) {
	opts := options{}
	for _, o := range opt {
		o.apply(&opts)
	}

	if opts.useStackdriver && opts.gcpProjectID == "" {
		return nil, fmt.Errorf("Stackdriver requires a GCP project ID")
	}

	var stackdriverClient *stackdriver.Client
	var stackdriverLogger *stackdriver.Logger
	var logrusLogger *logrus.Logger

	if opts.useStackdriver {
		client, logger, err := createStackdriverLogger(opts)
		if err != nil {
			return nil, err
		}
		stackdriverClient = client
		stackdriverLogger = logger
	}

	if opts.useLocal {
		logrusLogger = createLogrusLogger(opts)
	}

	l := &Logger{
		logLevel:          Info,
		stackdriverClient: stackdriverClient,
		stackdriverLogger: stackdriverLogger,
		logrusLogger:      logrusLogger,
	}

	l.updateLogrusLoglevel()

	return l, nil
}

// NewComputeEngineLogger returns a Logger suitable for use in Compute Engine
// instance. Local logs are always written.
// If the stackdriver flag is set to true, logs are written to
// Stackdriver API; if the fluentd flag is set to true, local logs are
// formatted using FluentD formatter, making it possible to combine this
// with Stackdriver Logging Agent (https://cloud.google.com/logging/docs/agent/)
// that will take care of writing the logs to Stackdriver.
func NewComputeEngineLogger() (*Logger, error) {
	// See about using https://godoc.org/cloud.google.com/go/logging#CommonResource
	// with values from:
	//https://cloud.google.com/logging/docs/api/v2/resource-list#resource-types

	//TODO

	return nil, nil
}

// NewCloudFunctionLogger returns a Logger suitable for use in Google
// Cloud Functions. It will emit the logs using the Stackdriver API.
func NewCloudFunctionLogger() (*Logger, error) {
	// See about using https://godoc.org/cloud.google.com/go/logging#CommonResource
	// with values from:
	//https://cloud.google.com/logging/docs/api/v2/resource-list#resource-types

	//TODO

	return nil, nil
}

// NewAppEngineLogger returns a Logger suitable for use in AppEngine.
// On local dev server it uses the local stdout -logger and in the cloud it
// uses the Stackdriver logger.
func NewAppEngineLogger() (*Logger, error) {
	opts := []LogOption{}

	// We'll write to the default GAE request log
	logID := "appengine.googleapis.com/request_log"

	if os.Getenv("NODE_ENV") == "production" {
		projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
		if projectID == "" {
			return nil, fmt.Errorf("env var GOOGLE_CLOUD_PROJECT missing")
		}

		service := os.Getenv("GAE_SERVICE")
		if service == "" {
			return nil, fmt.Errorf("env var GAE_SERVICE missing")
		}

		version := os.Getenv("GAE_VERSION")
		if version == "" {
			return nil, fmt.Errorf("env var GAE_VERSION missing")
		}

		// Create a monitored resource descriptor that will target GAE
		monitoredRes := &monitoredres.MonitoredResource{
			Type: "gae_app",
			Labels: map[string]string{
				"project_id": projectID,
				"module_id":  service,
				"version_id": version,
			},
		}

		opts = append(opts, WithStackdriver(projectID, "", logID, monitoredRes))
	} else {
		opts = append(opts, WithLocal())
	}

	return NewLogger(opts...)
}

func (l *Logger) updateLogrusLoglevel() {
	if l.logrusLogger == nil {
		return
	}

	switch l.logLevel {
	case Trace:
		l.logrusLogger.SetLevel(logrus.TraceLevel)
	case Debug:
		l.logrusLogger.SetLevel(logrus.DebugLevel)
	case Info:
		l.logrusLogger.SetLevel(logrus.InfoLevel)
	case Warning:
		l.logrusLogger.SetLevel(logrus.WarnLevel)
	case Error:
		l.logrusLogger.SetLevel(logrus.ErrorLevel)
	case Fatal:
		l.logrusLogger.SetLevel(logrus.FatalLevel)
	}
}

// SetLogLevel sets the log levels of the underlying logger interfaces.
// Note that this operation is not mutexed and thus not thread-safe.
func (l *Logger) SetLogLevel(logLevel int) {
	l.logLevel = logLevel
	l.updateLogrusLoglevel()
}

// SetStackdriverLogger can be used to set a customized *stackdriver.Logger.
// Note that this operation is not mutexed and thus not thread-safe.
func (l *Logger) SetStackdriverLogger(logger *stackdriver.Logger) {
	l.stackdriverLogger = logger
}

// SetLogrusLogger can be used to set a customized *logrus.Logger.
// Note that this operation is not mutexed and thus not thread-safe.
func (l *Logger) SetLogrusLogger(logger *logrus.Logger) {
	l.logrusLogger = logger
	l.updateLogrusLoglevel()
}

// LogrusLogger exposes our logrus logger.
// Note that this operation is not mutexed and thus not thread-safe.
func (l *Logger) LogrusLogger() *logrus.Logger {
	return l.logrusLogger
}

// Close closes the logger and flushes the underlying loggers'
// buffers. Returns error if there are errors.
func (l *Logger) Close() error {
	if l.stackdriverClient != nil {
		if err := l.stackdriverClient.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Flush flushes the underlying loggers' buffers. Returns error if
// there are errors.
func (l *Logger) Flush() error {
	if l.stackdriverLogger != nil {
		if err := l.stackdriverLogger.Flush(); err != nil {
			return err
		}
	}

	return nil
}

// levelToSeverity maps our log level to Stackdriver Severity
func levelToSeverity(level int) stackdriver.Severity {
	switch level {
	case Trace:
		return stackdriver.Debug
	case Debug:
		return stackdriver.Debug
	case Info:
		return stackdriver.Info
	case Warning:
		return stackdriver.Warning
	case Error:
		return stackdriver.Error
	case Fatal:
		return stackdriver.Critical
	default:
		return stackdriver.Default
	}
}

// levelToLogFunc maps our log level to logrus log func
func levelToLogFunc(level int, logger *logrus.Logger) logFunc {
	switch level {
	case Trace:
		return logger.Tracef
	case Debug:
		return logger.Debugf
	case Info:
		return logger.Infof
	case Warning:
		return logger.Warningf
	case Error:
		return logger.Errorf
	case Fatal:
		return logger.Fatalf
	default:
		return nil
	}
}

// Actually writes the log entries
func (l *Logger) logImpl(level int, format string, args ...interface{}) {
	if level < l.logLevel {
		return
	}

	// Stackdriver logging - if enabled
	if l.stackdriverLogger != nil {
		// payload := map[string]string{
		// 	"Message": fmt.Sprintf(format, args...),

		// 	//TODO we could support structured logs by adding the args here
		// }

		l.stackdriverLogger.Log(stackdriver.Entry{
			Payload:  fmt.Sprintf(format, args...),
			Severity: levelToSeverity(level),
		})
	}

	// Local logging - if enabled
	if l.logrusLogger != nil {
		levelToLogFunc(level, l.logrusLogger)(format, args...)
	}
}

// Tracef writes trace level logs
func (l *Logger) Tracef(format string, args ...interface{}) {
	l.logImpl(Trace, format, args...)
}

// Debugf writes debug level logs
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logImpl(Debug, format, args...)
}

// Printf writes debug level logs - included for compatibility with
// the standard "log" package.
func (l *Logger) Printf(format string, args ...interface{}) {
	l.logImpl(Debug, format, args...)
}

// Infof writes info level logs
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logImpl(Info, format, args...)
}

// Warningf writes warning level logs
func (l *Logger) Warningf(format string, args ...interface{}) {
	l.logImpl(Warning, format, args...)
}

// Errorf writes error level logs
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logImpl(Error, format, args...)
}

// Fatalf writes fatal level logs
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logImpl(Fatal, format, args...)
}
