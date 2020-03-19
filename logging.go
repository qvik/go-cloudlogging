// Package cloudlogging provides logging wrapper for Google Cloud Platform
// environments.
package cloudlogging

import (
	"fmt"
	stdlog "log"
	"os"

	stackdriver "cloud.google.com/go/logging"
	zap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/genproto/googleapis/api/monitoredres"
)

// Level is our log level type
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
// Logger uses the Zap (https://github.com/uber-go/zap)
// library for local logging.
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

	// Default log parameters. These are added to every structured log message
	// in addition to the parameters issued in the actual logging call.
	// Notice that this only applies to structured logging
	// and not the formatted logging (eg. Debug(), but not Debugf()).
	// The format is: key1, value1, key2, value2, ...
	defaultKeysAndValues []interface{}

	// Base ("root") logger to use for actual low-level logging if defined.
	// If it is defined, all logging calls are forwarded to it, but all
	// structured logging calls use the defaultKeysAndValues defined
	// by the current instance.
	baseLog *Logger
}

// WithAdditionalKeysAndValues creates a new logger that uses the current
// logger as its base logger (all logging calls will be forwarded to the base
// logger). Making changes to the base logger will be reflected in any
// calls to the new logger. Additional keys and values may be added for
// structured logging purposes.
// Panics if number of elements in keysAndValues is not even.
func (l *Logger) WithAdditionalKeysAndValues(
	keysAndValues ...interface{}) *Logger {

	if len(keysAndValues)%2 != 0 {
		stdlog.Panicf("must pass even number of keysAndValues")
	}

	// Find the "root" base logger
	base := l
	for base.baseLog != nil {
		base = base.baseLog
	}

	newLogger := &Logger{baseLog: base}
	newLogger.defaultKeysAndValues =
		append(base.defaultKeysAndValues, keysAndValues...)

	return newLogger
}

// NewLogger creates a new Logger instance using the given options.
// The default log level is Info.
func NewLogger(opt ...LogOption) (*Logger, error) {
	opts := options{logLevel: Debug}

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
			return nil, fmt.Errorf("failed to create Stackdriver log: %v", err)
		}

		stackdriverClient = client
		stackdriverLogger = logger
	}

	if opts.useLocal {
		stdlog.Printf("Creating local ZAP logger.")

		logger, zl, err := createZapLogger(opts)
		if err != nil {
			return nil, fmt.Errorf("failed to create Zap logger: %v", err)
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

// MustNewLocalOnlyLogger creates a new local stdout only logger or panics.
func MustNewLocalOnlyLogger() *Logger {
	log, err := NewLocalOnlyLogger()
	if err != nil {
		stdlog.Panicf("failed to create logger: %v", err)
	}

	return log
}

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
	if arg0, ok := getArg(0, args...); ok && arg0 != "" {
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

	opts = append(opts, WithStackdriver(projectID, "", logID, monitoredRes))

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
// On local dev server it uses the local stdout -logger and in the cloud it
// uses the Stackdriver logger.
// The first value of args is the logID. If omitted or empty string is given,
// the default value of "appengine.googleapis.com/request_log" is used.
func NewAppEngineLogger(args ...string) (*Logger, error) {
	opts := []LogOption{}

	logID := "appengine.googleapis.com/request_log"
	if arg0, ok := getArg(0, args...); ok && arg0 != "" {
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

		opts = append(opts, WithStackdriver(projectID, "", logID, monitoredRes))
	} else {
		opts = append(opts, WithLocal())
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

// SetLogLevel sets the log levels of the underlying logger interfaces.
// Note that this operation is not mutexed and thus not inherently thread-safe.
func (l *Logger) SetLogLevel(logLevel Level) *Logger {
	l.logLevel = logLevel

	if l.zapLogger != nil {
		// Adjust zap's atomic level
		zapLevel := zapcore.InfoLevel
		if l, ok := levelToZapLevelMap[logLevel]; ok {
			zapLevel = l
		}
		l.zapLevel.SetLevel(zapLevel)
	}

	return l
}

// SetDefaultKeysAndValues sets the set of default structured logging
// keys and values added to every (structured) logging call.
// Format is: key1, value1, key2, value2, ..
// Panics if number of elements in keysAndValues is not even.
func (l *Logger) SetDefaultKeysAndValues(
	keysAndValues ...interface{}) {

	if len(keysAndValues)%2 != 0 {
		stdlog.Panicf("must pass even number of keysAndValues")
	}

	l.defaultKeysAndValues = keysAndValues
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
	// Use base logger if it is defined
	log := l
	if l.baseLog != nil {
		log = l.baseLog
	}

	if level < log.logLevel {
		return
	}

	// Emit Stackdriver logging - if enabled
	if log.stackdriverLogger != nil {
		severity := stackdriver.Default
		if s, ok := levelToStackdriverSeverityMap[level]; ok {
			severity = s
		}

		log.stackdriverLogger.Log(stackdriver.Entry{
			Payload:  fmt.Sprintf(format, args...),
			Severity: severity,
		})
	}

	// Emit local logging - if enabled
	if log.zapLogger != nil {
		f := levelToZapFlatLogFunc(level, log.zapLogger)
		if f != nil {
			f(format, args...)
		}
	}
}

// Writes a structured log entry.
func (l *Logger) logImpl(level Level, payload interface{},
	keysAndValues ...interface{}) {

	if len(keysAndValues)%2 != 0 {
		stdlog.Panicf("must pass even number of keysAndValues")
	}

	// Use base logger if it is defined
	log := l
	if l.baseLog != nil {
		log = l.baseLog
	}

	// Emit Stackdriver logging - if enabled
	if log.stackdriverLogger != nil {
		severity := stackdriver.Default
		if s, ok := levelToStackdriverSeverityMap[level]; ok {
			severity = s
		}

		// Create the labels map from the param keys and values
		labels := map[string]string{}
		for i := 0; i < len(keysAndValues); i += 2 {
			key := fmt.Sprintf("%v", keysAndValues[i])
			value := fmt.Sprintf("%+v", keysAndValues[i+1])

			labels[key] = value
		}

		// Add the default set of param keys and values.
		// Note that here we want to use l.defaultKeysAndValues
		// instead of log.defaultKeysAndValues.
		for i := 0; i < len(l.defaultKeysAndValues); i += 2 {
			key := fmt.Sprintf("%v", l.defaultKeysAndValues[i])
			value := fmt.Sprintf("%+v", l.defaultKeysAndValues[i+1])

			labels[key] = value
		}

		log.stackdriverLogger.Log(stackdriver.Entry{
			Payload:  payload,
			Labels:   labels,
			Severity: severity,
		})
	}

	// Emit local logging - if enabled
	if log.zapLogger != nil {
		f := levelToZapStructuredLogFunc(level, log.zapLogger)
		if f != nil {
			// Add the default set of param keys and values.
			// Note that here we want to use l.defaultKeysAndValues
			// instead of log.defaultKeysAndValues.
			params := append(l.defaultKeysAndValues, keysAndValues...)
			f(fmt.Sprintf("%+v", payload), params...)
		}
	}
}

// FLAT LOGGING

// Tracef writes debug level logs.
// Compatibility alias for Debugf().
func (l *Logger) Tracef(format string, args ...interface{}) {
	l.logImplf(Debug, format, args...)
}

// Debugf writes debug level logs
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logImplf(Debug, format, args...)
}

// Printf writes debug level logs.
// Compatibility alias for Debugf().
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

// Panicf writes fatal level logs and exits.
// Compatibility alias for Fatalf().
func (l *Logger) Panicf(format string, args ...interface{}) {
	l.Fatalf(format, args...)
}

// STRUCTURED LOGGING

// Trace writes a structured log entry using the debug level.
func (l *Logger) Trace(payload interface{}, keysAndValues ...interface{}) {
	l.logImpl(Debug, payload, keysAndValues...)
}

// Debug writes a structured log entry using the debug level.
// Compatibility alias for Debug().
func (l *Logger) Debug(payload interface{}, keysAndValues ...interface{}) {
	l.logImpl(Debug, payload, keysAndValues...)
}

// Print writes a structured log entry using the debug level.
// Compatibility alias for Debug().
func (l *Logger) Print(payload interface{}, keysAndValues ...interface{}) {
	l.logImpl(Debug, payload, keysAndValues...)
}

// Info writes a structured log entry using the info level.
func (l *Logger) Info(payload interface{}, keysAndValues ...interface{}) {
	l.logImpl(Info, payload, keysAndValues...)
}

// Warning writes a structured log entry using the warning level.
func (l *Logger) Warning(payload interface{}, keysAndValues ...interface{}) {
	l.logImpl(Warning, payload, keysAndValues...)
}

// Error writes a structured log entry using the error level.
func (l *Logger) Error(payload interface{}, keysAndValues ...interface{}) {
	l.logImpl(Error, payload, keysAndValues...)
}

// Fatal writes a structured log entry using the fatal level.
func (l *Logger) Fatal(payload interface{}, keysAndValues ...interface{}) {
	l.logImpl(Fatal, payload, keysAndValues...)
}

// Panic writes a structured log entry using the fatal level.
// Compatibility alias for Fatal().
func (l *Logger) Panic(payload interface{}, keysAndValues ...interface{}) {
	l.logImpl(Fatal, payload, keysAndValues...)
}
