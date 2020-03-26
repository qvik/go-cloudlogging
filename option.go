package cloudlogging

import (
	"github.com/qvik/go-cloudlogging/internal"
	"google.golang.org/genproto/googleapis/api/monitoredres"
)

type options struct {
	logLevel                     Level
	gcpProjectID                 string
	credentialsFilePath          string
	useZap                       bool
	zapOutputPath                string
	zapErrorOutputPath           string
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

type withZap struct {
	// Path of output log file; default is stdout
	outputPath string

	// Path of output errors log file; default is stdout
	errorOutputPath string
}

func (w withZap) apply(opts *options) {
	opts.useZap = true
	opts.zapOutputPath = w.outputPath
	opts.zapErrorOutputPath = w.errorOutputPath
}

// WithZap returns a LogOption that enables the local Zap logger.
// To override log output. setpaths to point to
// a log file(s). First argument should normal output log file path and the second
// should be the error output log file path. The defaults are stdout / stderr.
func WithZap(paths ...string) LogOption {
	outputPath, _ := internal.GetArg(0, paths...)
	errorOutputPath, _ := internal.GetArg(1, paths...)

	return withZap{
		outputPath:      outputPath,
		errorOutputPath: errorOutputPath,
	}
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
