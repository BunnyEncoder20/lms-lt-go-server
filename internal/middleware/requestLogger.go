package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// RequestLogger is a factory function using Closure Pattern to accept args which are needed by the middleware, and returns the actual middleware function that will be used in the routes.
func RequestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			latency := time.Since(start)

			// Log the req details
			log.Info("HTTP Request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Duration("latency", latency),
			)
		})
	}
}
