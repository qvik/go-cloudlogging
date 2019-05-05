package cloudlogging

import (
	"io"

	"google.golang.org/genproto/googleapis/api/monitoredres"
)

type options struct {
	logLevel                     Level
	gcpProjectID                 string
	credentialsFilePath          string
	useLocal                     bool
	useLocalFluentD              bool
	useStackdriver               bool
	stackdriverLogID             string
	stackDriverMonitoredResource *monitoredres.MonitoredResource
	localOutput                  io.Writer
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

type withLocalOutput struct {
	output io.Writer
}

func (w withLocalOutput) apply(opts *options) {
	opts.localOutput = w.output
}

// WithLocalOutput returns a LogOption that redirects the output to the
// given io.Writer. The default ouput is os.Stdout.
func WithLocalOutput(output io.Writer) LogOption {
	return withLocalOutput{output: output}
}

type withLocal bool

func (w withLocal) apply(opts *options) {
	opts.useLocal = true
}

// WithLocal returns a LogOption that enables the local logger
// and configures to use the standard output.
func WithLocal() LogOption {
	return withLocal(true)
}

type withLocalFluentD bool

func (w withLocalFluentD) apply(opts *options) {
	opts.useLocal = true
	opts.useLocalFluentD = true
}

// WithLocalFluentD returns a LogOption that enables local logger
// and configures it to use FluentD output.
func WithLocalFluentD() LogOption {
	return withLocalFluentD(true)
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
