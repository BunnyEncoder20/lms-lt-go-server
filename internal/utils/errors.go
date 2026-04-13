// Package utils provides utility functions for error handling and response writing
package utils

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"go-server/internal/models"
)

// HandleError inspects the error, logs it appropriately, and sends the correct HTTP response.
func HandleError(w http.ResponseWriter, r *http.Request, log *slog.Logger, err error) {
	// 1. Is it a "Not Found" error? (e.g., sqlc couldn't find the row)
	if errors.Is(err, sql.ErrNoRows) {
		log.Warn("resource not found", "error", err, "path", r.URL.Path)
		WriteJSON(w, http.StatusNotFound, models.JSONResponse{
			Success: false,
			Message: "the requested resource was not found",
		})
		return
	}

	// 2. Is it a JSON parsing error? (The client sent bad JSON)
	var syntaxErr *json.SyntaxError
	var unmarshalErr *json.UnmarshalTypeError
	if errors.As(err, &syntaxErr) || errors.As(err, &unmarshalErr) {
		log.Warn("malformed json payload", "error", err, "path", r.URL.Path)
		WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "invalid json payload",
		})
		return
	}

	// 3. Add custom business logic errors here
	if err.Error() == "unauthorized" {
		log.Warn("unauthorized access attempt", "error", err, "path", r.URL.Path)
		WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
			Success: false,
			Message: "unauthorized",
		})
		return
	}

	// 4. THE FALLBACK: If we don't recognize it, it's a true 500.
	// We log it as an ERROR (red text in terminal!) so we know we need to fix it.
	log.Error("internal server error", "error", err, "path", r.URL.Path)
	WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
		Success: false,
		Message: "an unexpected internal server error occurred",
	})
}
