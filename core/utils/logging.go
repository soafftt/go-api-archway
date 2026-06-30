package utils

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

const (
	LogFormatText = "text"
	LogFormatJSON = "json"
)

type LoggerConfig struct {
	Level     slog.Level
	Format    string
	AddSource bool
	Writer    io.Writer
}

func init() {
	ConfigureLogger(LoggerConfig{})
}

func ConfigureLogger(config LoggerConfig) *slog.Logger {
	logger := NewLogger(config)
	slog.SetDefault(logger)

	return logger
}

func NewLogger(config LoggerConfig) *slog.Logger {
	writer := config.Writer
	if writer == nil {
		writer = os.Stderr
	}

	options := &slog.HandlerOptions{
		AddSource: config.AddSource,
		Level:     config.Level,
	}

	switch normalizeLogFormat(config.Format) {
	case LogFormatJSON:
		return slog.New(slog.NewJSONHandler(writer, options))
	default:
		return slog.New(slog.NewTextHandler(writer, options))
	}
}

func ParseLogLevel(level string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("unsupported log level: %s", level)
	}
}

func Logger() *slog.Logger {
	return slog.Default()
}

func With(args ...any) *slog.Logger {
	return slog.Default().With(args...)
}

func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}

func normalizeLogFormat(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case LogFormatJSON:
		return LogFormatJSON
	default:
		return LogFormatText
	}
}
