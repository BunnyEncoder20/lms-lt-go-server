package admin

import (
	"log/slog"
	"net/http"
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

func (h *Handler) HandleGetKpis(w http.ResponseWriter, r *http.Request)                 {}
func (h *Handler) HandleGetMonthlyStats(w http.ResponseWriter, r *http.Request)         {}
func (h *Handler) HandleGetCategoryDistribution(w http.ResponseWriter, r *http.Request) {}
func (h *Handler) HandleGetClusterStats(w http.ResponseWriter, r *http.Request)         {}
