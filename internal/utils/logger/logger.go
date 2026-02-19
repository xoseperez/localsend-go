// logger/logger.go
package logger

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

// LogConfig defines logging configuration options
type LogConfig struct {
	Level        logrus.Level
	Output       io.Writer
	Formatter    logrus.Formatter
	ReportCaller bool
}

var (
	logger *Logger
	once   sync.Once

	// ANSI color codes
	green = "\033[32m"
	red   = "\033[31m"
	reset = "\033[0m"
)

// DefaultConfig returns the default configuration
func DefaultConfig() LogConfig {
	return LogConfig{
		Level:  logrus.InfoLevel,
		Output: os.Stdout,
		Formatter: &logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
		},
		ReportCaller: false,
	}
}

// InitLogger initializes the Logger
func InitLogger(config ...LogConfig) {
	once.Do(func() {
		cfg := DefaultConfig()
		if len(config) > 0 {
			cfg = config[0]
		}

		log := logrus.New()
		log.SetOutput(cfg.Output)
		log.SetFormatter(cfg.Formatter)
		log.SetLevel(cfg.Level)
		log.SetReportCaller(cfg.ReportCaller)

		logger = &Logger{log}
	})
}

// checkLogger ensures the logger is initialized
func checkLogger() {
	if logger == nil {
		InitLogger() // Initialize with default config
	}
}

// GetLogger returns the underlying Logger instance
func GetLogger() *Logger {
	checkLogger()
	return logger
}

// Success prints a message with a green [Success] tag
func Success(args ...interface{}) {
	checkLogger()
	logger.Infof("%s[Success]%s %s", green, reset, fmt.Sprint(args...))
}

// Successf prints a formatted message with a green [Success] tag
func Successf(format string, args ...interface{}) {
	checkLogger()
	logger.Infof("%s[Success]%s %s", green, reset, fmt.Sprintf(format, args...))
}

// Failed prints a message with a red [Failed] tag
func Failed(args ...interface{}) {
	checkLogger()
	logger.Errorf("%s[Failed]%s %s", red, reset, fmt.Sprint(args...))
}

// Failedf prints a formatted message with a red [Failed] tag
func Failedf(format string, args ...interface{}) {
	checkLogger()
	logger.Errorf("%s[Failed]%s %s", red, reset, fmt.Sprintf(format, args...))
}

func Debug(args ...interface{}) {
	checkLogger()
	logger.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	checkLogger()
	logger.Debugf(format, args...)
}

// Info prints an info-level log message
func Info(args ...interface{}) {
	checkLogger()
	logger.Info(args...)
}

// Infof prints a formatted info-level log message
func Infof(format string, args ...interface{}) {
	checkLogger()
	logger.Infof(format, args...)
}

// Warn prints a warn-level log message
func Warn(args ...interface{}) {
	checkLogger()
	logger.Warn(args...)
}

// Warnf prints a formatted warn-level log message
func Warnf(format string, args ...interface{}) {
	checkLogger()
	logger.Warnf(format, args...)
}

// Error prints an error-level log message
func Error(args ...interface{}) {
	checkLogger()
	logger.Error(args...)
}

// Errorf prints a formatted error-level log message
func Errorf(format string, args ...interface{}) {
	checkLogger()
	logger.Errorf(format, args...)
}

// WithFields supports structured logging
func WithFields(fields logrus.Fields) *Logger {
	checkLogger()
	return &Logger{logger.WithFields(fields).Logger}
}
