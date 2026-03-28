// Package trainings contains all the handlers and services related to training management.
package trainings

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

// HandleListTraining handles requests to list all trainings.
// Accessible by: All authenticated users.
func (h *Handler) HandleListTraining(w http.ResponseWriter, r *http.Request) {
	trainings, err := h.svc.List(r.Context())
	if err != nil {
		h.log.Error("failed to list trainings", "error", err)
		models.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "failed to retrieve trainings",
		})
		return
	}

	models.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    trainings,
	})
}

// HandleGetTraining handles requests to get a single training by ID.
// Accessible by: All authenticated users.
func (h *Handler) HandleGetTraining(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		models.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "missing training id",
		})
		h.log.Debug("GetTraining called without training ID")
		return
	}

	training, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.log.Error("training not found", "id", id, "error", err)
		models.WriteJSON(w, http.StatusNotFound, models.JSONResponse{
			Success: false,
			Message: "training not found",
		})
		return
	}

	models.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    training,
	})
}

// HandleGetTrainingCategory handles requests to get trainings by category.
// Accessible by: All authenticated users.
func (h *Handler) HandleGetTrainingCategory(w http.ResponseWriter, r *http.Request) {
	// Get category from path or query parameter
	category := r.PathValue("category")
	if category == "" {
		category = r.URL.Query().Get("category")
	}

	if category == "" {
		models.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "missing category parameter",
		})
		h.log.Debug("GetTrainingCategory called without category")
		return
	}

	trainings, err := h.svc.GetByCategory(r.Context(), category)
	if err != nil {
		h.log.Error("failed to get trainings by category", "category", category, "error", err)
		models.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "failed to retrieve trainings",
		})
		return
	}

	models.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    trainings,
	})
}

// HandleGetUpcomingTraining handles requests to get upcoming trainings.
// Accessible by: All authenticated users.
func (h *Handler) HandleGetUpcomingTraining(w http.ResponseWriter, r *http.Request) {
	trainings, err := h.svc.GetUpcoming(r.Context())
	if err != nil {
		h.log.Error("failed to get upcoming trainings", "error", err)
		models.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "failed to retrieve upcoming trainings",
		})
		return
	}

	models.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    trainings,
	})
}

// HandleGetEmployeeTraining handles requests to get trainings for the current employee.
// Accessible by: All authenticated users (returns own trainings).
func (h *Handler) HandleGetEmployeeTraining(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		h.log.Error("failed to get user ID from context", "error", err)
		models.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
			Success: false,
			Message: "unauthorized",
		})
		return
	}

	trainings, err := h.svc.GetEmployeeTrainings(r.Context(), userID)
	if err != nil {
		h.log.Error("failed to get employee trainings", "userID", userID, "error", err)
		models.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "failed to retrieve employee trainings",
		})
		return
	}

	models.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    trainings,
	})
}

// HandleCreateTraining handles requests to create a new training.
// Security: Admin protected route.
func (h *Handler) HandleCreateTraining(w http.ResponseWriter, r *http.Request) {
	// 1. Get the admin ID from context for audit logging
	adminID, err := auth.GetUserID(r.Context())
	if err != nil {
		h.log.Error("failed to get admin ID from context", "error", err)
		models.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
			Success: false,
			Message: "unauthorized",
		})
		return
	}

	// 2. Decode the request body
	var req models.CreateTrainingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("invalid request body for create training", "error", err)
		models.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "invalid request body",
		})
		return
	}

	// 3. Create the training
	training, err := h.svc.Create(r.Context(), adminID, req)
	if err != nil {
		h.log.Error("failed to create training", "adminID", adminID, "error", err)
		models.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "failed to create training",
		})
		return
	}

	// 4. Audit log
	h.log.Info("training created successfully",
		slog.String("created_by_admin", adminID),
		slog.String("training_id", training.ID),
		slog.String("training_title", training.Title),
	)

	// 5. Return success response
	models.WriteJSON(w, http.StatusCreated, models.JSONResponse{
		Success: true,
		Message: "training created successfully",
		Data:    training,
	})
}

// HandleUpdateTraining handles requests to update an existing training.
// Security: Admin protected route.
func (h *Handler) HandleUpdateTraining(w http.ResponseWriter, r *http.Request) {
	// 1. Get the admin ID from context for audit logging
	adminID, err := auth.GetUserID(r.Context())
	if err != nil {
		h.log.Error("failed to get admin ID from context", "error", err)
		models.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
			Success: false,
			Message: "unauthorized",
		})
		return
	}

	// 2. Get training ID from path
	trainingID := r.PathValue("id")
	if trainingID == "" {
		models.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "missing training id",
		})
		return
	}

	// 3. Decode the request body
	var req models.UpdateTrainingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("invalid request body for update training", "error", err)
		models.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "invalid request body",
		})
		return
	}

	// 4. Update the training
	training, err := h.svc.Update(r.Context(), trainingID, req)
	if err != nil {
		h.log.Error("failed to update training",
			slog.String("adminID", adminID),
			slog.String("trainingID", trainingID),
			slog.Any("error", err),
		)
		models.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "failed to update training",
		})
		return
	}

	// 5. Audit log
	h.log.Info("training updated successfully",
		slog.String("updated_by_admin", adminID),
		slog.String("training_id", training.ID),
	)

	// 6. Return success response
	models.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Message: "training updated successfully",
		Data:    training,
	})
}

// HandleDeleteTraining handles requests to permanently delete a training.
// Security: Admin protected route.
func (h *Handler) HandleDeleteTraining(w http.ResponseWriter, r *http.Request) {
	// 1. Get the admin ID from context for audit logging
	adminID, err := auth.GetUserID(r.Context())
	if err != nil {
		h.log.Error("failed to get admin ID from context", "error", err)
		models.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
			Success: false,
			Message: "unauthorized",
		})
		return
	}

	// 2. Get training ID from path
	trainingID := r.PathValue("id")
	if trainingID == "" {
		models.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "missing training id",
		})
		return
	}

	// 3. Delete the training
	if err := h.svc.Delete(r.Context(), trainingID); err != nil {
		h.log.Error("failed to delete training",
			slog.String("adminID", adminID),
			slog.String("trainingID", trainingID),
			slog.Any("error", err),
		)
		models.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "failed to delete training",
		})
		return
	}

	// 4. Audit log - crucial for production
	h.log.Info("training deleted successfully",
		slog.String("deleted_by_admin", adminID),
		slog.String("deleted_training_id", trainingID),
	)

	// 5. Return success response
	models.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Message: "training deleted successfully",
	})
}
