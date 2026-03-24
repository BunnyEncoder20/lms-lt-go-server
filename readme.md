# Todo Targets

✅ Context Helpers

- Currently, you're manually extracting UserID and UserRole from the context with type assertions (e.g., r.Context().Value(models.UserIDKey).(string)).
- Recommendation: Add helper functions in internal/models/requests.go (or a new context.go) to handle this safely.

```go
func GetUserID(ctx context.Context) (string, bool) {
id, ok := ctx.Value(UserIDKey).(string)
return id, ok
}
```

✅ Structured Logging (slog)

- You're currently using the standard log package. Since Go 1.21, log/slog is the standard for structured logging.
- Recommendation: Initialize a global or injected slog.Logger. This will allow you to log with context (e.g., logger.Info("user logged in", "user_id", id)) which is invaluable for debugging production systems.

1. Centralized Configuration

- You're using os.Getenv in multiple places (routes.go, database.go, auth/handler.go). Recommendation: Create an internal/config package that loads all environment variables into a single struct at startup. This provides a single source of truth and allows you to fail-fast if a required variable is missing.

1. Request Validation

- As you add modules with complex POST/PUT bodies, you'll need a consistent way to validate input (e.g., "email is required", "password must be 8+ chars").
- Recommendation: Consider adding a Validate() error method to your request structs or using a library like github.com/go-playground/validator.

1. Global Middleware (Recovery & Logging)

- It's good practice to have:
  - Recovery Middleware: To catch panics so the whole server doesn't crash on a nil-pointer dereference.
  - Request Logger: To log every incoming request's method, path, status code, and duration.

1. Centralized Error Handling

- Instead of calling models.WriteJSON with http.StatusInternalServerError everywhere, you could define a custom AppError type and a helper that maps internal errors to HTTP responses consistently.
