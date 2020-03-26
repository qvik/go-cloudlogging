package cloudlogging

import (
	"google.golang.org/genproto/googleapis/api/monitoredres"
)

type options struct {
	logLevel                     Level
	gcpProjectID                 string
	credentialsFilePath          string
	useZap                       bool
	outputPaths                  []string
	errorOutputPaths             []string
	useStackdriver               bool
	stackdriverLogID             string
	stackDriverMonitoredResource *monitoredres.MonitoredResource
	commonLabels                 map[string]string
}

// LogOption is an option for the cloudlogging API.
type LogOption interface {
	apply(*options)
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

type withZap bool

func (w withZap) apply(opts *options) {
	opts.useZap = true
}

// WithZap returns a LogOption that enables the local Zap logger.
// To override log output. setpaths to point to
// a log file(s). First argument should normal output log file path and the second
// should be the error output log file path. The defaults are stdout / stderr.
func WithZap() LogOption {
	return withZap(true)
}

type withStackdriver struct {
	gcpProjectID        string
	credentialsFilePath string
	stackdriverLogID    string
	monitoredResource   *monitoredres.MonitoredResource
}

func (w withStackdriver) apply(opts *options) {
	opts.useStackdriver = true
	opts.gcpProjectID = w.gcpProjectID
	opts.stackdriverLogID = w.stackdriverLogID
	opts.credentialsFilePath = w.credentialsFilePath
	opts.stackDriverMonitoredResource = w.monitoredResource
}

// WithStackdriver returns a LogOption that enables Stackdriver Logger
// and configures it to use the given project id.
// If you supply empty string for credentialsFilePath, the default
// service account is used.
func WithStackdriver(gcpProjectID, credentialsFilePath,
	stackdriverLogID string,
	monitoredResource *monitoredres.MonitoredResource) LogOption {

	return withStackdriver{
		gcpProjectID:        gcpProjectID,
		credentialsFilePath: credentialsFilePath,
		stackdriverLogID:    stackdriverLogID,
		monitoredResource:   monitoredResource,
	}
}

type withCommonLabels struct {
	commonLabels map[string]string
}

func (w withCommonLabels) apply(opts *options) {
	opts.commonLabels = w.commonLabels
}

// WithCommonLabels returns a LogOption that adds a set of string=string
// labels (fields) to all structured log messages.
func WithCommonLabels(commonLabels map[string]string) LogOption {
	return withCommonLabels{commonLabels: commonLabels}
}
