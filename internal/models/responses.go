// Package models contains the models for the API endpoints for proper structure of certain types.
package models

type JSONResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type UserResponse struct {
	ID               string  `json:"id"`
	PesNumber        string  `json:"pes_number"`
	FirstName        string  `json:"first_name"`
	LastName         string  `json:"last_name"`
	FullName         *string `json:"full_name,omitempty"`
	Email            string  `json:"email"`
	Role             Role    `json:"role"`
	Cluster          *string `json:"cluster,omitempty"`
	Location         *string `json:"location,omitempty"`
	Title            *string `json:"title"`
	Gender           *string `json:"gender"`
	Band             *string `json:"band"`
	Grade            *string `json:"grade"`
	EmploymentStatus *string `json:"employment_status,omitempty"`
	IsPsn            *string `json:"is_psn,omitempty"`
	IsName           *string `json:"is_name,omitempty"`
	NsPsn            *string `json:"ns_psn,omitempty"`
	NsName           *string `json:"ns_name,omitempty"`
	DhPsn            *string `json:"dh_psn,omitempty"`
	DhName           *string `json:"dh_name,omitempty"`
	Ic               *string `json:"ic,omitempty"`
	Sbg              *string `json:"sbg,omitempty"`
	Bu               *string `json:"bu,omitempty"`
	Segment          *string `json:"segment,omitempty"`
	Department       *string `json:"department,omitempty"`
	BaseLocation     *string `json:"base_location,omitempty"`
	ManagerID        *string `json:"manager_id,omitempty"`
	SkipManagerID    *string `json:"skip_manager_id,omitempty"`
	IsID             *string `json:"is_id,omitempty"`
	NsID             *string `json:"ns_id,omitempty"`
	DhID             *string `json:"dh_id,omitempty"`
}

type TrainingResponse struct {
	ID                   string           `json:"id"`
	Title                string           `json:"title"`
	Description          *string          `json:"description,omitempty"`
	Category             TrainingCategory `json:"category"`
	InstructorName       *string          `json:"instructor_name,omitempty"`
	LearningOutcomes     *string          `json:"learning_outcomes,omitempty"`
	MonthTag             *string          `json:"month_tag,omitempty"`
	StartDate            string           `json:"start_date"`
	EndDate              string           `json:"end_date"`
	StartTime            *string          `json:"start_time,omitempty"`
	EndTime              *string          `json:"end_time,omitempty"`
	Timezone             *string          `json:"timezone,omitempty"`
	Format               *string          `json:"format,omitempty"`
	RegistrationDeadline *string          `json:"registration_deadline,omitempty"`
	MaxCapacity          *int64           `json:"max_capacity,omitempty"`
	TargetClusters       *string          `json:"target_clusters,omitempty"`
	PrerequisitesUrl     *string          `json:"prerequisites_url,omitempty"`
	VenueCost            int64            `json:"venue_cost"`
	ProfessionalFees     int64            `json:"professional_fees"`
	StationaryCost       int64            `json:"stationary_cost"`
	Status               string           `json:"status"`
	Location             *string          `json:"location,omitempty"`
	VirtualLink          *string          `json:"virtual_link,omitempty"`
	PreReadUrl           *string          `json:"pre_read_url,omitempty"`
	DeadlineDays         int64            `json:"deadline_days"`
	IsActive             bool             `json:"is_active"`
	CreatedAt            string           `json:"created_at"`
	UpdatedAt            string           `json:"updated_at"`
	CreatedByID          string           `json:"created_by_id"`
	HrProgramID          *string          `json:"hr_program_id,omitempty"`
	MappedCategory       *string          `json:"mapped_category,omitempty"`
	ModeOfDelivery       DeliveryMode     `json:"mode_of_delivery"`
	InstitutePartnerName *string          `json:"institute_partner_name,omitempty"`
	ProcessOwnerName     *string          `json:"process_owner_name,omitempty"`
	ProcessOwnerEmail    *string          `json:"process_owner_email,omitempty"`
	DurationManhours     *float64         `json:"duration_manhours,omitempty"`
	TrainingMandays      *float64         `json:"training_mandays,omitempty"`
	FacilityID           *string          `json:"facility_id,omitempty"`
}

type NominationResponse struct {
	ID                   string           `json:"id"`
	Status               NominationStatus `json:"status"`
	UserID               string           `json:"user_id"`
	TrainingID           string           `json:"training_id"`
	NominatedByID        string           `json:"nominated_by_id"`
	HrCompletionStatus   *string          `json:"hr_completion_status,omitempty"`
	ProfFees             *int64           `json:"prof_fees,omitempty"`
	VenueCost            *int64           `json:"venue_cost,omitempty"`
	OtherCost            *int64           `json:"other_cost,omitempty"`
	NonTemsTravel        *int64           `json:"non_tems_travel,omitempty"`
	NonTemsAccommodation *int64           `json:"non_tems_accommodation,omitempty"`
	TotalCost            *int64           `json:"total_cost,omitempty"`
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
	ProfFees             *int64           `json:"prof_fees,omitempty"`
	VenueCost            *int64           `json:"venue_cost,omitempty"`
	OtherCost            *int64           `json:"other_cost,omitempty"`
	NonTemsTravel        *int64           `json:"non_tems_travel,omitempty"`
	NonTemsAccommodation *int64           `json:"non_tems_accommodation,omitempty"`
	TotalCost            *int64           `json:"total_cost,omitempty"`
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

type ImportResponse struct {
	TotalRows       int64 `json:"total_rows"`
	Imported        int64 `json:"imported"`
	UniqueTrainings int64 `json:"unique_trainings"`
	MonthCoverage   int64 `json:"month_coverage"`
}

type CourseAuthorResponse struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type CourseCountResponse struct {
	Modules     int64 `json:"modules"`
	Assignments int64 `json:"assignments"`
}

type CourseListResponse struct {
	ID                 string                `json:"id"`
	Title              string                `json:"title"`
	Description        *string               `json:"description,omitempty"`
	CoverImageURL      *string               `json:"cover_image_url,omitempty"`
	Status             CourseStatus          `json:"status"`
	Category           TrainingCategory      `json:"category"`
	EstimatedDuration  *int64                `json:"estimated_duration,omitempty"`
	LearningOutcomes   []string              `json:"learning_outcomes"`
	IsStrictSequencing bool                  `json:"is_strict_sequencing"`
	Version            int64                 `json:"version"`
	PublishedAt        *string               `json:"published_at,omitempty"`
	CreatedAt          string                `json:"created_at"`
	UpdatedAt          string                `json:"updated_at"`
	Author             *CourseAuthorResponse `json:"author,omitempty"`
	Count              CourseCountResponse   `json:"_count"`
}

type LessonResponse struct {
	ID              string            `json:"id"`
	Title           string            `json:"title"`
	ContentType     LessonContentType `json:"content_type"`
	AssetURL        *string           `json:"asset_url,omitempty"`
	RichTextContent *string           `json:"rich_text_content,omitempty"`
	DurationMinutes *int64            `json:"duration_minutes,omitempty"`
	SequenceOrder   int64             `json:"sequence_order"`
}

type CourseModuleResponse struct {
	ID            string           `json:"id"`
	Title         string           `json:"title"`
	Description   *string          `json:"description,omitempty"`
	SequenceOrder int64            `json:"sequence_order"`
	Lessons       []LessonResponse `json:"lessons"`
}

type CourseDetailResponse struct {
	ID                 string                 `json:"id"`
	Title              string                 `json:"title"`
	Description        *string                `json:"description,omitempty"`
	CoverImageURL      *string                `json:"cover_image_url,omitempty"`
	Status             CourseStatus           `json:"status"`
	Category           TrainingCategory       `json:"category"`
	EstimatedDuration  *int64                 `json:"estimated_duration,omitempty"`
	LearningOutcomes   []string               `json:"learning_outcomes"`
	IsStrictSequencing bool                   `json:"is_strict_sequencing"`
	Version            int64                  `json:"version"`
	PublishedAt        *string                `json:"published_at,omitempty"`
	CreatedAt          string                 `json:"created_at"`
	UpdatedAt          string                 `json:"updated_at"`
	AuthorID           *string                `json:"author_id,omitempty"`
	Author             *CourseAuthorResponse  `json:"author,omitempty"`
	Modules            []CourseModuleResponse `json:"modules"`
	Count              CourseCountResponse    `json:"_count"`
}

type CourseDashboardStatsResponse struct {
	TotalCourses         int64 `json:"total_courses"`
	Published            int64 `json:"published"`
	Draft                int64 `json:"draft"`
	Archived             int64 `json:"archived"`
	TotalLessons         int64 `json:"total_lessons"`
	TotalAssignments     int64 `json:"total_assignments"`
	CompletedAssignments int64 `json:"completed_assignments"`
	CompletionRate       int64 `json:"completion_rate"`
}
