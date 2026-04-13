package notifications

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

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

func (h *Handler) HandleGetHrFeed(w http.ResponseWriter, r *http.Request) {
	limit := parseFeedLimit(r.URL.Query().Get("limit"))
	items, err := h.svc.GetHrFeed(r.Context(), limit)
	if err != nil {
		utils.HandleError(w, r, h.log, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    models.HrNotificationFeedResponse{Items: items},
	})
}

func parseFeedLimit(raw string) int {
	if strings.TrimSpace(raw) == "" {
		return 30
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return 30
	}
	if parsed < 1 {
		return 1
	}
	if parsed > 100 {
		return 100
	}
	return parsed
}
