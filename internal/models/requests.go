package models

import (
	"github.com/google/uuid"
)

type LoginRequest struct {
	PesNumber string `json:"pes_number"`
	Password  string `json:"password"`
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

// Nomination Requests

type CreateNominationRequest struct {
	TrainingID string   `json:"training_id"`
	UserIDs    []string `json:"user_ids"`
}

type NominationResponseRequest struct {
	Status string `json:"status"` // ACCEPTED, DECLINED
}

type SelfNominationRequest struct {
	TrainingID string `json:"training_id"`
}

type UpdateNominationStatusRequest struct {
	Status NominationStatus `json:"status"`
}

type NominationFilters struct {
	Status     *string `json:"status,omitempty"`
	TrainingID *string `json:"training_id,omitempty"`
	UserID     *string `json:"user_id,omitempty"`
	ManagerID  *string `json:"manager_id,omitempty"`
	StartDate  *string `json:"start_date,omitempty"`
	EndDate    *string `json:"end_date,omitempty"`
}

// Course Requests

type CreateCourseRequest struct {
	Title              string   `json:"title"`
	Description        *string  `json:"description,omitempty"`
	CoverImageURL      *string  `json:"cover_image_url,omitempty"`
	Category           string   `json:"category"`
	EstimatedDuration  *int64   `json:"estimated_duration,omitempty"`
	LearningOutcomes   []string `json:"learning_outcomes"`
	IsStrictSequencing *bool    `json:"is_strict_sequencing,omitempty"`
}

type UpdateCourseRequestV2 struct {
	Title              *string  `json:"title,omitempty"`
	Description        *string  `json:"description,omitempty"`
	CoverImageURL      *string  `json:"cover_image_url,omitempty"`
	Category           *string  `json:"category,omitempty"`
	EstimatedDuration  *int64   `json:"estimated_duration,omitempty"`
	LearningOutcomes   []string `json:"learning_outcomes,omitempty"`
	IsStrictSequencing *bool    `json:"is_strict_sequencing,omitempty"`
}

type CreateCourseModuleRequest struct {
	Title         string  `json:"title"`
	Description   *string `json:"description,omitempty"`
	SequenceOrder *int64  `json:"sequence_order,omitempty"`
}

type UpdateCourseModuleRequest struct {
	Title         *string `json:"title,omitempty"`
	Description   *string `json:"description,omitempty"`
	SequenceOrder *int64  `json:"sequence_order,omitempty"`
}

type ReorderItemRequest struct {
	ID string `json:"id"`
}

type ReorderCourseModulesRequest struct {
	Modules []ReorderItemRequest `json:"modules"`
}

type CreateLessonRequest struct {
	Title           string  `json:"title"`
	ContentType     string  `json:"content_type"`
	AssetURL        *string `json:"asset_url,omitempty"`
	RichTextContent *string `json:"rich_text_content,omitempty"`
	DurationMinutes *int64  `json:"duration_minutes,omitempty"`
	SequenceOrder   *int64  `json:"sequence_order,omitempty"`
}

type UpdateLessonRequest struct {
	Title           *string `json:"title,omitempty"`
	ContentType     *string `json:"content_type,omitempty"`
	AssetURL        *string `json:"asset_url,omitempty"`
	RichTextContent *string `json:"rich_text_content,omitempty"`
	DurationMinutes *int64  `json:"duration_minutes,omitempty"`
	SequenceOrder   *int64  `json:"sequence_order,omitempty"`
}

type ReorderLessonsRequest struct {
	Lessons []ReorderItemRequest `json:"lessons"`
}

// Learning Requests

type BulkAssignRequest struct {
	CourseID string   `json:"course_id"`
	UserIDs  []string `json:"user_ids"`
	DueDate  *string  `json:"due_date,omitempty"`
}

type UpdateProgressRequest struct {
	IsCompleted          *bool  `json:"is_completed,omitempty"`
	LastPlaybackPosition *int64 `json:"last_playback_position,omitempty"`
}

type HeartbeatProgressRequest struct {
	LastPlaybackPosition int64 `json:"last_playback_position"`
}

type CompleteLessonRequest struct {
	WatchedPercent   *float64 `json:"watched_percent,omitempty"`
	ListenedPercent  *float64 `json:"listened_percent,omitempty"`
	ScrolledPercent  *float64 `json:"scrolled_percent,omitempty"`
	ViewedSeconds    *int64   `json:"viewed_seconds,omitempty"`
	ReachedLastPage  *bool    `json:"reached_last_page,omitempty"`
	OpenedInLightbox *bool    `json:"opened_in_lightbox,omitempty"`
}
