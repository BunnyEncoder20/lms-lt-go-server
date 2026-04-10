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
	Create(ctx context.Context, usersData []models.CreateUserRequest) ([]models.UserResponse, error)
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

func (s *service) Create(ctx context.Context, usersData []models.CreateUserRequest) ([]models.UserResponse, error) {
	var responses []models.UserResponse

	err := s.db.ExecTx(ctx, func(qtx *db.Queries) error {
		for _, userData := range usersData {
			// 1. hash the password
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("hashing password for %s: %w", userData.Email, err)
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
				Title:        sql.NullString{String: userData.Title, Valid: userData.Title != ""},
				Gender:       sql.NullString{String: userData.Gender, Valid: userData.Gender != ""},
				Band:         sql.NullString{String: userData.Band, Valid: userData.Band != ""},
				Grade:        sql.NullString{String: userData.Grade, Valid: userData.Grade != ""},
				Ic:           sql.NullString{String: userData.Ic, Valid: userData.Ic != ""},
				Sbg:          sql.NullString{String: userData.Sbg, Valid: userData.Sbg != ""},
				Bu:           sql.NullString{String: userData.Bu, Valid: userData.Bu != ""},
				Segment:      sql.NullString{String: userData.Segment, Valid: userData.Segment != ""},
				Department:   sql.NullString{String: userData.Department, Valid: userData.Department != ""},
				BaseLocation: sql.NullString{String: userData.BaseLocation, Valid: userData.BaseLocation != ""},
			}

			// 3. Taking care of nullable
			if userData.Cluster != nil {
				params.Cluster = sql.NullString{String: *userData.Cluster, Valid: true}
			}

			if userData.IsID != nil && *userData.IsID != "" {
				id, err := uuid.Parse(*userData.IsID)
				if err != nil {
					return fmt.Errorf("invalid IS ID for %s: %w", userData.Email, err)
				}
				params.IsID = uuid.NullUUID{UUID: id, Valid: true}
			}

			if userData.NsID != nil && *userData.NsID != "" {
				id, err := uuid.Parse(*userData.NsID)
				if err != nil {
					return fmt.Errorf("invalid NS ID for %s: %w", userData.Email, err)
				}
				params.NsID = uuid.NullUUID{UUID: id, Valid: true}
			}

			if userData.DhID != nil && *userData.DhID != "" {
				id, err := uuid.Parse(*userData.DhID)
				if err != nil {
					return fmt.Errorf("invalid DH ID for %s: %w", userData.Email, err)
				}
				params.DhID = uuid.NullUUID{UUID: id, Valid: true}
			}

			// finally, writing to the DB
			user, err := qtx.CreateUser(ctx, params)
			if err != nil {
				return err
			}
			responses = append(responses, MapUserToResponse(user))
		}

		// returining nil tells the trx to commit
		return nil
	})
	if err != nil {
		return nil, err
	}

	return responses, nil
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
		ID:        u.ID.String(),
		PesNumber: u.PesNumber,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Role:      u.Role,
	}

	if u.FullName.Valid {
		resp.FullName = &u.FullName.String
	}
	if u.Cluster.Valid {
		resp.Cluster = &u.Cluster.String
	}
	if u.Location.Valid {
		resp.Location = &u.Location.String
	}
	if u.EmploymentStatus.Valid {
		resp.EmploymentStatus = &u.EmploymentStatus.String
	}
	if u.IsPsn.Valid {
		resp.IsPsn = &u.IsPsn.String
	}
	if u.IsName.Valid {
		resp.IsName = &u.IsName.String
	}
	if u.NsPsn.Valid {
		resp.NsPsn = &u.NsPsn.String
	}
	if u.NsName.Valid {
		resp.NsName = &u.NsName.String
	}
	if u.DhPsn.Valid {
		resp.DhPsn = &u.DhPsn.String
	}
	if u.DhName.Valid {
		resp.DhName = &u.DhName.String
	}
	if u.Ic.Valid {
		resp.Ic = &u.Ic.String
	}
	if u.Sbg.Valid {
		resp.Sbg = &u.Sbg.String
	}
	if u.Bu.Valid {
		resp.Bu = &u.Bu.String
	}
	if u.Segment.Valid {
		resp.Segment = &u.Segment.String
	}
	if u.Department.Valid {
		resp.Department = &u.Department.String
	}
	if u.BaseLocation.Valid {
		resp.BaseLocation = &u.BaseLocation.String
	}

	if u.ManagerID.Valid {
		id := u.ManagerID.UUID.String()
		resp.ManagerID = &id
	}
	if u.SkipManagerID.Valid {
		id := u.SkipManagerID.UUID.String()
		resp.SkipManagerID = &id
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
	if u.Title.Valid {
		resp.Title = &u.Title.String
	}
	if u.Gender.Valid {
		resp.Gender = &u.Title.String
	}
	if u.Band.Valid {
		resp.Band = &u.Band.String
	}
	if u.Grade.Valid {
		resp.Grade = &u.Grade.String
	}

	return resp
}

func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false} // Tells coalesce to ignore this field
	}
	return sql.NullString{String: *s, Valid: true} // Use the provided value to update
}
