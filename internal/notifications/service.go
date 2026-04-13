package notifications

import (
	"context"
	"log/slog"

	"go-server/internal/models"
)

type HrEvent struct {
	Type    models.HrNotificationType `json:"type"`
	Title   string                    `json:"title"`
	Message string                    `json:"message"`
	Payload map[string]any            `json:"payload"`
}

type DashboardScope string

const (
	DashboardScopeManager  DashboardScope = "MANAGER"
	DashboardScopeEmployee DashboardScope = "EMPLOYEE"
)

type Service interface {
	PublishHrEvent(ctx context.Context, event HrEvent) (models.HrNotificationEventResponse, error)
	SaveHrEvent(ctx context.Context, event HrEvent) (models.HrNotificationEventResponse, error)
	GetHrFeed(ctx context.Context, limit int) ([]models.HrNotificationEventResponse, error)
	PublishDashboardSyncForUser(ctx context.Context, userID string, scope DashboardScope, reason string, entityID *string)
	PublishDashboardSyncForRole(ctx context.Context, role DashboardScope, reason string, entityID *string)
	PublishUserEvent(ctx context.Context, userID, eventName string, payload any)
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

func (s *noopService) PublishHrEvent(_ context.Context, event HrEvent) (models.HrNotificationEventResponse, error) {
	s.log.Info("notification noop hr event", "type", event.Type)
	return models.HrNotificationEventResponse{}, nil
}

func (s *noopService) SaveHrEvent(_ context.Context, event HrEvent) (models.HrNotificationEventResponse, error) {
	s.log.Info("notification noop save hr event", "type", event.Type)
	return models.HrNotificationEventResponse{}, nil
}

func (s *noopService) GetHrFeed(_ context.Context, _ int) ([]models.HrNotificationEventResponse, error) {
	return []models.HrNotificationEventResponse{}, nil
}

func (s *noopService) PublishDashboardSyncForUser(_ context.Context, userID string, scope DashboardScope, reason string, entityID *string) {
	s.log.Info("notification noop dashboard sync user", "user_id", userID, "scope", scope, "reason", reason, "entity_id", entityID)
}

func (s *noopService) PublishDashboardSyncForRole(_ context.Context, role DashboardScope, reason string, entityID *string) {
	s.log.Info("notification noop dashboard sync role", "role", role, "reason", reason, "entity_id", entityID)
}

func (s *noopService) PublishUserEvent(_ context.Context, userID, eventName string, payload any) {
	s.log.Info("notification noop user event", "user_id", userID, "event_name", eventName, "payload", payload)
}

func (s *noopService) Handler() any {
	return nil
}
