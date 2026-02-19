package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func New(service, version, env string, debug bool) (*Logger, error) {
	var config zap.Config

	if debug || env == "dev" {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}

	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.StacktraceKey = "stacktrace"

	baseLogger, err := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}

	logger := baseLogger.With(
		zap.String("service", service),
		zap.String("version", version),
		zap.String("env", env),
	)

	return &Logger{Logger: logger}, nil
}

func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{Logger: l.Logger.With(zap.String("request_id", requestID))}
}

func (l *Logger) WithUserID(userID int64) *Logger {
	return &Logger{Logger: l.Logger.With(zap.Int64("user_id", userID))}
}

func (l *Logger) WithError(err error) *Logger {
	return &Logger{Logger: l.Logger.With(zap.Error(err))}
}

// Sync flushes any buffered Log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}
