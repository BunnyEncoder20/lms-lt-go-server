package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"go-server/internal/models"
)

type Handler struct {
	svc Service // references the Service interface from the service.go file of this module
}

func NewHandler(svc Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) HandleMe(w http.ResponseWriter, r *http.Request) {
	// checking if the values are actually in the context
	userID, err := GetUserID(r.Context())
	if err != nil {
		log.Printf("Context values missing: %v\n", err)
		models.WriteJSON(w, http.StatusNotFound, models.JSONResponse{
			Success: false,
			Message: "the token data could not extracted from the context, this should never happen if the middleware is working correctly",
		})
		return
	}
	userRole, err := GetUserRole(r.Context())
	if err != nil {
		log.Printf("Context value missing: %v\n", err)
		models.WriteJSON(w, http.StatusNotFound, models.JSONResponse{
			Success: false,
			Message: "the token data could not extracted from the context, this should never happen if the middleware is working correctly",
		})
		return
	}

	models.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Message: "Returing user data from auth token",
		Data: struct {
			UserID   any `json:"userId"`
			UserRole any `json:"userRole"`
		}{
			UserID:   userID,
			UserRole: userRole,
		}, // This is just an example, you would typically query the database for the user's info using the ID from the claims
	})
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Message: "invalid request body",
		})
		return
	}

	// Call the business logic in the service layer
	tokenString, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		models.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
			Message: "invalid email or password",
		})
		return
	}

	// Making and Setting a encrypted httpOnly cookie with the token
	http.SetCookie(w, &http.Cookie{
		Name:     "access-token",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,                                 // Prevents JavaScript access to the cookie
		Secure:   os.Getenv("APP_ENV") == "production", // Only sent over HTTPS
		SameSite: http.SameSiteStrictMode,              // Prevents CSRF attacks
		Path:     "/",                                  // Available to the entire site/routes
	})

	models.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Message: "login successful",
	})
}
