// Package logger logs the events of the server, and also sends json struct when in prod
package logger

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

type levelSourceHandler struct {
	slog.Handler
}

func (h *levelSourceHandler) Handle(ctx context.Context, r slog.Record) error {
	// Only include source for Debug, Warn, or Error (Level < Info or Level > Info)
	if r.Level == slog.LevelInfo {
		r.PC = 0
	}
	return h.Handler.Handle(ctx, r)
}

func NewLogger() *slog.Logger {
	var handler slog.Handler

	if os.Getenv("APP_ENV") == "production" {
		// Production: JSON format, only Info and above
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		// Local dev: colorized/formatted text output using tint
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.RFC3339,
			AddSource:  true,
		})
		// Wrap with our custom handler to dynamically hide source for Info logs
		handler = &levelSourceHandler{Handler: handler}
	}

	return slog.New(handler)
}
