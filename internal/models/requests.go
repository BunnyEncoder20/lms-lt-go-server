package models

import (
	"github.com/google/uuid"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// User Requests

type FindUsersRequest struct {
	Role    *string `json:"role"`
	Cluster *string `json:"cluster"`
	IsID    *string `json:"is_id"` // Search by supervisor ID
}

type CreateUserRequest struct {
	PesNumber    string  `json:"pes_number"`
	Password     string  `json:"password"`
	FirstName    string  `json:"first_name"`
	LastName     string  `json:"last_name"`
	Email        string  `json:"email"`
	Role         string  `json:"role"`
	Cluster      *string `json:"cluster"`
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
	IsID         *string `json:"is_id"`
	NsID         *string `json:"ns_id"`
	DhID         *string `json:"dh_id"`
}

type UpdateUserRequest struct {
	FirstName    *string `json:"first_name"`
	LastName     *string `json:"last_name"`
	Title        *string `json:"title"`
	Department   *string `json:"department"`
	BaseLocation *string `json:"base_location"`
}

type UserStatusRequest struct {
	IsActive bool `json:"is_active"`
}

// UserID is a helper for when we need to pass just the ID
type UserID struct {
	ID uuid.UUID
}
