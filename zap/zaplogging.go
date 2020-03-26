package zap

import (
	stdlog "log"

	"github.com/qvik/go-cloudlogging"
)

// NewZapLogger returns a logger that only logs to a local Zap logger.
// To override log output. setpaths to point to
// a log file(s). First argument should normal output log file path and the second
// should be the error output log file path. The defaults are stdout / stderr.
func NewZapLogger(paths ...string) (*cloudlogging.Logger, error) {
	return cloudlogging.NewLogger(cloudlogging.WithZap(paths...))
}

// MustNewZapLogger returns a logger that only logs to a local Zap logger.
// Panics on error(s).
// To override log output. setpaths to point to
// a log file(s). First argument should normal output log file path and the second
// should be the error output log file path. The defaults are stdout / stderr.
func MustNewZapLogger(paths ...string) *cloudlogging.Logger {
	log, err := NewZapLogger(paths...)
	if err != nil {
		stdlog.Panicf("failed to create logger: %v", err)
	}

	return log
}
