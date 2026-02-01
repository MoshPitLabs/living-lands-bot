package services

import (
	"log/slog"
	"os"
)

// getTestLogger returns a logger configured for tests.
func getTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}
