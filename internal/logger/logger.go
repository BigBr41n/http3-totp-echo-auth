package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.Logger

// Init initializes the global logger with file rotation and JSON formatting.
func Init() {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

	// Configure file rotation using lumberjack
	rotatingFile := &lumberjack.Logger{
		Filename:   "./logs/server.log",
		MaxSize:    10, // megabytes
		MaxBackups: 7,
		MaxAge:     30, // days
		Compress:   true,
	}

	writer := zapcore.NewMultiWriteSyncer(
		zapcore.AddSync(os.Stdout),
		zapcore.AddSync(rotatingFile),
	)

	core := zapcore.NewCore(jsonEncoder, writer, zap.DebugLevel)

	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

// Info logs an informational message.
func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

// Error logs an error message.
func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

// Debug logs a debug message.
func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

// Warn logs a warning message.
func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

// Fatal errors
func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}

// Singelton pattern
func Logger() *zap.Logger {
	return logger
}
