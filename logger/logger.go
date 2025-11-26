package logger

import (
	"log/slog"
	"os"
)

// Level represents log levels
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var (
	logger *slog.Logger
	level  Level = LevelInfo
)

// GetLevel returns the current logging level
func GetLevel() Level {
	return level
}

// Init initializes the logger with the specified level
func Init(l Level) {
	level = l
	opts := &slog.HandlerOptions{
		Level: toSlogLevel(l),
	}
	logger = slog.New(slog.NewTextHandler(os.Stderr, opts))
}

func toSlogLevel(l Level) slog.Level {
	switch l {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// ParseLevel parses a string level to Level
func ParseLevel(s string) Level {
	switch s {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	if logger == nil {
		Init(LevelInfo)
	}
	logger.Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...any) {
	if logger == nil {
		Init(LevelInfo)
	}
	logger.Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	if logger == nil {
		Init(LevelInfo)
	}
	logger.Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	if logger == nil {
		Init(LevelInfo)
	}
	logger.Error(msg, args...)
}

// GetLogger returns the underlying slog logger
func GetLogger() *slog.Logger {
	if logger == nil {
		Init(LevelInfo)
	}
	return logger
}
