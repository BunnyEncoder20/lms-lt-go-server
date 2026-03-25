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

	// Prevent admin from deactivating themselves
	if !req.IsActive {
		adminID, _ := auth.GetUserID(r.Context())
		if adminID == id {
			h.log.Warn("admin attempted to deactivate their own account via status update", "adminID", adminID)
			models.WriteJSON(w, http.StatusForbidden, models.JSONResponse{
				Success: false,
				Message: "you cannot deactivate your own admin account",
			})
			return
		}
	}

	if err := h.svc.DeactivateUser(r.Context(), id, req.IsActive); err != nil {
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

func (h *Handler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	// 1. Fetch the user id from the auth of admin making the req
	adminID, err := auth.GetUserID(r.Context())
	if err != nil {
		// Even though middleware checked auth, the context extraction could theoretically fail, so we handle that case here
		h.log.Error("failed to extract admin ID from context", "error", err)
		models.WriteJSON(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// 2. grab the ID the target user to be deleted from the url path
	targetUserID := r.PathValue("id")
	if targetUserID == "" {
		models.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Message: "internal server error",
		})
		return
	}

	// 3. Ensure admin doesn't delete themselves
	if adminID == targetUserID {
		h.log.Warn("admin attempted to delete their own account", "adminID", adminID)
		models.WriteJSON(w, http.StatusForbidden, models.JSONResponse{
			Success: false,
			Message: "you cannot delete your own admin account",
		})
		return
	}

	// 4. Hand off to service layer
	if err := h.svc.PermanentlyDeleteUser(r.Context(), targetUserID); err != nil {
		h.log.Error("failed to delete user",
			slog.String("adminID", adminID),
			slog.String("targetUserID", targetUserID),
			slog.Any("error", err),
		)
		models.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Message: "failed to delete the user",
		})
		return
	}

	// 5. Audit log - cruicial for prod logs
	h.log.Info("user deleted successfully",
		slog.String("action_by_admin", adminID),
		slog.String("deleted_user", targetUserID),
	)

	// 6. Returning success response
	models.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Message: "user permanently deleted successfully",
	})
}

// HandleSoftDeleteUser deactivates a user account (is_active = false)
func (h *Handler) HandleSoftDeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		models.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "missing user id",
		})
		return
	}

	// Prevent admin from deactivating themselves
	adminID, _ := auth.GetUserID(r.Context())
	if adminID == id {
		h.log.Warn("admin attempted to soft-delete their own account", "adminID", adminID)
		models.WriteJSON(w, http.StatusForbidden, models.JSONResponse{
			Success: false,
			Message: "you cannot deactivate your own admin account",
		})
		return
	}

	if err := h.svc.DeactivateUser(r.Context(), id, false); err != nil {
		h.log.Error("failed to soft delete user", "id", id, "error", err)
		models.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "failed to deactivate user",
		})
		return
	}

	models.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Message: "user deactivated successfully",
	})
}
