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

// Training Requests

type CreateTrainingRequest struct {
	Title                string   `json:"title"`
	Description          *string  `json:"description,omitempty"`
	Category             string   `json:"category"`
	StartDate            string   `json:"start_date"`
	EndDate              string   `json:"end_date"`
	Location             *string  `json:"location,omitempty"`
	VirtualLink          *string  `json:"virtual_link,omitempty"`
	PreReadURI           *string  `json:"pre_read_uri,omitempty"`
	DeadlineDays         int64    `json:"deadline_days"`
	HrProgramID          string   `json:"hr_program_id"`
	MappedCategory       string   `json:"mapped_category"`
	ModeOfDelivery       string   `json:"mode_of_delivery"`
	InstructorName       string   `json:"instructor_name"`
	InstitutePartnerName *string  `json:"institute_partner_name,omitempty"`
	ProcessOwnerName     *string  `json:"process_owner_name,omitempty"`
	ProcessOwnerEmail    *string  `json:"process_owner_email,omitempty"`
	DurationManhours     *float64 `json:"duration_manhours,omitempty"`
	TrainingMandays      *float64 `json:"training_mandays,omitempty"`
	FacilityID           string   `json:"facility_id"`
}

type UpdateTrainingRequest struct {
	Title                *string  `json:"title,omitempty"`
	Description          *string  `json:"description,omitempty"`
	Category             *string  `json:"category,omitempty"`
	StartDate            *string  `json:"start_date,omitempty"`
	EndDate              *string  `json:"end_date,omitempty"`
	Location             *string  `json:"location,omitempty"`
	VirtualLink          *string  `json:"virtual_link,omitempty"`
	PreReadURI           *string  `json:"pre_read_uri,omitempty"`
	DeadlineDays         *int64   `json:"deadline_days,omitempty"`
	MappedCategory       *string  `json:"mapped_category,omitempty"`
	ModeOfDelivery       *string  `json:"mode_of_delivery,omitempty"`
	InstructorName       *string  `json:"instructor_name,omitempty"`
	InstitutePartnerName *string  `json:"institute_partner_name,omitempty"`
	ProcessOwnerName     *string  `json:"process_owner_name,omitempty"`
	ProcessOwnerEmail    *string  `json:"process_owner_email,omitempty"`
	DurationManhours     *float64 `json:"duration_manhours,omitempty"`
	TrainingMandays      *float64 `json:"training_mandays,omitempty"`
	IsActive             *bool    `json:"is_active,omitempty"`
}
