// Package cloudlogging provides logging wrapper for Google Cloud Platform
// environments.
package cloudlogging

import (
	"fmt"
	stdlog "log"
	"os"

	gcloudlog "cloud.google.com/go/logging"
	"github.com/qvik/go-cloudlogging/internal"
	"go.uber.org/zap"
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
// the Google Cloud Logging cloud logs. Logger is mostly immutable - the only thing
// that can be modified is the log level.
//
// Logger is thread-safe to use for any of the logging calls.
// Some methods such as SetLogLevel(), Close(), Flush() however are not and
// you should either call them at the start / end of your program or synchronize
// your access to the Logger instance if setting on them on the fly. The
// reasoning here being that standard Logger calls should not suffer a
// performance hit from locking.
//
// Logger uses the Zap (https://github.com/uber-go/zap)
// library for local logging.
//
// Logger uses Google Cloud Logging Go library for cloud logging.
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
	zapConfig *zap.Config
	zapLogger *zap.SugaredLogger

	// Google Cloud Logging client
	googleCloudLoggingClient *gcloudlog.Client

	// Google Cloud Logging logger
	googleCloudLoggingLogger *gcloudlog.Logger

	// Common log parameters. These are added to every structured log message
	// in addition to the parameters issued in the actual logging call.
	// Notice that this only applies to structured logging
	// and not the formatted logging (eg. Debug(), but not Debugf()).
	// The format is: key1, value1, key2, value2, ...
	commonKeysAndValues map[interface{}]interface{}

	// When set, the logger emits all Google Cloud Logging here instead of the actual
	// logger. This is meant to be used in unit testing.
	googleCloudLoggingDebugHook func(gcloudlog.Entry)
}

// WithAdditionalKeysAndValues creates a new logger that uses the current
// logger as its base logger (all logging calls will be forwarded to the base
// logger). Making changes to the base logger will be reflected in any
// calls to the new logger. Additional keys and values may be added for
// structured logging purposes.
// This is a light operation.
// Panics if number of elements in keysAndValues is not even.
// Panics on internal errors.
func (l *Logger) WithAdditionalKeysAndValues(
	keysAndValues ...interface{}) *Logger {

	if len(keysAndValues) == 0 {
		stdlog.Panicf("must pass keys and values")
	}

	if len(keysAndValues)%2 != 0 {
		stdlog.Panicf("must pass even number of keysAndValues")
	}

	// Create a new logger object which is an exact copy of its base,
	// but a fresh object.
	newLogger := *l

	// Make a new map for the keys and values
	newLogger.commonKeysAndValues = make(map[interface{}]interface{})
	for k, v := range l.commonKeysAndValues {
		newLogger.commonKeysAndValues[k] = v
	}

	// Apply the added common keys and values
	internal.MustApplyKeysAndValues(keysAndValues, newLogger.commonKeysAndValues)

	// Create a new Zap logger which wraps the new properties
	if newLogger.zapLogger != nil {
		zapLogger, err := newLogger.zapConfig.Build()
		if err != nil {
			stdlog.Panicf("failed to create new zaplogger: %v", err)
		}

		keysAndValues := internal.MapToKeysAndValuesList(newLogger.commonKeysAndValues)
		newLogger.zapLogger = zapLogger.Sugar().With(keysAndValues...)
	}

	return &newLogger
}

// NewLogger creates a new Logger instance using the given options.
// The default log level is Debug.
func NewLogger(opt ...LogOption) (*Logger, error) {
	opts := options{logLevel: Debug}

	for _, o := range opt {
		o.apply(&opts)
	}

	if opts.useGoogleCloudLogging && opts.gcpProjectID == "" {
		return nil, fmt.Errorf("google cloud logging requires a GCP project ID")
	}

	var googleCloudLoggingClient *gcloudlog.Client
	var googleCloudLoggingLogger *gcloudlog.Logger
	var zapConfig *zap.Config
	var zapLogger *zap.SugaredLogger

	if opts.useGoogleCloudLogging {
		if opts.googleCloudLoggingUnitTestHook != nil {
			googleCloudLoggingClient = &gcloudlog.Client{}
			googleCloudLoggingLogger = &gcloudlog.Logger{}
		} else {
			client, logger, err := createGoogleCloudLoggingLogger(opts)
			if err != nil {
				return nil, fmt.Errorf("failed to create google cloud logging log: %w", err)
			}

			googleCloudLoggingClient = client
			googleCloudLoggingLogger = logger
		}
	}

	if opts.useZap {
		stdlog.Printf("Creating local ZAP logger.")

		logger, config, err := createZapLogger(opts)
		if err != nil {
			return nil, fmt.Errorf("failed to create Zap logger: %w", err)
		}

		zapConfig = config
		zapLogger = logger.Sugar()

		// Add the initial common labels, if any
		if len(opts.commonKeysAndValues) > 0 {
			keysAndValues := internal.MapToKeysAndValuesList(opts.commonKeysAndValues)
			zapLogger = zapLogger.With(keysAndValues...)
		}
	}

	l := &Logger{
		logLevel:                    opts.logLevel,
		googleCloudLoggingClient:    googleCloudLoggingClient,
		googleCloudLoggingLogger:    googleCloudLoggingLogger,
		zapConfig:                   zapConfig,
		zapLogger:                   zapLogger,
		commonKeysAndValues:         opts.commonKeysAndValues,
		googleCloudLoggingDebugHook: opts.googleCloudLoggingUnitTestHook,
	}

	return l, nil
}

// MustNewLogger creates a new Logger instance using the given options.
// The default log level is Debug.
// Panics if logger creation fails.
func MustNewLogger(opt ...LogOption) *Logger {
	logger, err := NewLogger(opt...)
	if err != nil {
		stdlog.Fatalf("failed to create logger: %v", err)
	}

	return logger
}

// SetLogLevel sets the log levels of the underlying logger interfaces.
// Note that this operation is not mutexed and thus not inherently thread-safe.
func (l *Logger) SetLogLevel(logLevel Level) *Logger {
	l.logLevel = logLevel

	if l.zapLogger != nil {
		// Adjust zap's atomic level
		setZapLogLevel(l.zapConfig, logLevel)
	}

	return l
}

// Close closes the logger and flushes the underlying loggers'
// buffers. Returns error if there are errors.
func (l *Logger) Close() error {
	// Attempt to flush the loggers' buffers; nevermind errors
	_ = l.Flush()

	if l.googleCloudLoggingClient != nil {
		if err := l.googleCloudLoggingClient.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Flush flushes the underlying loggers' buffers. Returns error if
// there are errors.
func (l *Logger) Flush() error {
	if l.googleCloudLoggingLogger != nil {
		if err := l.googleCloudLoggingLogger.Flush(); err != nil {
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

	// Emit Google Cloud Logging logging - if enabled
	if l.googleCloudLoggingLogger != nil {
		severity := gcloudlog.Default
		if s, ok := levelToGoogleCloudLoggingSeverityMap[level]; ok {
			severity = s
		}

		l.googleCloudLoggingLogger.Log(gcloudlog.Entry{
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
func (l *Logger) logImpl(level Level, payload interface{},
	keysAndValues ...interface{}) {

	if len(keysAndValues)%2 != 0 {
		stdlog.Panicf("must pass even number of keysAndValues")
	}

	// Emit Google Cloud Logging logging - if enabled
	if l.googleCloudLoggingLogger != nil {
		severity := gcloudlog.Default
		if s, ok := levelToGoogleCloudLoggingSeverityMap[level]; ok {
			severity = s
		}

		labels := make(map[string]string, len(l.commonKeysAndValues)+len(keysAndValues))

		for key, value := range l.commonKeysAndValues {
			if stringKey, ok := key.(string); ok {
				if stringValue, ok := value.(string); ok {
					labels[stringKey] = stringValue
				} else {
					labels[stringKey] = fmt.Sprint(value)
				}
			} else {
				labels[fmt.Sprint(key)] = fmt.Sprint(value)
			}
		}

		count := 0
		for count < len(keysAndValues) {
			key := keysAndValues[count]
			value := keysAndValues[count+1]

			if stringKey, ok := key.(string); ok {
				if stringValue, ok := value.(string); ok {
					labels[stringKey] = stringValue
				} else {
					labels[stringKey] = fmt.Sprint(value)
				}
			} else {
				labels[fmt.Sprint(key)] = fmt.Sprint(value)
			}

			count += 2
		}

		entry := gcloudlog.Entry{
			Payload:  payload,
			Labels:   labels,
			Severity: severity,
		}

		if l.googleCloudLoggingDebugHook != nil {
			l.googleCloudLoggingDebugHook(entry)
		} else {
			l.googleCloudLoggingLogger.Log(entry)
		}
	}

	// Emit local logging - if enabled
	if l.zapLogger != nil {
		f := levelToZapStructuredLogFunc(level, l.zapLogger)
		if f != nil {
			f(fmt.Sprintf("%+v", payload), keysAndValues...)
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
