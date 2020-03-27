package cloudlogging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	levelToZapLevelMap map[Level]zapcore.Level
)

// Zap SugaredLogger logger function
type logFunc func(string, ...interface{})

func createConfig(opts options) *zap.Config {
	zapLevel := zapcore.InfoLevel
	if l, ok := levelToZapLevelMap[opts.logLevel]; ok {
		zapLevel = l
	}
	atomicLevel := zap.NewAtomicLevelAt(zapLevel)

	outputPaths := opts.outputPaths
	if len(outputPaths) == 0 {
		outputPaths = []string{"stdout"}
	}

	errorOutputPaths := opts.errorOutputPaths
	if len(errorOutputPaths) == 0 {
		errorOutputPaths = []string{"stderr"}
	}

	encoding := "console"
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	disableCaller := false

	// Process output hints
	for _, h := range opts.outputHints {
		if h == JSONFormat {
			encoding = "json"
			disableCaller = true
			encoderConfig = zapcore.EncoderConfig{
				// Keys can be anything except the empty string.
				TimeKey:        "timestamp",
				LevelKey:       "level",
				MessageKey:     "message",
				StacktraceKey:  "stacktrace",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.CapitalLevelEncoder,
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeDuration: zapcore.StringDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			}
		}
	}

	cfg := &zap.Config{
		Level:            atomicLevel,
		Development:      true,
		DisableCaller:    disableCaller,
		Encoding:         encoding,
		EncoderConfig:    encoderConfig,
		OutputPaths:      outputPaths,
		ErrorOutputPaths: errorOutputPaths,
	}

	return cfg
}

// createZapLogger creates a new Zap logger
func createZapLogger(opts options) (*zap.Logger, *zap.Config, error) {
	// We use the config specified on options if the API user defined one.
	// If not, we're creating one based on the OutputHints.
	cfg := opts.zapConfig
	if cfg == nil {
		cfg = createConfig(opts)
	}

	logger, err := cfg.Build()

	if err != nil {
		return nil, cfg, err
	}

	return logger, cfg, nil
}

func setZapLogLevel(zapConfig *zap.Config, logLevel Level) {
	zapLevel := zapcore.InfoLevel
	if l, ok := levelToZapLevelMap[logLevel]; ok {
		zapLevel = l
	}
	zapConfig.Level.SetLevel(zapLevel)
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

func levelToZapStructuredLogFunc(level Level,
	logger *zap.SugaredLogger) logFunc {
	switch level {
	case Debug:
		return logger.Debugw
	case Info:
		return logger.Infow
	case Warning:
		return logger.Warnw
	case Error:
		return logger.Errorw
	case Fatal:
		return logger.Fatalw
	default:
		return nil
	}
}
