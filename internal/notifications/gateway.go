package notifications

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"go-server/internal/auth"
	"go-server/internal/database"
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

func (s *gatewayService) PublishHrEvent(_ context.Context, event HrEvent) {
	s.server.BroadcastToRoom("/notifications", s.roleRoom(string(models.RoleAdmin)), "hr:event", event)
}

func (s *gatewayService) PublishDashboardSyncForUser(_ context.Context, userID, role, reason, courseID string) {
	payload := map[string]any{
		"reason":    reason,
		"course_id": courseID,
		"user_id":   userID,
	}
	s.server.BroadcastToRoom("/notifications", s.userRoom(userID), "dashboard:sync", payload)
	s.server.BroadcastToRoom("/notifications", s.roleRoom(role), "dashboard:sync", payload)
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
	if raw := conn.URL().RawQuery; raw != "" {
		values, err := url.ParseQuery(raw)
		if err == nil {
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
