// Package models contains the response models for the API endpoints.
package models

type JSONResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"` // omitempty hides the field if it's nil
}
