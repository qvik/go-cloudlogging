package cloudlogging

import (
	zap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	levelToZapLevelMap map[Level]zapcore.Level
)

// Zap SugaredLogger logger function
type logFunc func(string, ...interface{})

func createZapLogger(opts options) (*zap.Logger, zap.AtomicLevel, error) {
	zapLevel := zapcore.InfoLevel
	if l, ok := levelToZapLevelMap[opts.logLevel]; ok {
		zapLevel = l
	}
	atomicLevel := zap.NewAtomicLevelAt(zapLevel)

	cfg := zap.Config{
		Level:            atomicLevel,
		Development:      true, //TODO do something about this?
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, atomicLevel, err
	}

	return logger, atomicLevel, nil
}

func init() {
	levelToZapLevelMap = map[Level]zapcore.Level{
		Debug:   zapcore.DebugLevel,
		Info:    zapcore.InfoLevel,
		Warning: zapcore.WarnLevel,
		Error:   zapcore.ErrorLevel,
		Fatal:   zapcore.FatalLevel,
	}
}

func levelToZapFlatLogFunc(level Level, logger *zap.SugaredLogger) logFunc {
	switch level {
	case Debug:
		return logger.Debugf
	case Info:
		return logger.Infof
	case Warning:
		return logger.Warnf
	case Error:
		return logger.Errorf
	case Fatal:
		return logger.Fatalf
	default:
		return nil
	}
}
