// Package users contains all the handlers and services related to user management, including creating users, listing users, and updating user status. It interacts with the database through the Service interface and handles HTTP requests in the Handler struct.
package users

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"go-server/internal/auth"
	"go-server/internal/models"
)

type Handler struct {
	svc Service
	log *slog.Logger
}

func NewHandler(svc Service, logger *slog.Logger) *Handler {
	return &Handler{
		svc: svc,
		log: logger,
	}
}

func (h *Handler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	// In a real app, you'd extract the ID from the URL using a router like chi or mux
	// For standard library http.ServeMux in Go 1.22+, you can use r.PathValue("id")
	id := r.PathValue("id")
	if id == "" {
		models.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "missing user id",
		})
		h.log.Debug("GetUser called without user ID")
		return
	}

	user, err := h.svc.FindOne(r.Context(), id)
	if err != nil {
		h.log.Error("user not found", "id", id, "error", err)
		models.WriteJSON(w, http.StatusNotFound, models.JSONResponse{
			Success: false,
			Message: "user not found",
		})
		return
	}

	models.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    user,
	})
}

func (h *Handler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.FindAll(r.Context())
	if err != nil {
		h.log.Error("failed to list users", "error", err)
		models.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "failed to retrieve users",
		})
		return
	}

	models.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    users,
	})
}

func (h *Handler) HandleGetMyTeam(w http.ResponseWriter, r *http.Request) {
	managerID, err := auth.GetUserID(r.Context())
	if err != nil {
		h.log.Error("failed to get manager ID from context", "error", err)
		models.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
			Success: false,
			Message: "unauthorized",
		})
		return
	}

	team, err := h.svc.GetMyTeam(r.Context(), managerID)
	if err != nil {
		h.log.Error("could not retreive team members", "managerID", managerID, "error", err)
		models.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "could not retrieve team members",
		})
		return
	}

	models.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    team,
	})
}

func (h *Handler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "invalid request body",
		})
		return
	}

	user, err := h.svc.Create(r.Context(), req)
	if err != nil {
		h.log.Error("failed to create user", "error", err)
		models.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "failed to create user",
		})
		return
	}

	models.WriteJSON(w, http.StatusCreated, models.JSONResponse{
		Success: true,
		Message: "user created successfully",
		Data:    user,
	})
}

func (h *Handler) HandleUpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req models.UserStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "invalid request body",
		})
		return
	}

	err := h.svc.UpdateStatus(r.Context(), id, req.IsActive)
	if err != nil {
		h.log.Error("failed to update user status", "id", id, "error", err)
		models.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "failed to update user status",
		})
		return
	}

	models.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Message: "user status updated",
	})
}

func (h *Handler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	// 1. Id the user
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		h.log.Warn("unauthorized profile update attempt", "error", err)
		models.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
			Success: false,
			Message: "unauthorized",
		})
		return
	}

	// 2. Decode the incoming json payload
	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("invalid json payload for profile update", "userID", userID, "error", err)
		models.WriteJSON(w, http.StatusBadGateway, models.JSONResponse{
			Success: false,
			Message: "invalid request body",
		})
		return
	}

	// 3. Hand off data to the service layer
	updatedProfile, err := h.svc.Update(r.Context(), userID, req)
	if err != nil {
		h.log.Error("failed to update user data", "userID", userID, "error", err)
		models.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "failed to update user data",
		})
		return
	}

	// 4. Return Sucess response
	models.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Message: "profile updated successfully",
		Data:    updatedProfile,
	})
}
