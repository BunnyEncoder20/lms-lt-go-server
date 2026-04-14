// Package auth contains all the servuces related to authentication
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"

	"go-server/internal/database"
	"go-server/internal/database/db"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type Service interface {
	Login(ctx context.Context, pesNumber, password string) (*TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
}

type service struct {
	db        database.Service
	jwtSecret []byte
	log       *slog.Logger
}

func NewService(db database.Service, secret string, logger *slog.Logger) Service {
	return &service{
		db:        db,
		jwtSecret: []byte(secret),
		log:       logger,
	}
}

const (
	accessTokenExpiry  = 15 * time.Minute
	refreshTokenExpiry = 7 * 24 * time.Hour
)

func (s *service) generateAccessToken(userID, role string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func generateRefreshToken() (plain string, hashed string, err error) {
	bytes := make([]byte, 32)
	if _, err = rand.Read(bytes); err != nil {
		return "", "", err
	}
	plain = hex.EncodeToString(bytes)    // plain refreshToken sent to client
	hash := sha256.Sum256([]byte(plain)) // hashed
	hashed = hex.EncodeToString(hash[:]) // string of hashed refreshToken to be stored in db
	return plain, hashed, nil
}

func (s *service) Login(ctx context.Context, pesNumber, password string) (*TokenPair, error) {
	user, err := s.db.Read().GetUserByPesNumber(ctx, pesNumber)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	accessToken, err := s.generateAccessToken(user.ID.String(), string(user.Role))
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	plainRefresh, hashedRefresh, err := generateRefreshToken()
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	refreshExpiry := time.Now().Add(refreshTokenExpiry)
	err = s.db.Write().UpdateRefreshToken(ctx, db.UpdateRefreshTokenParams{
		RefreshTokenHash:      sql.NullString{String: hashedRefresh, Valid: true},
		RefreshTokenExpiresAt: sql.NullTime{Time: refreshExpiry, Valid: true},
		ID:                    user.ID,
	})
	if err != nil {
		return nil, errors.New("failed to store refresh token")
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: plainRefresh,
	}, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	hash := sha256.Sum256([]byte(refreshToken))
	hashedToken := hex.EncodeToString(hash[:])

	user, err := s.db.Read().GetUserByRefreshTokenHash(ctx, sql.NullString{String: hashedToken, Valid: true})
	if err != nil {
		s.log.Warn("refresh token reuse attempt or invalid token", "token_hash_prefix", hashedToken[:8])
		return nil, errors.New("invalid or revoked refresh token")
	}

	if !user.RefreshTokenExpiresAt.Valid || user.RefreshTokenExpiresAt.Time.Before(time.Now()) {
		s.log.Warn("expired refresh token used", "user_id", user.ID)
		_ = s.db.Write().RevokeRefreshToken(ctx, user.ID)
		return nil, errors.New("refresh token expired")
	}

	newAccessToken, err := s.generateAccessToken(user.ID.String(), string(user.Role))
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	newPlainRefresh, newHashedRefresh, err := generateRefreshToken()
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	newRefreshExpiry := time.Now().Add(refreshTokenExpiry)
	err = s.db.Write().UpdateRefreshToken(ctx, db.UpdateRefreshTokenParams{
		RefreshTokenHash:      sql.NullString{String: newHashedRefresh, Valid: true},
		RefreshTokenExpiresAt: sql.NullTime{Time: newRefreshExpiry, Valid: true},
		ID:                    user.ID,
	})
	if err != nil {
		return nil, errors.New("failed to rotate refresh token")
	}

	s.log.Info("refresh token rotated", "user_id", user.ID)

	return &TokenPair{
		AccessToken:  newAccessToken,
		RefreshToken: newPlainRefresh,
	}, nil
}
