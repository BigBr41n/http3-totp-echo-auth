package logger

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogConfig holds all logger configuration
type LogConfig struct {
	Level        string
	FilePath     string
	MaxSize      int
	MaxBackups   int
	MaxAge       int
	Compress     bool
	ReportCaller bool
	Format       string // "json" or "text"
}

// DefaultConfig returns sensible default configuration
func DefaultConfig() LogConfig {
	return LogConfig{
		Level:        "info",
		FilePath:     "logs/app.log",
		MaxSize:      10, // 10 MB
		MaxBackups:   5,  // 5 backups
		MaxAge:       30, // 30 days
		Compress:     true,
		ReportCaller: true,
		Format:       "json",
	}
}

var (
	instance *logrus.Logger
	once     sync.Once
	mu       sync.Mutex
)

// GetLogger returns the singleton logger instance
func GetLogger() *logrus.Logger {
	if instance == nil {
		// If not initialized yet, create with default config
		_ = InitLogger(DefaultConfig())
	}
	return instance
}

// InitLogger initializes the logger with the given configuration
// This should be called early in your application startup
func InitLogger(config LogConfig) error {
	mu.Lock()
	defer mu.Unlock()

	// Ensure instance is only created once
	var err error
	once.Do(func() {
		instance, err = configureLogger(config)
	})

	// Allow reconfiguration if instance already exists
	if err == nil && instance != nil {
		instance.Info("Logger initialized successfully")
	}
	return err
}

// ResetLogger allows reconfiguration of the logger
// Useful primarily for testing or configuration changes
func ResetLogger(config LogConfig) error {
	mu.Lock()
	defer mu.Unlock()

	// Reset the once flag (for testing purposes)
	once = sync.Once{}

	var err error
	instance, err = configureLogger(config)
	if err == nil && instance != nil {
		instance.Info("Logger reconfigured successfully")
	}
	return err
}

// Internal function to configure the logger
func configureLogger(config LogConfig) (*logrus.Logger, error) {
	log := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		return nil, err
	}
	log.SetLevel(level)

	// Configure format
	if config.Format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{})
	} else {
		log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	}

	// Set caller reporting
	log.SetReportCaller(config.ReportCaller)

	// Set up outputs
	var outputs []io.Writer
	outputs = append(outputs, os.Stdout)

	// Configure file output if path provided
	if config.FilePath != "" {
		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(config.FilePath), 0755); err != nil {
			return nil, err
		}

		fileLogger := &lumberjack.Logger{
			Filename:   config.FilePath,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		}
		outputs = append(outputs, fileLogger)
	}

	// Set output to multi-writer
	log.SetOutput(io.MultiWriter(outputs...))

	return log, nil
}

// Convenience methods that delegate to the singleton instance
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

func WithField(key string, value interface{}) *logrus.Entry {
	return GetLogger().WithField(key, value)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return GetLogger().WithFields(fields)
}
