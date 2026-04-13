// Package learning
package learning

import (
	"encoding/json"
	"errors"
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
	return &Handler{svc: svc, log: logger}
}

func (h *Handler) HandleGetPublishedCourses(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.FindPublishedCourses(r.Context())
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Data: data})
}

func (h *Handler) HandleBulkAssign(w http.ResponseWriter, r *http.Request) {
	adminID, err := auth.GetUserID(r.Context())
	if err != nil {
		utils.HandleError(w, r, h.log, errors.New("unauthorized"))
		return
	}

	var req models.BulkAssignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.HandleError(w, r, h.log, err)
		return
	}

	data, err := h.svc.BulkAssign(r.Context(), req, adminID)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Data: data})
}

func (h *Handler) HandleGetAllAssignments(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.FindAllAssignments(r.Context())
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Data: data})
}

func (h *Handler) HandleGetMyAssignments(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		utils.HandleError(w, r, h.log, errors.New("unauthorized"))
		return
	}

	data, err := h.svc.FindMyAssignments(r.Context(), userID)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Data: data})
}

func (h *Handler) HandleGetCoursePlayer(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		utils.HandleError(w, r, h.log, errors.New("unauthorized"))
		return
	}

	courseID := r.PathValue("courseId")
	if courseID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "missing course id"})
		return
	}

	data, err := h.svc.GetCoursePlayerData(r.Context(), courseID, userID)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Data: data})
}

func (h *Handler) HandleGetCourseNavigation(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		utils.HandleError(w, r, h.log, errors.New("unauthorized"))
		return
	}

	courseID := r.PathValue("courseId")
	if courseID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "missing course id"})
		return
	}

	data, err := h.svc.GetCourseNavigation(r.Context(), courseID, userID)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Data: data})
}

func (h *Handler) HandleGetNextLesson(w http.ResponseWriter, r *http.Request) {
	h.handleAdjacentLesson(w, r, "next")
}

func (h *Handler) HandleGetPreviousLesson(w http.ResponseWriter, r *http.Request) {
	h.handleAdjacentLesson(w, r, "previous")
}

func (h *Handler) handleAdjacentLesson(w http.ResponseWriter, r *http.Request, direction string) {
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		utils.HandleError(w, r, h.log, errors.New("unauthorized"))
		return
	}

	lessonID := r.PathValue("lessonId")
	if lessonID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "missing lesson id"})
		return
	}

	data, err := h.svc.GetAdjacentLesson(r.Context(), lessonID, userID, direction)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Data: data})
}

func (h *Handler) HandleUpdateProgress(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		utils.HandleError(w, r, h.log, errors.New("unauthorized"))
		return
	}

	lessonID := r.PathValue("lessonId")
	if lessonID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "missing lesson id"})
		return
	}

	var req models.UpdateProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.HandleError(w, r, h.log, err)
		return
	}

	data, err := h.svc.UpdateLessonProgress(r.Context(), lessonID, userID, req)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Data: data})
}

func (h *Handler) HandleHeartbeatProgress(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		utils.HandleError(w, r, h.log, errors.New("unauthorized"))
		return
	}

	lessonID := r.PathValue("lessonId")
	if lessonID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "missing lesson id"})
		return
	}

	var req models.HeartbeatProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.HandleError(w, r, h.log, err)
		return
	}

	data, err := h.svc.HeartbeatLessonProgress(r.Context(), lessonID, userID, req)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Data: data})
}

func (h *Handler) HandleCompleteLesson(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		utils.HandleError(w, r, h.log, errors.New("unauthorized"))
		return
	}

	lessonID := r.PathValue("lessonId")
	if lessonID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "missing lesson id"})
		return
	}

	var req models.CompleteLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.HandleError(w, r, h.log, err)
		return
	}

	data, err := h.svc.CompleteLesson(r.Context(), lessonID, userID, req)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Data: data})
}

func (h *Handler) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	h.log.Error("learning request failed", "error", err, "path", r.URL.Path)

	switch {
	case errors.Is(err, ErrLearningCourseNotFound),
		errors.Is(err, ErrLearningLessonNotFound),
		errors.Is(err, ErrLearningAssignmentMissing):
		utils.WriteJSON(w, http.StatusNotFound, models.JSONResponse{Success: false, Message: err.Error()})
	case errors.Is(err, ErrLearningAlreadyCompleted):
		utils.WriteJSON(w, http.StatusConflict, models.JSONResponse{Success: false, Message: err.Error()})
	case errors.Is(err, ErrLearningThresholdFailed),
		errors.Is(err, ErrLearningInvalidDirection):
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: err.Error()})
	default:
		utils.HandleError(w, r, h.log, err)
	}
}
