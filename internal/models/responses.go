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

type UserResponse struct {
	ID           string  `json:"id"`
	PesNumber    string  `json:"pes_number"`
	FirstName    string  `json:"first_name"`
	LastName     string  `json:"last_name"`
	Email        string  `json:"email"`
	Role         Role    `json:"role"`
	Cluster      *string `json:"cluster,omitempty"`
	Title        string  `json:"title"`
	Gender       string  `json:"gender"`
	Band         string  `json:"band"`
	Grade        string  `json:"grade"`
	Ic           string  `json:"ic"`
	Sbg          string  `json:"sbg"`
	Bu           string  `json:"bu"`
	Segment      string  `json:"segment"`
	Department   string  `json:"department"`
	BaseLocation string  `json:"base_location"`
	IsID         *string `json:"is_id,omitempty"`
	NsID         *string `json:"ns_id,omitempty"`
	DhID         *string `json:"dh_id,omitempty"`
}
