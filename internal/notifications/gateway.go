package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"go-server/internal/auth"
	"go-server/internal/database"
	"go-server/internal/database/db"
	"go-server/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	socketio "github.com/googollee/go-socket.io"
)

type gatewayService struct {
	db     database.Service
	log    *slog.Logger
	server *socketio.Server
}

func NewGatewayService(db database.Service, logger *slog.Logger) Service {
	if logger == nil {
		logger = slog.Default()
	}

	server := socketio.NewServer(nil)
	svc := &gatewayService{db: db, log: logger, server: server}
	svc.registerHandlers()
	go func() {
		if err := server.Serve(); err != nil {
			logger.Error("socket.io serve error", "error", err)
		}
	}()
	return svc
}

func (s *gatewayService) registerHandlers() {
	namespace := "/notifications"

	s.server.OnConnect(namespace, func(conn socketio.Conn) error {
		user, err := s.authenticate(conn)
		if err != nil {
			s.log.Warn("notification socket auth failed", "error", err)
			conn.Close()
			return nil
		}

		conn.SetContext(user)
		conn.Join(s.roleRoom(string(user.Role)))
		conn.Join(s.userRoom(user.ID.String()))

		s.log.Debug("notification socket connected", "user_id", user.ID.String(), "conn_id", conn.ID())
		return nil
	})

	s.server.OnDisconnect(namespace, func(conn socketio.Conn, reason string) {
		ctx := conn.Context()
		if user, ok := ctx.(dbUser); ok {
			s.log.Debug("notification socket disconnected", "user_id", user.ID.String(), "conn_id", conn.ID(), "reason", reason)
		}
	})

	s.server.OnError(namespace, func(conn socketio.Conn, err error) {
		s.log.Error("notification socket error", "error", err)
	})
}

func (s *gatewayService) Handler() any {
	return s.server
}

func (s *gatewayService) PublishHrEvent(ctx context.Context, event HrEvent) (models.HrNotificationEventResponse, error) {
	saved, err := s.SaveHrEvent(ctx, event)
	if err != nil {
		return models.HrNotificationEventResponse{}, err
	}

	s.server.BroadcastToRoom("/notifications", s.roleRoom(string(models.RoleAdmin)), "hr.notification", saved)
	return saved, nil
}

func (s *gatewayService) SaveHrEvent(ctx context.Context, event HrEvent) (models.HrNotificationEventResponse, error) {
	payload := event.Payload
	if payload == nil {
		payload = map[string]any{}
	}

	payloadRaw, err := json.Marshal(payload)
	if err != nil {
		return models.HrNotificationEventResponse{}, err
	}

	row, err := s.db.Write().CreateHrNotification(ctx, db.CreateHrNotificationParams{
		ID:        uuid.New(),
		Type:      event.Type,
		Title:     event.Title,
		Message:   event.Message,
		Payload:   string(payloadRaw),
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		return models.HrNotificationEventResponse{}, err
	}

	return mapHrNotificationRow(row), nil
}

func (s *gatewayService) GetHrFeed(ctx context.Context, limit int) ([]models.HrNotificationEventResponse, error) {
	safeLimit := int64(30)
	if limit > 0 {
		safeLimit = int64(limit)
	}
	if safeLimit < 1 {
		safeLimit = 1
	}
	if safeLimit > 100 {
		safeLimit = 100
	}

	rows, err := s.db.Read().ListHrNotifications(ctx, safeLimit)
	if err != nil {
		return nil, err
	}

	items := make([]models.HrNotificationEventResponse, len(rows))
	for i, row := range rows {
		items[i] = mapHrNotificationRow(row)
	}

	return items, nil
}

func (s *gatewayService) PublishDashboardSyncForUser(_ context.Context, userID string, scope DashboardScope, reason string, entityID *string) {
	payload := map[string]any{
		"scope":    scope,
		"reason":   reason,
		"issuedAt": time.Now().UTC().Format(time.RFC3339),
	}
	if entityID != nil && *entityID != "" {
		payload["entityId"] = *entityID
	}

	s.server.BroadcastToRoom("/notifications", s.userRoom(userID), "dashboard.sync", payload)
}

func (s *gatewayService) PublishDashboardSyncForRole(_ context.Context, role DashboardScope, reason string, entityID *string) {
	payload := map[string]any{
		"scope":    role,
		"reason":   reason,
		"issuedAt": time.Now().UTC().Format(time.RFC3339),
	}
	if entityID != nil && *entityID != "" {
		payload["entityId"] = *entityID
	}

	s.server.BroadcastToRoom("/notifications", s.roleRoom(string(role)), "dashboard.sync", payload)
}

func (s *gatewayService) PublishUserEvent(_ context.Context, userID, eventName string, payload any) {
	s.server.BroadcastToRoom("/notifications", s.userRoom(userID), eventName, payload)
}

type dbUser struct {
	ID        uuid.UUID
	Role      models.Role
	FirstName string
	LastName  string
}

func (s *gatewayService) authenticate(conn socketio.Conn) (dbUser, error) {
	token := s.extractToken(conn)
	if token == "" {
		return dbUser{}, errors.New("missing token")
	}

	claims := &auth.JWTClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !parsed.Valid {
		return dbUser{}, errors.New("invalid token")
	}

	if claims.UserID == "" {
		return dbUser{}, errors.New("missing subject")
	}

	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		return dbUser{}, err
	}

	user, err := s.db.Read().GetUserByID(context.Background(), uid)
	if err != nil {
		return dbUser{}, err
	}

	return dbUser{
		ID:        user.ID,
		Role:      user.Role,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, nil
}

func (s *gatewayService) extractToken(conn socketio.Conn) string {
	if val := strings.TrimSpace(conn.RemoteHeader().Get("X-Auth-Token")); val != "" {
		return val
	}

	if raw := conn.URL().RawQuery; raw != "" {
		values, err := url.ParseQuery(raw)
		if err == nil {
			if val := strings.TrimSpace(values.Get("auth[token]")); val != "" {
				return val
			}
			if val := strings.TrimSpace(values.Get("authToken")); val != "" {
				return val
			}
			if val := strings.TrimSpace(values.Get("token")); val != "" {
				return val
			}
		}
	}

	header := conn.RemoteHeader().Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	}

	request := &http.Request{Header: conn.RemoteHeader()}
	if cookie, err := request.Cookie("access-token"); err == nil {
		return strings.TrimSpace(cookie.Value)
	}

	return ""
}

func (s *gatewayService) roleRoom(role string) string {
	return "role:" + role
}

func (s *gatewayService) userRoom(userID string) string {
	return "user:" + userID
}

func AsHTTPHandler(svc Service) http.Handler {
	h, ok := svc.Handler().(http.Handler)
	if !ok {
		return http.NotFoundHandler()
	}
	return h
}

func mapHrNotificationRow(row db.HrNotification) models.HrNotificationEventResponse {
	payload := map[string]any{}
	if row.Payload != "" {
		_ = json.Unmarshal([]byte(row.Payload), &payload)
	}

	return models.HrNotificationEventResponse{
		ID:        row.ID.String(),
		Type:      row.Type,
		Title:     row.Title,
		Message:   row.Message,
		CreatedAt: row.CreatedAt.UTC().Format(time.RFC3339),
		Payload:   payload,
	}
}
