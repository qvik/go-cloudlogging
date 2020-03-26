package zap

import (
	stdlog "log"

	"github.com/qvik/go-cloudlogging"
)

// NewZapLogger returns a logger that only logs to a local stdout/stderr
// Zap logger.
func NewZapLogger() (*cloudlogging.Logger, error) {
	return cloudlogging.NewLogger(cloudlogging.WithZap())
}

// MustNewZapLogger returns a logger that only logs to a local stdout/stderr
// Zap logger.
// Panics on error(s).
func MustNewZapLogger() *cloudlogging.Logger {
	log, err := NewZapLogger()
	if err != nil {
		stdlog.Panicf("failed to create logger: %v", err)
	}

	return log
}
