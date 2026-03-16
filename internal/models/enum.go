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
	NomRejected  NominationStatus = "REJECTED"
	NomCompleted NominationStatus = "COMPLETED"
	NomAttended  NominationStatus = "ATTENDED"
)

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
