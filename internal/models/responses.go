// Package models contains the models for the API endpoints for proper structure of certain types.
package models

import (
	"encoding/json"
	"log"
	"net/http"
)

type JSONResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"` // omitempty hides the field if it's nil
}

// WriteJSON Helper func to write json responses
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// NewEncoder(http.ResponseWriter).Encode(resp) streams the data (slightly faster for larger responses and also memeory efficient cause it doesn't hold the string in a var)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("error encoding json: %v", err)
	}
}
