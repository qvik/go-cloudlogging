package cloudlogging

import (
	stdlog "log"

	gcloudlog "cloud.google.com/go/logging"
	"github.com/qvik/go-cloudlogging/internal"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/api/monitoredres"
)

// OutputHint provides a mechanism to instruct log backends to
// adjust their output / formatting. Not all log backends react to any / all
// of the hints.
type OutputHint int32

const (
	// JSONFormat output hint requests the log backend to output JSON(NL).
	JSONFormat OutputHint = iota
)

type options struct {
	logLevel                            Level
	gcpProjectID                        string
	credentialsFilePath                 string
	useZap                              bool
	zapConfig                           *zap.Config
	outputPaths                         []string
	errorOutputPaths                    []string
	outputHints                         []OutputHint
	useGoogleCloudLogging               bool
	googleCloudLoggingLogID             string
	googleCloudLoggingMonitoredResource *monitoredres.MonitoredResource
	commonKeysAndValues                 map[interface{}]interface{}
	googleCloudLoggingUnitTestHook      func(gcloudlog.Entry)
}

// LogOption is an option for the cloudlogging API.
type LogOption interface {
	apply(*options)
}

type withGoogleCloudLoggingUnitTestHook func(gcloudlog.Entry)

func (w withGoogleCloudLoggingUnitTestHook) apply(opts *options) {
	opts.googleCloudLoggingUnitTestHook = w
}

type withLevel Level

func (w withLevel) apply(opts *options) {
	opts.logLevel = Level(w)
}

// WithLevel returns a LogOption that defines our log level.
func WithLevel(logLevel Level) LogOption {
	return withLevel(logLevel)
}

type withOutputPaths []string

func (w withOutputPaths) apply(opts *options) {
	opts.outputPaths = w
}

// WithOutputPaths log output paths, eg. URLs or file paths. Different
// loggers use these in different ways. For example Zap uses them as
// log output file paths.
func WithOutputPaths(paths ...string) LogOption {
	return withOutputPaths(paths)
}

type withErrorOutputPaths []string

func (w withErrorOutputPaths) apply(opts *options) {
	opts.errorOutputPaths = w
}

// WithErrorOutputPaths log output paths, eg. URLs or file paths. Different
// loggers use these in different ways. For example Zap uses them as
// log output file paths.
func WithErrorOutputPaths(paths ...string) LogOption {
	return withErrorOutputPaths(paths)
}

type withOutputHints []OutputHint

func (w withOutputHints) apply(opts *options) {
	opts.outputHints = w
}

// WithOutputHints adds output hints to the log backend.
func WithOutputHints(hints ...OutputHint) LogOption {
	return withOutputHints(hints)
}

type withZap struct {
	zapConfig *zap.Config
}

func (w withZap) apply(opts *options) {
	opts.useZap = true
	opts.zapConfig = w.zapConfig
}

// WithZap returns a LogOption that enables the local Zap logger, optionally
// with a Zap configuration.
func WithZap(config ...*zap.Config) LogOption {
	var cfg *zap.Config
	if len(config) > 0 {
		cfg = config[0]
	}

	return withZap{zapConfig: cfg}
}

type withGoogleCloudLogging struct {
	gcpProjectID            string
	credentialsFilePath     string
	googleCloudLoggingLogID string
	monitoredResource       *monitoredres.MonitoredResource
}

func (w withGoogleCloudLogging) apply(opts *options) {
	opts.useGoogleCloudLogging = true
	opts.gcpProjectID = w.gcpProjectID
	opts.googleCloudLoggingLogID = w.googleCloudLoggingLogID
	opts.credentialsFilePath = w.credentialsFilePath
	opts.googleCloudLoggingMonitoredResource = w.monitoredResource
}

// WithGoogleCloudLogging returns a LogOption that enables Google Cloud Logging Logger
// and configures it to use the given project id.
// If you supply empty string for credentialsFilePath, the default
// service account is used.
// Google cloud logging log backend does not react to OutputHints.
func WithGoogleCloudLogging(gcpProjectID, credentialsFilePath,
	googleCloudLoggingLogID string,
	monitoredResource *monitoredres.MonitoredResource) LogOption {

	return withGoogleCloudLogging{
		gcpProjectID:            gcpProjectID,
		credentialsFilePath:     credentialsFilePath,
		googleCloudLoggingLogID: googleCloudLoggingLogID,
		monitoredResource:       monitoredResource,
	}
}

type withCommonKeysAndValues []interface{}

func (w withCommonKeysAndValues) apply(opts *options) {
	opts.commonKeysAndValues = make(map[interface{}]interface{})
	internal.MustApplyKeysAndValues(w, opts.commonKeysAndValues)
}

// WithCommonKeysAndValues returns a LogOption that adds a set of
// common keys and values (labels / fields) to all structured log messages.
// For parameters should be: key1, value1, key2, value2, ..
func WithCommonKeysAndValues(commonKeysAndValues ...interface{}) LogOption {
	if len(commonKeysAndValues)%2 != 0 {
		stdlog.Fatalf("number of keys + values must be even")
	}

	return withCommonKeysAndValues(commonKeysAndValues)
}
