package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// logger is global variable that stores the zap.Logger reference
var logger *zap.Logger

func Init(env string) error {
	var cfg zap.Config

	if env == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.EncodeTime = customTimeEncoder // set custom time format

	var err error
	logger, err = cfg.Build() // build the config
	if err != nil {
		return err
	}

	zap.ReplaceGlobals(logger) // replace global logger with new one (`logger`)
	return nil
}

// customTimeEncoder formats time in zap encoder in 'DD-MM-YYYY HH:MM:SS' format
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("02-01-2006 15:04:05")) // DD-MM-YYYY HH:MM:SS
}

// Below provided convenient functions to log messages of different level.

// Info logs an informational message
func Info(msg string, fields ...zap.Field) {
	zap.L().Info(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	zap.L().Error(msg, fields...)
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	zap.L().Debug(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	zap.L().Warn(msg, fields...)
}

// Fatal logs a fatal message and exits program running
func Fatal(msg string, fields ...zap.Field) {
	zap.L().Fatal(msg, fields...)
}

// Sync flushes any buffered log entries
func Sync() {
	if logger != nil {
		_ = logger.Sync()
	}
}
