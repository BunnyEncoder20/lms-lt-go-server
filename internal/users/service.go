package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"go-server/internal/database"
	"go-server/internal/database/db"
	"go-server/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	FindAll(ctx context.Context) ([]models.UserResponse, error)
	FindOne(ctx context.Context, userID string) (models.UserResponse, error)
	GetMyTeam(ctx context.Context, managerID string) ([]models.UserResponse, error)
	Create(ctx context.Context, userData models.CreateUserRequest) (models.UserResponse, error)
	Update(ctx context.Context, userID string, userData models.UpdateUserRequest) (models.UserResponse, error)
	DeactivateUser(ctx context.Context, userID string, isActive bool) error
	PermanentlyDeleteUser(ctx context.Context, userID string) error
}

type service struct {
	db database.Service
}

func NewService(db database.Service, logger *slog.Logger) Service {
	return &service{
		db: db,
	}
}

func (s *service) FindAll(ctx context.Context) ([]models.UserResponse, error) {
	users, err := s.db.Read().ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]models.UserResponse, len(users))
	for i, u := range users {
		responses[i] = MapUserToResponse(u)
	}
	return responses, nil
}

func (s *service) FindOne(ctx context.Context, userID string) (models.UserResponse, error) {
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return models.UserResponse{}, err
	}

	user, err := s.db.Read().GetUserByID(ctx, parsedID)
	if err != nil {
		return models.UserResponse{}, err
	}

	return MapUserToResponse(user), nil
}

func (s *service) Create(ctx context.Context, userData models.CreateUserRequest) (models.UserResponse, error) {
	// 1. hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.UserResponse{}, err
	}

	// 2. Prepare the params for sqlc
	params := db.CreateUserParams{
		ID:           uuid.New(),
		PesNumber:    userData.PesNumber,
		Password:     string(hashedPassword),
		FirstName:    userData.FirstName,
		LastName:     userData.LastName,
		Email:        userData.Email,
		Role:         models.Role(userData.Role),
		Title:        userData.Title,
		Gender:       userData.Gender,
		Band:         userData.Band,
		Grade:        userData.Grade,
		Ic:           userData.Ic,
		Sbg:          userData.Sbg,
		Bu:           userData.Bu,
		Segment:      userData.Segment,
		Department:   userData.Department,
		BaseLocation: userData.BaseLocation,
	}

	// 3. Taking care of nullable
	if userData.Cluster != nil {
		params.Cluster = sql.NullString{String: *userData.Cluster, Valid: true}
	}

	if userData.IsID != nil {
		id, _ := uuid.Parse(*userData.IsID)
		params.IsID = uuid.NullUUID{UUID: id, Valid: true}
	}

	if userData.NsID != nil {
		id, _ := uuid.Parse(*userData.NsID)
		params.NsID = uuid.NullUUID{UUID: id, Valid: true}
	}

	if userData.DhID != nil {
		id, _ := uuid.Parse(*userData.DhID)
		params.DhID = uuid.NullUUID{UUID: id, Valid: true}
	}

	user, err := s.db.Write().CreateUser(ctx, params)
	if err != nil {
		return models.UserResponse{}, err
	}
	return MapUserToResponse(user), nil
}

func (s *service) GetMyTeam(ctx context.Context, managerID string) ([]models.UserResponse, error) {
	parsedID, err := uuid.Parse(managerID)
	if err != nil {
		return nil, errors.New("invalid manager id format")
	}

	// wrapping the uuid in a NullUUID so that sqlc type is satisfied
	managerNullUUID := uuid.NullUUID{
		UUID:  parsedID,
		Valid: true,
	}

	// Pass the NullUUID field to the sqlc query
	users, err := s.db.Read().GetTeamMembers(ctx, managerNullUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch team mumbers from database: %w", err)
	}

	// Map the users to the common usres struct
	responses := make([]models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = MapUserToResponse(user)
	}

	return responses, nil
}

func (s *service) DeactivateUser(ctx context.Context, userID string, isActive bool) error {
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	return s.db.Write().UpdateUserStatus(ctx, db.UpdateUserStatusParams{
		ID:       parsedID,
		IsActive: isActive,
	})
}

func (s *service) Update(ctx context.Context, userID string, userData models.UpdateUserRequest) (models.UserResponse, error) {
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return models.UserResponse{}, errors.New("invalid user id format")
	}

	// Build the sqlc params struct
	params := db.UpdateUserParams{
		ID:           parsedID,
		FirstName:    toNullString(userData.FirstName),
		LastName:     toNullString(userData.LastName),
		Title:        toNullString(userData.Title),
		Department:   toNullString(userData.Department),
		BaseLocation: toNullString(userData.BaseLocation),
	}

	updatedUser, err := s.db.Write().UpdateUser(ctx, params)
	if err != nil {
		return models.UserResponse{}, fmt.Errorf("failed to update user in database: %w", err)
	}

	results := MapUserToResponse(updatedUser)

	return results, nil
}

func (s *service) PermanentlyDeleteUser(ctx context.Context, userID string) error {
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("error deleting user from database")
	}

	// Permanently delete the user record from the database
	return s.db.Write().DeleteUser(ctx, parsedID)
}

// MapUserToResponse converts db.User to models.UserResponse
func MapUserToResponse(u db.User) models.UserResponse {
	resp := models.UserResponse{
		ID:           u.ID.String(),
		PesNumber:    u.PesNumber,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Email:        u.Email,
		Role:         u.Role,
		Title:        u.Title,
		Gender:       u.Gender,
		Band:         u.Band,
		Grade:        u.Grade,
		Ic:           u.Ic,
		Sbg:          u.Sbg,
		Bu:           u.Bu,
		Segment:      u.Segment,
		Department:   u.Department,
		BaseLocation: u.BaseLocation,
	}

	if u.Cluster.Valid {
		resp.Cluster = &u.Cluster.String
	}
	if u.IsID.Valid {
		id := u.IsID.UUID.String()
		resp.IsID = &id
	}
	if u.NsID.Valid {
		id := u.NsID.UUID.String()
		resp.NsID = &id
	}
	if u.DhID.Valid {
		id := u.DhID.UUID.String()
		resp.DhID = &id
	}

	return resp
}

func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false} // Tells coalesce to ignore this field
	}
	return sql.NullString{String: *s, Valid: true} // Use the provided value to update
}
