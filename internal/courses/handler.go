package courses

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"go-server/internal/auth"
	"go-server/internal/models"
	"go-server/internal/utils"

	"github.com/google/uuid"
)

const maxFileSize int64 = 100 * 1024 * 1024 // 100 MB

var allowedMimeTypes = map[string]struct{}{
	"video/mp4":                     {},
	"video/webm":                    {},
	"video/ogg":                     {},
	"audio/mpeg":                    {},
	"audio/wav":                     {},
	"audio/ogg":                     {},
	"application/pdf":               {},
	"image/jpeg":                    {},
	"image/png":                     {},
	"image/gif":                     {},
	"image/webp":                    {},
	"application/vnd.ms-powerpoint": {},
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": {},
	"application/vnd.oasis.opendocument.presentation":                           {},
}

type Handler struct {
	svc Service
	log *slog.Logger
}

func NewHandler(svc Service, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, log: logger}
}

func (h *Handler) HandleGetDashboardStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.GetDashboardStats(r.Context())
	if err != nil {
		h.log.Error("failed to get course dashboard stats", "error", err)
		utils.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{Success: false, Message: "failed to retrieve stats"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Data: stats})
}

func (h *Handler) HandleListCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := h.svc.FindAll(r.Context())
	if err != nil {
		h.log.Error("failed to list courses", "error", err)
		utils.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{Success: false, Message: "failed to retrieve courses"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Data: courses})
}

func (h *Handler) HandleListPublishedCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := h.svc.FindPublished(r.Context())
	if err != nil {
		h.log.Error("failed to list published courses", "error", err)
		utils.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{Success: false, Message: "failed to retrieve courses"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Data: courses})
}

func (h *Handler) HandleGetCourse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "missing course id"})
		return
	}

	course, err := h.svc.FindOne(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Data: course})
}

func (h *Handler) HandleGetPublishedCourse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "missing course id"})
		return
	}

	course, err := h.svc.FindOnePublished(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Data: course})
}

func (h *Handler) HandleCreateCourse(w http.ResponseWriter, r *http.Request) {
	authorID, err := auth.GetUserID(r.Context())
	if err != nil {
		utils.HandleError(w, r, h.log, errors.New("unauthorized"))
		return
	}

	var req models.CreateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.HandleError(w, r, h.log, err)
		return
	}

	course, err := h.svc.Create(r.Context(), req, authorID)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, models.JSONResponse{Success: true, Message: "course created successfully", Data: course})
}

func (h *Handler) HandleUpdateCourse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "missing course id"})
		return
	}

	var req models.UpdateCourseRequestV2
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.HandleError(w, r, h.log, err)
		return
	}

	course, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Message: "course updated successfully", Data: course})
}

func (h *Handler) HandlePublishCourse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "missing course id"})
		return
	}

	course, err := h.svc.Publish(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Message: "course published successfully", Data: course})
}

func (h *Handler) HandleArchiveCourse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "missing course id"})
		return
	}

	course, err := h.svc.Archive(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Message: "course archived successfully", Data: course})
}

func (h *Handler) HandleRestoreCourse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "missing course id"})
		return
	}

	course, err := h.svc.Restore(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Message: "course restored successfully", Data: course})
}

func (h *Handler) HandleDeleteCourse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "missing course id"})
		return
	}

	if err := h.svc.Remove(r.Context(), id); err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Message: "course deleted"})
}

func (h *Handler) HandleCreateModule(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("courseId")
	if courseID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "missing course id"})
		return
	}

	var req models.CreateCourseModuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.HandleError(w, r, h.log, err)
		return
	}

	module, err := h.svc.CreateModule(r.Context(), courseID, req)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, models.JSONResponse{Success: true, Message: "module created successfully", Data: module})
}

func (h *Handler) HandleUpdateModule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "missing module id"})
		return
	}

	var req models.UpdateCourseModuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.HandleError(w, r, h.log, err)
		return
	}

	module, err := h.svc.UpdateModule(r.Context(), id, req)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Message: "module updated successfully", Data: module})
}

func (h *Handler) HandleDeleteModule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "missing module id"})
		return
	}

	if err := h.svc.RemoveModule(r.Context(), id); err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Message: "module deleted"})
}

func (h *Handler) HandleReorderModules(w http.ResponseWriter, r *http.Request) {
	var req models.ReorderCourseModulesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.HandleError(w, r, h.log, err)
		return
	}

	if err := h.svc.ReorderModules(r.Context(), req); err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Message: "modules reordered"})
}

func (h *Handler) HandleCreateLesson(w http.ResponseWriter, r *http.Request) {
	moduleID := r.PathValue("moduleId")
	if moduleID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "missing module id"})
		return
	}

	var req models.CreateLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.HandleError(w, r, h.log, err)
		return
	}

	lesson, err := h.svc.CreateLesson(r.Context(), moduleID, req)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, models.JSONResponse{Success: true, Message: "lesson created successfully", Data: lesson})
}

func (h *Handler) HandleUpdateLesson(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "missing lesson id"})
		return
	}

	var req models.UpdateLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.HandleError(w, r, h.log, err)
		return
	}

	lesson, err := h.svc.UpdateLesson(r.Context(), id, req)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Message: "lesson updated successfully", Data: lesson})
}

func (h *Handler) HandleDeleteLesson(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "missing lesson id"})
		return
	}

	if err := h.svc.RemoveLesson(r.Context(), id); err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Message: "lesson deleted"})
}

func (h *Handler) HandleReorderLessons(w http.ResponseWriter, r *http.Request) {
	var req models.ReorderLessonsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.HandleError(w, r, h.log, err)
		return
	}

	if err := h.svc.ReorderLessons(r.Context(), req); err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{Success: true, Message: "lessons reordered"})
}

func (h *Handler) HandleUploadFile(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize+1024)
	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "invalid multipart payload or file too large"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "no file provided"})
		return
	}
	defer file.Close()

	if header.Size > maxFileSize {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "file exceeds maximum size of 100MB"})
		return
	}

	mimeType, err := sniffMimeType(file, header)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: "failed to inspect file"})
		return
	}
	if _, ok := allowedMimeTypes[mimeType]; !ok {
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: fmt.Sprintf("file type %s not allowed", mimeType)})
		return
	}

	if err := os.MkdirAll(uploadsDir(), 0o755); err != nil {
		h.log.Error("failed to create uploads dir", "error", err)
		utils.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{Success: false, Message: "failed to save file"})
		return
	}

	ext := filepath.Ext(header.Filename)
	uniqueName := fmt.Sprintf("%s%s", uuid.NewString(), ext)
	target := filepath.Join(uploadsDir(), uniqueName)

	dst, err := os.Create(target)
	if err != nil {
		h.log.Error("failed to create upload file", "error", err)
		utils.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{Success: false, Message: "failed to save file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		h.log.Error("failed to write upload file", "error", err)
		utils.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{Success: false, Message: "failed to save file"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data: map[string]any{
			"url":          "/api/uploads/" + uniqueName,
			"originalName": header.Filename,
			"mimeType":     mimeType,
			"size":         header.Size,
		},
	})
}

func (h *Handler) HandleServeUpload(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	if filename == "" {
		http.NotFound(w, r)
		return
	}

	cleaned := filepath.Base(filename)
	if cleaned != filename || strings.Contains(cleaned, "..") {
		http.NotFound(w, r)
		return
	}

	target := filepath.Join(uploadsDir(), cleaned)
	if _, err := os.Stat(target); err != nil {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, target)
}

func uploadsDir() string {
	return filepath.Join("uploads")
}

func sniffMimeType(file multipart.File, header *multipart.FileHeader) (string, error) {
	declared := header.Header.Get("Content-Type")
	if declared != "" {
		if idx := strings.Index(declared, ";"); idx > 0 {
			declared = declared[:idx]
		}
	}

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	detected := http.DetectContentType(buffer[:n])
	if declared != "" {
		return declared, nil
	}
	if idx := strings.Index(detected, ";"); idx > 0 {
		detected = detected[:idx]
	}
	return detected, nil
}

func (h *Handler) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	h.log.Error("courses request failed", "error", err, "path", r.URL.Path)

	switch {
	case errors.Is(err, ErrCourseNotFound), errors.Is(err, ErrModuleNotFound), errors.Is(err, ErrLessonNotFound):
		utils.WriteJSON(w, http.StatusNotFound, models.JSONResponse{Success: false, Message: err.Error()})
	case errors.Is(err, ErrInvalidCourseID), errors.Is(err, ErrInvalidModuleID), errors.Is(err, ErrInvalidLessonID):
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: err.Error()})
	case errors.Is(err, ErrArchivedCourse),
		errors.Is(err, ErrCourseAlreadyPub),
		errors.Is(err, ErrRestoreOnlyArchived),
		errors.Is(err, ErrCourseNoModules),
		errors.Is(err, ErrCourseNoLessons):
		utils.WriteJSON(w, http.StatusBadRequest, models.JSONResponse{Success: false, Message: err.Error()})
	default:
		utils.HandleError(w, r, h.log, err)
	}
}
