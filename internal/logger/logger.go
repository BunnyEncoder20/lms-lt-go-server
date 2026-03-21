// package logger logs the events of the server, and also sends json struct when in prod
package logger

import (
	"log/slog"
	"os"
)

func NewLogger() *slog.Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug, // Log everything in dev
	}

	if env == "production" {
		// Production: JSON format , only Info and above
		opts.Level = slog.LevelInfo
	} else {
		// Local dev: colorized/formatted text output
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
