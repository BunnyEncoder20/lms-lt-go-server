// Package auth provides authentication services for the application
package auth

import (
	"context"
	"errors"
	"log"
	"time"

	"go-server/internal/database"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWT Claims
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
	db        database.Service
	jwtSecret []byte
}

func NewService(db database.Service, secret string) Service {
	return &service{
		db:        db,
		jwtSecret: []byte(secret),
	}
}

func (s *service) Login(ctx context.Context, email, password string) (string, error) {
	// 1. Get user from the db
	user, err := s.db.Read().GetUserByEmail(ctx, email)
	if err != nil {
		return "", errors.New("Invalid credentials")
	}
	log.Println(user)

	// 2. Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("Invalid credentials")
	}

	// 3. Generate JWT
	claims := JWTClaims{
		UserID: string(user.ID),
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiredAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour())), // 1 day sessoin
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
}
