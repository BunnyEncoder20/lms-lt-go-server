package notifications

import (
	"context"
	"log/slog"
)

type HrEvent struct {
	Type    string         `json:"type"`
	Title   string         `json:"title"`
	Message string         `json:"message"`
	Payload map[string]any `json:"payload"`
}

type Service interface {
	PublishHrEvent(ctx context.Context, event HrEvent)
	PublishDashboardSyncForUser(ctx context.Context, userID, role, reason, courseID string)
	Handler() any
}

type noopService struct {
	log *slog.Logger
}

func NewNoopService(logger *slog.Logger) Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &noopService{log: logger}
}

func (s *noopService) PublishHrEvent(_ context.Context, event HrEvent) {
	s.log.Info("notification noop hr event", "type", event.Type)
}

func (s *noopService) PublishDashboardSyncForUser(_ context.Context, userID, role, reason, courseID string) {
	s.log.Info("notification noop dashboard sync", "user_id", userID, "role", role, "reason", reason, "course_id", courseID)
}

func (s *noopService) Handler() any {
	return nil
}
