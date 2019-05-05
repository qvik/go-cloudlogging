// Package cloudlogging provides logging wrapper for Google Cloud Platform
// environments.
package cloudlogging

import (
	"fmt"
	"os"

	stackdriver "cloud.google.com/go/logging"
	zap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/genproto/googleapis/api/monitoredres"
)

//Level is our log level type
type Level int8

// Log levels
const (
	Debug Level = iota
	Info
	Warning
	Error
	Fatal
)

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
// - https://github.com/uber-go/zap
type Logger struct {
	// Current log level
	logLevel Level

	// Zap logger
	zapLogger *zap.SugaredLogger

	// Zap atomic logging level handle
	zapLevel zap.AtomicLevel

	// Stackdriver client
	stackdriverClient *stackdriver.Client

	// Stackdriver logger
	stackdriverLogger *stackdriver.Logger
}

// NewLogger creates a new Logger instance using the given options.
// The default log level is Info.
func NewLogger(opt ...LogOption) (*Logger, error) {
	opts := options{logLevel: Info}

	for _, o := range opt {
		o.apply(&opts)
	}

	if opts.useStackdriver && opts.gcpProjectID == "" {
		return nil, fmt.Errorf("Stackdriver requires a GCP project ID")
	}

	var stackdriverClient *stackdriver.Client
	var stackdriverLogger *stackdriver.Logger
	var zapLogger *zap.SugaredLogger
	var zapLevel zap.AtomicLevel

	if opts.useStackdriver {
		client, logger, err := createStackdriverLogger(opts)
		if err != nil {
			return nil, err
		}
		stackdriverClient = client
		stackdriverLogger = logger
	}

	if opts.useLocal {
		logger, zl, err := createZapLogger(opts)
		if err != nil {
			return nil, err
		}

		zapLogger = logger.Sugar()
		zapLevel = zl
	}

	l := &Logger{
		logLevel:          opts.logLevel,
		stackdriverClient: stackdriverClient,
		stackdriverLogger: stackdriverLogger,
		zapLogger:         zapLogger,
		zapLevel:          zapLevel,
	}

	return l, nil
}

// NewLocalOnlyLogger returns a logger that only logs to the local
// standard output.
func NewLocalOnlyLogger() (*Logger, error) {
	return NewLogger(WithLocal())
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

// SetLogLevel sets the log levels of the underlying logger interfaces.
// Note that this operation is not mutexed and thus not inherently thread-safe.
func (l *Logger) SetLogLevel(logLevel Level) {
	l.logLevel = logLevel

	if l.zapLogger != nil {
		// Adjust zap's atomic level
		zapLevel := zapcore.InfoLevel
		if l, ok := levelToZapLevelMap[logLevel]; ok {
			zapLevel = l
		}
		l.zapLevel.SetLevel(zapLevel)
	}
}

// Close closes the logger and flushes the underlying loggers'
// buffers. Returns error if there are errors.
func (l *Logger) Close() error {
	// Attempt to flush the loggers' buffers; nevermind errors
	_ = l.Flush()

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

	if l.zapLogger != nil {
		if err := l.zapLogger.Sync(); err != nil {
			return err
		}
	}

	return nil
}

// Writes a flat log entry.
func (l *Logger) logImplf(level Level, format string, args ...interface{}) {
	if level < l.logLevel {
		return
	}

	// Emit Stackdriver logging - if enabled
	if l.stackdriverLogger != nil {
		severity := stackdriver.Default
		if s, ok := levelToStackdriverSeverityMap[level]; ok {
			severity = s
		}

		l.stackdriverLogger.Log(stackdriver.Entry{
			Payload:  fmt.Sprintf(format, args...),
			Severity: severity,
		})
	}

	// Emit local logging - if enabled
	if l.zapLogger != nil {
		f := levelToZapFlatLogFunc(level, l.zapLogger)
		if f != nil {
			f(format, args...)
		}
	}
}

// Writes a structured log entry.
func (l *Logger) logImpl(level Level, msg string,
	keysAndValues ...interface{}) {

	//TODO
}

// FLAT LOGGING

// Tracef writes debug level logs - it exists to provide API compatibility
// with some logging libraries.
func (l *Logger) Tracef(format string, args ...interface{}) {
	l.logImplf(Debug, format, args...)
}

// Debugf writes debug level logs
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logImplf(Debug, format, args...)
}

// Printf writes debug level logs - included for compatibility with
// the standard "log" package.
func (l *Logger) Printf(format string, args ...interface{}) {
	l.logImplf(Debug, format, args...)
}

// Infof writes info level logs
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logImplf(Info, format, args...)
}

// Warningf writes warning level logs
func (l *Logger) Warningf(format string, args ...interface{}) {
	l.logImplf(Warning, format, args...)
}

// Errorf writes error level logs
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logImplf(Error, format, args...)
}

// Fatalf writes fatal level logs and calls os.Exit(1)
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logImplf(Fatal, format, args...)

	// Fatal log; the program execution should stop. If the local logger
	// is in use, it has already done this; otherwise we will need to do
	// it ourselves
	if l.zapLogger == nil {
		os.Exit(1)
	}
}

// STRUCTURED LOGGING

// Debug writes a structured log entry using the debug level.
func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	l.logImpl(Debug, msg, keysAndValues)
}
