// Package auth provides authentication services for the application
package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"go-server/internal/database"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWTClaims cause go needs strict shape for the tokens
type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Service Interface defines what the auth package can do
type Service interface {
	Login(ctx context.Context, email, password string) (string, error)
}

type service struct {
	db        database.Service // DI of db into  the auth module
	log       *slog.Logger     // DI of logger
	jwtSecret []byte
}

func NewService(db database.Service, secret string, logger *slog.Logger) Service {
	return &service{
		db:        db,
		log:       logger,
		jwtSecret: []byte(secret),
	}
}

func (s *service) Login(ctx context.Context, email, password string) (string, error) {
	// 1. Get user from the db
	user, err := s.db.Read().GetUserByEmail(ctx, email)
	if err != nil {
		s.log.Debug("failed login attempt: user not found",
			slog.String("email", email),
		)
		return "", errors.New("invalid credentials")
	}

	// 2. Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		s.log.Debug("incorrect password entered",
			slog.String("email", email),
		)
		return "", errors.New("invalid credentials")
	}

	// 3. Generate JWT
	claims := JWTClaims{
		UserID: user.ID.String(),
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 1 day session
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	s.log.Debug("user logged in successfully",
		slog.String("email", email),
		slog.String("userID", user.ID.String()),
		slog.String("role", string(user.Role)),
	)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
