package admin

import (
	"log/slog"
	"net/http"

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

func (h *Handler) HandleGetKpis(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetKpis(r.Context())
	if err != nil {
		h.log.Error("there was an error retrieving the kpis", "err", err)
		models.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "Failed to retrieve KPIs",
		})
		return
	}

	models.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    data,
	})
}

func (h *Handler) HandleGetMonthlyStats(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetMonthlyStats(r.Context())
	if err != nil {
		h.log.Error("there was an error retrieving monthly stats", "err", err)
		models.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "Failed to retrieve monthly stats",
		})
		return
	}

	models.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    data,
	})
}

func (h *Handler) HandleGetCategoryDistribution(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetCategoryDistribution(r.Context())
	if err != nil {
		h.log.Error("there was an error retrieving category distribution", "err", err)
		models.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "Failed to retrieve category distribution",
		})
		return
	}

	models.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    data,
	})
}

func (h *Handler) HandleGetClusterStats(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetClusterStats(r.Context())
	if err != nil {
		h.log.Error("there was an error retrieving cluster stats", "err", err)
		models.WriteJSON(w, http.StatusInternalServerError, models.JSONResponse{
			Success: false,
			Message: "Failed to retrieve cluster stats",
		})
		return
	}

	models.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Data:    data,
	})
}
