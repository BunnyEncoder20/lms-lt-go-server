package models

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Define custom types for context keys to prevent conlision with third party auths (not in this project at least)
// NOTE: never user string types for making context with values
type contextKey string

const (
	UserIDKey   contextKey = "user_ID"
	UserRoleKey contextKey = "user_role"
)
