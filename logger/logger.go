package logger

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func InitLogger(level string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
}

func Debug(msg string, args ...interface{}) {
	Logger.Debug(msg, args...)
}

func Info(msg string, args ...interface{}) {
	Logger.Info(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	Logger.Warn(msg, args...)
}

func Error(msg string, args ...interface{}) {
	Logger.Error(msg, args...)
}
