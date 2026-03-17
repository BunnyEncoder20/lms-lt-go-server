// Package middleware provides middleware for the application.
package middleware

import (
	"context"
	"errors"
	"net/http"
	"os"
	"slices"

	"go-server/internal/auth"
	"go-server/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

// 1. Define custom types for context keys to prevent conlision with third party auths (not in this project at least)
// NOTE: never user string types for making context with values
type contextKey string

const (
	UserIDKey   contextKey = "user_ID"
	UserRoleKey contextKey = "user_role"
)

// RequireAuth is a middleware for attaching the user's info onto the request context
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Grabe the cookie
		cookie, err := r.Cookie("access-token")
		if err != nil {
			models.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
				Message: "unauthorized: missing token",
			})
		}

		// 2. Parse and validate the JWT
		tokenString := cookie.Value
		secret := []byte(os.Getenv("JWT_SECRET"))

		// We pass an empty instance of our custom claims struct to the jwt.ParseWithClaims function, so that it can populate it with the data from the token if it's valid.
		claims := &auth.JWTClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}

			return secret, nil
		})

		if err != nil || !token.Valid {
			models.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
				Message: "unauthorized: invalid or expired token",
			})
		}

		// 3. Put the extracted claims into the request context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

		// 4. Create a new request with the updated context and pass it to the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// ReqruireRoles check if the user's role (ectracted by RequireAuth) is in hte allowed list
// The variadic parameter (...string) allows us to pass one or more roles.
func RequireRoles(allowedRoles ...string) func(http.HandlerFunc) http.HandlerFunc {
	// tis returns the actual middleware function
	return func(next http.HandlerFunc) http.HandlerFunc {
		// tis returns the HTTP handler
		return func(w http.ResponseWriter, r *http.Request) {
			// 1. Pul the role safely out of the context
			// We use the Type Assertion .(string) again because Context values are stored as empty interfaces (any)
			userRole, ok := r.Context().Value(UserRoleKey).(string)
			if !ok || userRole == "" {
				// this acts as a failsafe in case RequuredRoles is accedentally used without RequireAuth
				models.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
					Message: "unauthorized: role identity missing",
				})
				return
			}

			// 2. Check if their roles exists in the allowed list
			if slices.Contains(allowedRoles, userRole) {
				// Access granted Pass the baton to the target handler
				next.ServeHTTP(w, r)
				return
			}
		}
	}
}
