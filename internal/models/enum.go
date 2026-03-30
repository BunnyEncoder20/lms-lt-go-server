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
	NomPending   NominationStatus = "PENDING_MANAGER"
	NomApproved  NominationStatus = "APPROVED"
	NomCompleted NominationStatus = "COMPLETED"
	NomDeclined  NominationStatus = "DECLINED"
	NomRejected  NominationStatus = "REJECTED"
	NomAttended  NominationStatus = "ATTENDED"
)

// IsValidNominationStatus Helper func in models package
func IsValidNominationStatus(status NominationStatus) bool {
	switch status {
	case NomPending, NomApproved, NomCompleted, NomDeclined, NomRejected, NomAttended:
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
	TrainingBehavioral TrainingCategory = "BEHAVIORAL"
)

// LessonContentType represents the type of content in a lesson.
type LessonContentType string

const (
	LessonVideo    LessonContentType = "VIDEO"
	LessonAudio    LessonContentType = "AUDIO"
	LessonPdf      LessonContentType = "PDF"
	LessonImage    LessonContentType = "IMAGE"
	LessonRichText LessonContentType = "RICH_TEXT"
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

// DeliveryMode enum represents the mdoe of delivery of the training
type DeliveryMode string

const (
	InPerson    DeliveryMode = "IN_PERSON"
	VirtualLink DeliveryMode = "VIRTUAL_LINK"
	Hybrid      DeliveryMode = "HYBRID"
	Elearning   DeliveryMode = "E_LEARNING"
)
