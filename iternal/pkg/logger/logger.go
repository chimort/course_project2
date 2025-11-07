package logger

import (
	"log/slog"
	"os"
)

func NewLogger(serviceName string, level slog.Level) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	logger := slog.New(handler).With("service", serviceName)
	return logger
}
