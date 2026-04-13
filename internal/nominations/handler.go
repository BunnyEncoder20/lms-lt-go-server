// Package nominations contains all the handlers and services related to nomination management.
package nominations

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"go-server/internal/auth"
	"go-server/internal/models"
	"go-server/internal/utils"
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

// HandleNominateEmployees handles requests for managers to nominate employees for training.
// Security: Manager protected route.
func (h *Handler) HandleNominateEmployees(w http.ResponseWriter, r *http.Request) {
	// 1. Get manager ID from context
	managerID, err := auth.GetUserID(r.Context())
	if err != nil {
		h.log.Error("failed to get manager ID from context", "error", err)
		utils.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
			Success: false,
			Message: "unauthorized",
		})
		return
	}

	// 2. Decode request body
	var req models.CreateNominationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("invalid request body for nominate employees", "error", err)
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "invalid request body",
		})
		return
	}

	// 3. Validate request
	if req.TrainingID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "training_id is required",
		})
		return
	}

	if len(req.UserIDs) == 0 {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "user_ids is required",
		})
		return
	}

	// 4. Create nominations
	nominations, err := h.svc.NominateEmployees(r.Context(), managerID, req)
	if err != nil {
		h.log.Error("failed to nominate employees",
			slog.String("managerID", managerID),
			slog.Any("error", err),
		)
		utils.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// 5. Audit log
	h.log.Info("employees nominated successfully",
		slog.String("manager_id", managerID),
		slog.String("training_id", req.TrainingID),
		slog.Int("nomination_count", len(nominations)),
	)

	// 6. Return success response
	utils.WriteJSON(w, http.StatusCreated, models.JSONResponse{
		Success: true,
		Message: "employees nominated successfully",
		Data:    nominations,
	})
}

// HandleSelfNomination handles requests for employees to self-nominate for training.
// Security: Authenticated users.
func (h *Handler) HandleSelfNomination(w http.ResponseWriter, r *http.Request) {
	// 1. Get user ID from context
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		h.log.Error("failed to get user ID from context", "error", err)
		utils.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
			Success: false,
			Message: "unauthorized",
		})
		return
	}

	// 2. Decode request body
	var req models.SelfNominationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("invalid request body for self nomination", "error", err)
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "invalid request body",
		})
		return
	}

	// 3. Validate request
	if req.TrainingID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "training_id is required",
		})
		return
	}

	// 4. Create self-nomination
	nomination, err := h.svc.SelfNomination(r.Context(), userID, req)
	if err != nil {
		h.log.Error("failed to create self nomination",
			slog.String("userID", userID),
			slog.Any("error", err),
		)
		utils.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// 5. Audit log
	h.log.Info("self nomination created successfully",
		slog.String("user_id", userID),
		slog.String("training_id", req.TrainingID),
	)

	// 6. Return success response
	utils.WriteJSON(w, http.StatusCreated, models.JSONResponse{
		Success: true,
		Message: "self nomination created successfully",
		Data:    nomination,
	})
}

// HandleRespondToNomination handles requests for employees to respond to nominations.
// Security: Authenticated users.
func (h *Handler) HandleRespondToNomination(w http.ResponseWriter, r *http.Request) {
	// 1. Get user ID from context
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		h.log.Error("failed to get user ID from context", "error", err)
		utils.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
			Success: false,
			Message: "unauthorized",
		})
		return
	}

	// 2. Get nomination ID from path
	nominationID := r.PathValue("id")
	if nominationID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "missing nomination id",
		})
		return
	}

	// 3. Decode request body
	var req models.NominationResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("invalid request body for respond to nomination", "error", err)
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "invalid request body",
		})
		return
	}

	// 4. Respond to nomination
	nomination, err := h.svc.RespondToNomination(r.Context(), userID, nominationID, req)
	if err != nil {
		h.log.Error("failed to respond to nomination",
			slog.String("userID", userID),
			slog.String("nominationID", nominationID),
			slog.Any("error", err),
		)
		status := http.StatusInternalServerError
		if err.Error() == "unauthorized: nomination does not belong to this employee" {
			status = http.StatusForbidden
		}
		utils.WriteJSON(w, status, models.JSONResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// 5. Audit log
	h.log.Info("nomination response recorded",
		slog.String("user_id", userID),
		slog.String("nomination_id", nominationID),
		slog.String("response", req.Status),
	)

	// 6. Return success response
	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Message: "nomination response recorded",
		Data:    nomination,
	})
}

// HandleRespondToSelfNomination handles requests for managers to approve/reject self-nominations.
// Security: Manager protected route.
func (h *Handler) HandleRespondToSelfNomination(w http.ResponseWriter, r *http.Request) {
	// 1. Get manager ID from context
	managerID, err := auth.GetUserID(r.Context())
	if err != nil {
		h.log.Error("failed to get manager ID from context", "error", err)
		utils.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
			Success: false,
			Message: "unauthorized",
		})
		return
	}

	// 2. Get nomination ID from path
	nominationID := r.PathValue("id")
	if nominationID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "missing nomination id",
		})
		return
	}

	// 3. Decode request body
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("invalid request body for respond to self nomination", "error", err)
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "invalid request body",
		})
		return
	}

	// 4. Validate status
	var status models.NominationStatus
	switch req.Status {
	case "APPROVED":
		status = models.NomEnrolled
	case "REJECTED":
		status = models.NomRejected
	default:
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "invalid status: must be APPROVED or REJECTED",
		})
		return
	}

	// 5. Respond to self-nomination
	nomination, err := h.svc.RespondToSelfNomination(r.Context(), managerID, nominationID, status)
	if err != nil {
		h.log.Error("failed to respond to self nomination",
			slog.String("managerID", managerID),
			slog.String("nominationID", nominationID),
			slog.Any("error", err),
		)
		status := http.StatusInternalServerError
		if err.Error() == "unauthorized: can only respond to nominations from your team" {
			status = http.StatusForbidden
		}
		utils.WriteJSON(w, status, models.JSONResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// 6. Audit log
	h.log.Info("self nomination response recorded",
		slog.String("manager_id", managerID),
		slog.String("nomination_id", nominationID),
		slog.String("response", req.Status),
	)

	// 7. Return success response
	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Message: "self nomination " + req.Status + " successfully",
		Data:    nomination,
	})
}

// HandleGetMyNominations handles requests for employees to get their nominations.
// Security: Authenticated users.
func (h *Handler) HandleGetMyNominations(w http.ResponseWriter, r *http.Request) {
	// 1. Get user ID from context
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		h.log.Error("failed to get user ID from context", "error", err)
		utils.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
			Success: false,
			Message: "unauthorized",
		})
		return
	}

	// 2. Get nominations
	nominations, err := h.svc.GetMyNominations(r.Context(), userID)
	if err != nil {
		h.log.Error("failed to get my nominations",
			slog.String("userID", userID),
			slog.Any("error", err),
		)
		utils.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "failed to retrieve nominations",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Message: fmt.Sprintf("retrieved %d nominations", len(nominations)),
		Data:    nominations,
	})
}

// HandleGetTeamNominations handles requests for managers to get their team nominations.
// Security: Manager protected route.
func (h *Handler) HandleGetTeamNominations(w http.ResponseWriter, r *http.Request) {
	// 1. Get manager ID from context
	managerID, err := auth.GetUserID(r.Context())
	if err != nil {
		h.log.Error("failed to get manager ID from context", "error", err)
		utils.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
			Success: false,
			Message: "unauthorized",
		})
		return
	}

	// 2. Get team nominations
	nominations, err := h.svc.GetTeamNominations(r.Context(), managerID)
	if err != nil {
		h.log.Error("failed to get team nominations",
			slog.String("managerID", managerID),
			slog.Any("error", err),
		)
		utils.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "failed to retrieve team nominations",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    nominations,
	})
}

// HandleGetAllNominations handles requests to get all nominations with filters.
// Security: Admin protected route.
func (h *Handler) HandleGetAllNominations(w http.ResponseWriter, r *http.Request) {
	// 1. Build filters from query parameters
	filters := models.NominationFilters{}

	if status := r.URL.Query().Get("status"); status != "" {
		filters.Status = &status
	}
	if trainingID := r.URL.Query().Get("training_id"); trainingID != "" {
		filters.TrainingID = &trainingID
	}
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		filters.UserID = &userID
	}
	if managerID := r.URL.Query().Get("manager_id"); managerID != "" {
		filters.ManagerID = &managerID
	}

	// 2. Get nominations
	nominations, err := h.svc.GetAllNominations(r.Context(), filters)
	if err != nil {
		h.log.Error("failed to get all nominations", "error", err)
		utils.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "failed to retrieve nominations",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    nominations,
	})
}

// HandleUpdateNominationStatus handles requests for admin to update nomination status.
// Security: Admin protected route.
func (h *Handler) HandleUpdateNominationStatus(w http.ResponseWriter, r *http.Request) {
	// 1. Get nomination ID from path
	nominationID := r.PathValue("id")
	if nominationID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "missing nomination id",
		})
		return
	}

	// 2. Get admin ID from context for audit
	adminID, err := auth.GetUserID(r.Context())
	if err != nil {
		h.log.Error("failed to get admin ID from context", "error", err)
		utils.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
			Success: false,
			Message: "unauthorized",
		})
		return
	}

	// 3. Decode request body
	var req models.UpdateNominationStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("invalid request body for update nomination status", "error", err)
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "invalid request body",
		})
		return
	}

	// 4. Update nomination status
	nomination, err := h.svc.UpdateNominationStatus(r.Context(), nominationID, req.Status)
	if err != nil {
		h.log.Error("failed to update nomination status",
			slog.String("adminID", adminID),
			slog.String("nominationID", nominationID),
			slog.Any("error", err),
		)
		utils.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// 5. Audit log
	h.log.Info("nomination status updated by admin",
		slog.String("admin_id", adminID),
		slog.String("nomination_id", nominationID),
		slog.String("new_status", string(req.Status)),
	)

	// 6. Return success response
	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Message: "nomination status updated successfully",
		Data:    nomination,
	})
}

// HandleGetManagerDashboard handles requests for manager dashboard KPIs.
// Security: Manager protected route.
func (h *Handler) HandleGetManagerDashboard(w http.ResponseWriter, r *http.Request) {
	// 1. Get manager ID from context
	managerID, err := auth.GetUserID(r.Context())
	if err != nil {
		h.log.Error("failed to get manager ID from context", "error", err)
		utils.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
			Success: false,
			Message: "unauthorized",
		})
		return
	}

	// 2. Get dashboard data
	dashboard, err := h.svc.GetManagerDashboard(r.Context(), managerID)
	if err != nil {
		h.log.Error("failed to get manager dashboard",
			slog.String("managerID", managerID),
			slog.Any("error", err),
		)
		utils.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "failed to retrieve dashboard",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    dashboard,
	})
}

// HandleGetEmployeeDashboard handles requests for employee dashboard KPIs.
// Security: Authenticated users.
func (h *Handler) HandleGetEmployeeDashboard(w http.ResponseWriter, r *http.Request) {
	// 1. Get user ID from context
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		h.log.Error("failed to get user ID from context", "error", err)
		utils.WriteJSON(w, http.StatusUnauthorized, models.JSONResponse{
			Success: false,
			Message: "unauthorized",
		})
		return
	}

	// 2. Get dashboard data
	dashboard, err := h.svc.GetEmployeeDashboard(r.Context(), userID)
	if err != nil {
		h.log.Error("failed to get employee dashboard",
			slog.String("userID", userID),
			slog.Any("error", err),
		)
		utils.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "failed to retrieve dashboard",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    dashboard,
	})
}

// HandleGetAllPublishedCourses handles requests to get all published courses.
// Security: All authenticated users.
func (h *Handler) HandleGetAllPublishedCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := h.svc.GetAllPublishedCourses(r.Context())
	if err != nil {
		h.log.Error("failed to get published courses", "error", err)
		utils.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "failed to retrieve courses",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    courses,
	})
}
