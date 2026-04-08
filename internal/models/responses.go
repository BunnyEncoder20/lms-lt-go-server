// Package models contains the models for the API endpoints for proper structure of certain types.
package models

import (
	"encoding/json"
	"log"
	"net/http"
)

type JSONResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
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

type TrainingResponse struct {
	ID                   string           `json:"id"`
	Title                string           `json:"title"`
	Description          *string          `json:"description,omitempty"`
	Category             TrainingCategory `json:"category"`
	StartDate            string           `json:"start_date"`
	EndDate              string           `json:"end_date"`
	Location             *string          `json:"location,omitempty"`
	VirtualLink          *string          `json:"virtual_link,omitempty"`
	PreReadURI           *string          `json:"pre_read_uri,omitempty"`
	CreatedByID          string           `json:"created_by_id"`
	DeadlineDays         int64            `json:"deadline_days"`
	HrProgramID          string           `json:"hr_program_id"`
	MappedCategory       string           `json:"mapped_category"`
	ModeOfDelivery       DeliveryMode     `json:"mode_of_delivery"`
	InstructorName       string           `json:"instructor_name"`
	InstitutePartnerName *string          `json:"institute_partner_name,omitempty"`
	ProcessOwnerName     *string          `json:"process_owner_name,omitempty"`
	ProcessOwnerEmail    *string          `json:"process_owner_email,omitempty"`
	DurationManhours     *float64         `json:"duration_manhours,omitempty"`
	TrainingMandays      *float64         `json:"training_mandays,omitempty"`
	FacilityID           string           `json:"facility_id"`
	IsActive             bool             `json:"is_active"`
	CreatedAt            string           `json:"created_at"`
	UpdatedAt            string           `json:"updated_at"`
}

type NominationResponse struct {
	ID                   string           `json:"id"`
	Status               NominationStatus `json:"status"`
	UserID               string           `json:"user_id"`
	TrainingID           string           `json:"training_id"`
	NominatedByID        string           `json:"nominated_by_id"`
	HrCompletionStatus   *string          `json:"hr_completion_status,omitempty"`
	ProfFees             *float64         `json:"prof_fees,omitempty"`
	VenueCost            *float64         `json:"venue_cost,omitempty"`
	OtherCost            *float64         `json:"other_cost,omitempty"`
	NonTemsTravel        *float64         `json:"non_tems_travel,omitempty"`
	NonTemsAccommodation *float64         `json:"non_tems_accommodation,omitempty"`
	TotalCost            *float64         `json:"total_cost,omitempty"`
	CreatedAt            string           `json:"created_at"`
	UpdatedAt            string           `json:"updated_at"`
}

type NominationWithDetailsResponse struct {
	ID                   string           `json:"id"`
	Status               NominationStatus `json:"status"`
	User                 UserResponse     `json:"user"`
	Training             TrainingResponse `json:"training"`
	NominatedByID        string           `json:"nominated_by_id"`
	HrCompletionStatus   *string          `json:"hr_completion_status,omitempty"`
	ProfFees             *float64         `json:"prof_fees,omitempty"`
	VenueCost            *float64         `json:"venue_cost,omitempty"`
	OtherCost            *float64         `json:"other_cost,omitempty"`
	NonTemsTravel        *float64         `json:"non_tems_travel,omitempty"`
	NonTemsAccommodation *float64         `json:"non_tems_accommodation,omitempty"`
	TotalCost            *float64         `json:"total_cost,omitempty"`
	CreatedAt            string           `json:"created_at"`
	UpdatedAt            string           `json:"updated_at"`
}

type ManagerDashboardResponse struct {
	TotalNominations     int64 `json:"total_nominations"`
	PendingNominations   int64 `json:"pending_nominations"`
	ApprovedNominations  int64 `json:"approved_nominations"`
	CompletedNominations int64 `json:"completed_nominations"`
	TeamSize             int64 `json:"team_size"`
}

type EmployeeDashboardResponse struct {
	TotalNominations     int64 `json:"total_nominations"`
	PendingNominations   int64 `json:"pending_nominations"`
	ApprovedNominations  int64 `json:"approved_nominations"`
	CompletedNominations int64 `json:"completed_nominations"`
	AvailableCourses     int64 `json:"available_courses"`
}

type AdminKpisResponse struct {
	TotalTrainings    int64   `json:"total_trainings"`
	TotalParticipants int64   `json:"total_participants"`
	CompletedCount    int64   `json:"completed_count"`
	EnrolledCount     int64   `json:"enrolled_count"`
	TotalManDays      float64 `json:"total_mandays"`
}

type MonthlyStatsResponse struct {
	MonthKey     string `json:"month_key"`
	MonthLabel   string `json:"month_label"`
	Participants int64  `json:"participants"`
	Trainings    int64  `json:"trainings"`
}

type CategoryDistributionResponse struct {
	Name  TrainingCategory `json:"training_name"`
	Value int64            `json:"value"`
}

type ClusterStatsResponse struct {
	Cluster        string `json:"cluster"`
	TotalEmployees int64  `json:"total_employees"`
	Trained        int64  `json:"trained"`
	Untrained      int64  `json:"untrained"`
}
