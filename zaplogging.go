package cloudlogging

import (
	stdlog "log"

	"go.uber.org/zap"
)

// NewZapLogger returns a logger that only logs to a local stdout/stderr
// Zap logger, optionally with a Zap configuration.
func NewZapLogger(config ...*zap.Config) (*Logger, error) {
	return NewLogger(WithZap(config...))
}

// MustNewZapLogger returns a logger that only logs to a local stdout/stderr
// Zap logger, optionally with a Zap configuration.
// Panics on error(s).
func MustNewZapLogger(config ...*zap.Config) *Logger {
	log, err := NewZapLogger(config...)
	if err != nil {
		stdlog.Panicf("failed to create logger: %v", err)
	}

	return log
}
