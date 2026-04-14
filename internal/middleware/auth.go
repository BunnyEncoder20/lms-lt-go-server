// Package middleware provides middleware for the application.
package middleware

import (
	"errors"
	"net/http"
	"os"
	"slices"

	"go-server/internal/auth"
	"go-server/internal/models"
	"go-server/internal/utils"

	"github.com/golang-jwt/jwt/v5"
)

// RequireAuth is a middleware for attaching the user's info onto the request context
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Grab the cookie
		cookie, err := r.Cookie("access-token")
		if err != nil {
			utils.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
				Message: "unauthorized: missing token",
			})
			return
		}

		// 2. Parse and validate the JWT
		tokenString := cookie.Value
		secret := []byte(os.Getenv("JWT_SECRET"))

		// We pass an empty instance of our custom claims struct to the jwt.ParseWithClaims function, so that it can populate it with the data from the token if it's valid.
		claims := &auth.JWTClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return secret, nil
		})

		if err != nil || !token.Valid {
			msg := "unauthorized: invalid token"
			if errors.Is(err, jwt.ErrTokenExpired) {
				msg = "unauthorized: token expired"
			}
			utils.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
				Message: msg,
			})
			return
		}

		// 3. Put the extracted claims into the request context
		ctx := auth.ContextWithUserID(r.Context(), claims.UserID)
		ctx = auth.ContextWithUserRole(ctx, claims.Role)

		// 4. Create a new request with the updated context and pass it to the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRoles check if the user's role (extracted by RequireAuth) is in the allowed list
// The variadic parameter (...string) allows us to pass one or more roles.
func RequireRoles(allowedRoles ...string) func(http.Handler) http.Handler {
	// this returns the actual middleware function
	return func(next http.Handler) http.Handler {
		// this returns the HTTP handler
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Pull the role safely out of the context
			userRole, err := auth.GetUserRole(r.Context())
			if err != nil || userRole == "" {
				// this acts as a failsafe in case RequireRoles is accidentally used without RequireAuth
				utils.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
					Message: "unauthorized: role identity missing",
				})
				return
			} else if slices.Contains(allowedRoles, userRole) {
				// 2. Check if their roles exists in the allowed list
				// Access granted Pass the baton to the target handler
				next.ServeHTTP(w, r)
				return
			}

			// 3. If the loop finishes without a match, block the request
			utils.WriteJSON(w, http.StatusForbidden, models.JSONResponse{
				Message: "forbidden: insufficient permissions",
			})
		})
	}
}
