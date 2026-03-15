// Package models contains the models for the API endpoints for proper structure of certain types.
package models

type JSONResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"` // omitempty hides the field if it's nil
}
