package models

type Role string

const (
	RoleAdmin          Role = "ADMIN"
	RoleManager        Role = "MANAGER"
	RoleEmployee       Role = "EMPLOYEE"
	RoleCourseDirector Role = "COURSE_DIRECTOR"
)
