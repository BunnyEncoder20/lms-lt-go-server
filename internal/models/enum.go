package models

// Role represents the role of a user in the system.
type Role string

const (
	RoleAdmin          Role = "ADMIN"
	RoleManager        Role = "MANAGER"
	RoleEmployee       Role = "EMPLOYEE"
	RoleCourseDirector Role = "COURSE_DIRECTOR"
)

// NominationStatus represents the status of a nomination in the system.
type NominationStatus string

const (
	NomPendingManagerAssignment NominationStatus = "PENDING_MANAGER_ASSIGNMENT"
	NomPendingEmployeeApproval   NominationStatus = "PENDING_EMPLOYEE_APPROVAL"
	NomEnrolled                  NominationStatus = "ENROLLED"
	NomPendingManagerApproval    NominationStatus = "PENDING_MANAGER_APPROVAL"
	NomDeclined                  NominationStatus = "DECLINED"
	NomRejected                  NominationStatus = "REJECTED"
	NomCompleted                 NominationStatus = "COMPLETED"
	NomAttended                  NominationStatus = "ATTENDED"
)

// IsValidNominationStatus Helper func in models package
func IsValidNominationStatus(status NominationStatus) bool {
	switch status {
	case NomPendingManagerAssignment,
		NomPendingEmployeeApproval,
		NomEnrolled,
		NomPendingManagerApproval,
		NomDeclined,
		NomRejected,
		NomCompleted,
		NomAttended:
		return true
	default:
		return false
	}
}

// CourseStatus represents the status of a course in the system.
type CourseStatus string

const (
	CourseDraft     CourseStatus = "DRAFT"
	CoursePublished CourseStatus = "PUBLISHED"
	CourseArchived  CourseStatus = "ARCHIVED"
)

// TrainingCategory represents the category of a course in the system.
type TrainingCategory string

const (
	TrainingTechnical  TrainingCategory = "TECHNICAL"
	TrainingITDigital  TrainingCategory = "IT_DIGITAL"
	TrainingQuality    TrainingCategory = "QUALITY"
	TrainingSafety     TrainingCategory = "SAFETY"
	TrainingBehavioral TrainingCategory = "BEHAVIORAL"
)

// LessonContentType represents the type of content in a lesson.
type LessonContentType string

const (
	LessonVideo        LessonContentType = "VIDEO"
	LessonAudio        LessonContentType = "AUDIO"
	LessonPdf          LessonContentType = "PDF"
	LessonImage        LessonContentType = "IMAGE"
	LessonRichText     LessonContentType = "RICH_TEXT"
	LessonPresentation LessonContentType = "PRESENTATION"
)

// CourseAssignmentStatus represents the status of a course assignment in the system.
type CourseAssignmentStatus string

const (
	AssignmentNotStarted CourseAssignmentStatus = "NOT_STARTED"
	AssignmentInProgress CourseAssignmentStatus = "IN_PROGRESS"
	AssignmentCompleted  CourseAssignmentStatus = "COMPLETED"
)

// CalendarPlanStatus enum represents the status of the planned events of calendar
type CalendarPlanStatus string

const (
	EventPlanned   CalendarPlanStatus = "PLANNED"
	EventFinalized CalendarPlanStatus = "FINALIZED"
	EventCancelled CalendarPlanStatus = "CANCELLED"
)

// TrainingFormat enum represents the format of the training
type TrainingFormat string

const (
	FormatInPerson TrainingFormat = "IN_PERSON"
	FormatVirtual  TrainingFormat = "VIRTUAL"
	FormatHybrid   TrainingFormat = "HYBRID"
)

// TrainingStatus enum represents the lifecycle status of a training
type TrainingStatus string

const (
	TrainingDraft     TrainingStatus = "DRAFT"
	TrainingScheduled TrainingStatus = "SCHEDULED"
	TrainingPublished TrainingStatus = "PUBLISHED"
)

// AttendanceRequestStatus enum represents the status of an attendance request
type AttendanceRequestStatus string

const (
	AttendanceSent      AttendanceRequestStatus = "SENT"
	AttendanceConfirmed AttendanceRequestStatus = "CONFIRMED"
	AttendanceExpired   AttendanceRequestStatus = "EXPIRED"
	AttendanceVoid      AttendanceRequestStatus = "VOID"
)

// AttendanceEntityType enum represents the entity type for attendance dispatches
type AttendanceEntityType string

const (
	EntityTraining AttendanceEntityType = "TRAINING"
	EntityCourse   AttendanceEntityType = "COURSE"
)

// DeliveryMode represents the mode of delivery for a training.
type DeliveryMode string

const (
	InPerson    DeliveryMode = "IN_PERSON"
	VirtualLink DeliveryMode = "VIRTUAL"
	Hybrid      DeliveryMode = "HYBRID"
	Elearning   DeliveryMode = "E-LEARNING"
)
