package admin

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

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

func (h *Handler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{
			Success: false,
			Message: "missing user id",
		})
		return
	}

	user, err := h.svc.GetUser(r.Context(), id)
	if err != nil {
		h.log.Error("admin user lookup failed", "id", id, "error", err)
		utils.WriteJSON(w, http.StatusNotFound, models.JSONResponse{
			Success: false,
			Message: "user not found",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    user,
	})
}

func (h *Handler) HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.GetUsers(r.Context())
	if err != nil {
		h.log.Error("failed to list users", "error", err)
		utils.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "failed to retrieve users",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    users,
	})
}

func (h *Handler) HandleGetKpis(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetKpis(r.Context())
	if err != nil {
		h.log.Error("there was an error retrieving the kpis", "err", err)
		utils.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "Failed to retrieve KPIs",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    data,
	})
}

func (h *Handler) HandleGetMonthlyStats(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetMonthlyStats(r.Context())
	if err != nil {
		h.log.Error("there was an error retrieving monthly stats", "err", err)
		utils.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "Failed to retrieve monthly stats",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    data,
	})
}

func (h *Handler) HandleGetCategoryDistribution(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetCategoryDistribution(r.Context())
	if err != nil {
		h.log.Error("there was an error retrieving category distribution", "err", err)
		utils.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "Failed to retrieve category distribution",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    data,
	})
}

func (h *Handler) HandleGetClusterStats(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetClusterStats(r.Context())
	if err != nil {
		h.log.Error("there was an error retrieving cluster stats", "err", err)
		utils.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "Failed to retrieve cluster stats",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    data,
	})
}

func (h *Handler) HandleImportHistory(w http.ResponseWriter, r *http.Request) {
	// 1. Parse the multipart form (Limit to 10MB in memory, rest goes to tmp filies)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		utils.HandleError(w, r, h.log, fmt.Errorf("faild to parse form: %w", err))
	}

	// 2. Grab the file by the form key
	file, header, err := r.FormFile("file")
	if err != nil {
		utils.HandleError(w, r, h.log, errors.New("file is required"))
		return
	}
	defer file.Close() // WARN: Critical: prevent memory leaks

	// 3. Pass the raw file directly to the service
	res, err := h.svc.ImportHistoricalWorkbook(r.Context(), file, header.Filename)
	if err != nil {
		utils.HandleError(w, r, h.log, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Message: "historial training report imported successfully",
		Data:    res,
	})
}
