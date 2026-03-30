// Package trainings contains all the handlers and services related to training management.
package trainings

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"go-server/internal/database"
	"go-server/internal/database/db"
	"go-server/internal/models"

	"github.com/google/uuid"
)

// Service defines the interface for training-related business logic.
type Service interface {
	// List retrieves all trainings
	List(ctx context.Context) ([]models.TrainingResponse, error)
	// Get retrieves a single training by ID
	Get(ctx context.Context, trainingID string) (models.TrainingResponse, error)
	// GetByCategory retrieves trainings filtered by category
	GetByCategory(ctx context.Context, category string) ([]models.TrainingResponse, error)
	// GetUpcoming retrieves trainings that haven't started yet
	GetUpcoming(ctx context.Context) ([]models.TrainingResponse, error)
	// GetEmployeeTrainings retrieves all trainings for a specific employee
	GetEmployeeTrainings(ctx context.Context, userID string) ([]models.NominationResponse, error)
	// Create creates a new training
	Create(ctx context.Context, createdByID string, req models.CreateTrainingRequest) (models.TrainingResponse, error)
	// Update updates an existing training
	Update(ctx context.Context, trainingID string, req models.UpdateTrainingRequest) (models.TrainingResponse, error)
	// Delete permanently deletes a training
	Delete(ctx context.Context, trainingID string) error
}

type service struct {
	db database.Service
}

// NewService a constructor creates a new training service instance.
func NewService(db database.Service) Service {
	return &service{
		db: db,
	}
}

// List retrieves all trainings from the database.
func (s *service) List(ctx context.Context) ([]models.TrainingResponse, error) {
	trainings, err := s.db.Read().ListTrainings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list trainings: %w", err)
	}

	responses := make([]models.TrainingResponse, len(trainings))
	for i, t := range trainings {
		responses[i] = MapTrainingToResponse(t)
	}
	return responses, nil
}

// Get retrieves a single training by its ID.
func (s *service) Get(ctx context.Context, trainingID string) (models.TrainingResponse, error) {
	parsedID, err := uuid.Parse(trainingID)
	if err != nil {
		return models.TrainingResponse{}, errors.New("invalid training ID format")
	}

	training, err := s.db.Read().GetTrainingByID(ctx, parsedID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.TrainingResponse{}, errors.New("training not found")
		}
		return models.TrainingResponse{}, fmt.Errorf("failed to get training: %w", err)
	}

	return MapTrainingToResponse(training), nil
}

// GetByCategory retrieves trainings filtered by category.
func (s *service) GetByCategory(ctx context.Context, category string) ([]models.TrainingResponse, error) {
	// Validate category
	trainingCategory := models.TrainingCategory(category)
	if trainingCategory != models.TrainingTechnical && trainingCategory != models.TrainingBehavioral {
		return nil, errors.New("invalid training category")
	}

	trainings, err := s.db.Read().ListTrainingsByCategory(ctx, trainingCategory)
	if err != nil {
		return nil, fmt.Errorf("failed to list trainings by category: %w", err)
	}

	responses := make([]models.TrainingResponse, len(trainings))
	for i, t := range trainings {
		responses[i] = MapTrainingToResponse(t)
	}
	return responses, nil
}

// GetUpcoming retrieves trainings that haven't started yet.
func (s *service) GetUpcoming(ctx context.Context) ([]models.TrainingResponse, error) {
	trainings, err := s.db.Read().ListUpcomingTrainings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list upcoming trainings: %w", err)
	}

	responses := make([]models.TrainingResponse, len(trainings))
	for i, t := range trainings {
		responses[i] = MapTrainingToResponse(t)
	}
	return responses, nil
}

// GetEmployeeTrainings retrieves all trainings for a specific employee.
func (s *service) GetEmployeeTrainings(ctx context.Context, userID string) ([]models.NominationResponse, error) {
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	nominations, err := s.db.Read().GetNominationsByUserID(ctx, parsedID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee trainings: %w", err)
	}

	responses := make([]models.NominationResponse, len(nominations))
	for i, n := range nominations {
		responses[i] = MapNominationToResponse(n)
	}
	return responses, nil
}

// Create creates a new training.
func (s *service) Create(ctx context.Context, createdByID string, req models.CreateTrainingRequest) (models.TrainingResponse, error) {
	// Parse UUIDs
	creatorID, err := uuid.Parse(createdByID)
	if err != nil {
		return models.TrainingResponse{}, errors.New("invalid creator ID format")
	}

	hrProgramID, err := uuid.Parse(req.HrProgramID)
	if err != nil {
		return models.TrainingResponse{}, errors.New("invalid HR program ID format")
	}

	facilityID, err := uuid.Parse(req.FacilityID)
	if err != nil {
		return models.TrainingResponse{}, errors.New("invalid facility ID format")
	}

	// Parse dates
	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		return models.TrainingResponse{}, errors.New("invalid start date format, expected RFC3339")
	}

	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		return models.TrainingResponse{}, errors.New("invalid end date format, expected RFC3339")
	}

	// Validate category
	category := models.TrainingCategory(req.Category)
	if category != models.TrainingTechnical && category != models.TrainingBehavioral {
		return models.TrainingResponse{}, errors.New("invalid training category")
	}

	// Validate delivery mode
	modeOfDelivery := models.DeliveryMode(req.ModeOfDelivery)
	validModes := []models.DeliveryMode{models.InPerson, models.VirtualLink, models.Hybrid, models.Elearning}
	isValidMode := slices.Contains(validModes, modeOfDelivery)
	if !isValidMode {
		return models.TrainingResponse{}, errors.New("invalid mode of delivery")
	}

	// Build params for database query
	params := db.CreateTrainingParams{
		ID:             uuid.New(),
		Title:          req.Title,
		Category:       category,
		StartDate:      startDate,
		EndDate:        endDate,
		CreatedByID:    creatorID,
		DeadlineDays:   req.DeadlineDays,
		HrProgramID:    hrProgramID,
		MappedCategory: req.MappedCategory,
		ModeOfDelivery: modeOfDelivery,
		InstructorName: req.InstructorName,
		FacilityID:     facilityID,
	}

	// Handle optional fields
	if req.Description != nil {
		params.Description = sql.NullString{String: *req.Description, Valid: true}
	}
	if req.Location != nil {
		params.Location = sql.NullString{String: *req.Location, Valid: true}
	}
	if req.VirtualLink != nil {
		params.VirtualLink = sql.NullString{String: *req.VirtualLink, Valid: true}
	}
	if req.PreReadURI != nil {
		params.PreReadUri = sql.NullString{String: *req.PreReadURI, Valid: true}
	}
	if req.InstitutePartnerName != nil {
		params.InstitutePartnerName = sql.NullString{String: *req.InstitutePartnerName, Valid: true}
	}
	if req.ProcessOwnerName != nil {
		params.ProcessOwnerName = sql.NullString{String: *req.ProcessOwnerName, Valid: true}
	}
	if req.ProcessOwnerEmail != nil {
		params.ProcessOwnerEmail = sql.NullString{String: *req.ProcessOwnerEmail, Valid: true}
	}
	if req.DurationManhours != nil {
		params.DurationManhours = sql.NullFloat64{Float64: *req.DurationManhours, Valid: true}
	}
	if req.TrainingMandays != nil {
		params.TrainingMandays = sql.NullFloat64{Float64: *req.TrainingMandays, Valid: true}
	}

	training, err := s.db.Write().CreateTraining(ctx, params)
	if err != nil {
		return models.TrainingResponse{}, fmt.Errorf("failed to create training: %w", err)
	}

	return MapTrainingToResponse(training), nil
}

// Update updates an existing training.
func (s *service) Update(ctx context.Context, trainingID string, req models.UpdateTrainingRequest) (models.TrainingResponse, error) {
	parsedID, err := uuid.Parse(trainingID)
	if err != nil {
		return models.TrainingResponse{}, errors.New("invalid training ID format")
	}

	// Build params for database query
	params := db.UpdateTrainingParams{
		ID: parsedID,
	}

	// Handle optional fields
	if req.Title != nil {
		params.Title = sql.NullString{String: *req.Title, Valid: true}
	}
	if req.Description != nil {
		params.Description = sql.NullString{String: *req.Description, Valid: true}
	}
	if req.Category != nil {
		category := models.TrainingCategory(*req.Category)
		if category != models.TrainingTechnical && category != models.TrainingBehavioral {
			return models.TrainingResponse{}, errors.New("invalid training category")
		}
		// We need to set the category as a valid enum
		params.Category = category
	}
	if req.StartDate != nil {
		startDate, err := time.Parse(time.RFC3339, *req.StartDate)
		if err != nil {
			return models.TrainingResponse{}, errors.New("invalid start date format, expected RFC3339")
		}
		params.StartDate = sql.NullTime{Time: startDate, Valid: true}
	}
	if req.EndDate != nil {
		endDate, err := time.Parse(time.RFC3339, *req.EndDate)
		if err != nil {
			return models.TrainingResponse{}, errors.New("invalid end date format, expected RFC3339")
		}
		params.EndDate = sql.NullTime{Time: endDate, Valid: true}
	}
	if req.Location != nil {
		params.Location = sql.NullString{String: *req.Location, Valid: true}
	}
	if req.VirtualLink != nil {
		params.VirtualLink = sql.NullString{String: *req.VirtualLink, Valid: true}
	}
	if req.PreReadURI != nil {
		params.PreReadUri = sql.NullString{String: *req.PreReadURI, Valid: true}
	}
	if req.DeadlineDays != nil {
		params.DeadlineDays = sql.NullInt64{Int64: *req.DeadlineDays, Valid: true}
	}
	if req.MappedCategory != nil {
		params.MappedCategory = sql.NullString{String: *req.MappedCategory, Valid: true}
	}
	if req.ModeOfDelivery != nil {
		modeOfDelivery := models.DeliveryMode(*req.ModeOfDelivery)
		validModes := []models.DeliveryMode{models.InPerson, models.VirtualLink, models.Hybrid, models.Elearning}
		isValidMode := slices.Contains(validModes, modeOfDelivery)
		if !isValidMode {
			return models.TrainingResponse{}, errors.New("invalid mode of delivery")
		}
		params.ModeOfDelivery = modeOfDelivery
	}
	if req.InstructorName != nil {
		params.InstructorName = sql.NullString{String: *req.InstructorName, Valid: true}
	}
	if req.InstitutePartnerName != nil {
		params.InstitutePartnerName = sql.NullString{String: *req.InstitutePartnerName, Valid: true}
	}
	if req.ProcessOwnerName != nil {
		params.ProcessOwnerName = sql.NullString{String: *req.ProcessOwnerName, Valid: true}
	}
	if req.ProcessOwnerEmail != nil {
		params.ProcessOwnerEmail = sql.NullString{String: *req.ProcessOwnerEmail, Valid: true}
	}
	if req.DurationManhours != nil {
		params.DurationManhours = sql.NullFloat64{Float64: *req.DurationManhours, Valid: true}
	}
	if req.TrainingMandays != nil {
		params.TrainingMandays = sql.NullFloat64{Float64: *req.TrainingMandays, Valid: true}
	}
	if req.IsActive != nil {
		params.IsActive = sql.NullBool{Bool: *req.IsActive, Valid: true}
	}

	training, err := s.db.Write().UpdateTraining(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.TrainingResponse{}, errors.New("training not found")
		}
		return models.TrainingResponse{}, fmt.Errorf("failed to update training: %w", err)
	}

	return MapTrainingToResponse(training), nil
}

// Delete permanently deletes a training from the database.
func (s *service) Delete(ctx context.Context, trainingID string) error {
	parsedID, err := uuid.Parse(trainingID)
	if err != nil {
		return errors.New("invalid training ID format")
	}

	err = s.db.Write().DeleteTraining(ctx, parsedID)
	if err != nil {
		return fmt.Errorf("failed to delete training: %w", err)
	}

	return nil
}

// MapTrainingToResponse converts db.Training to models.TrainingResponse.
func MapTrainingToResponse(t db.Training) models.TrainingResponse {
	resp := models.TrainingResponse{
		ID:             t.ID.String(),
		Title:          t.Title,
		Category:       t.Category,
		StartDate:      t.StartDate.Format(time.RFC3339),
		EndDate:        t.EndDate.Format(time.RFC3339),
		CreatedByID:    t.CreatedByID.String(),
		DeadlineDays:   t.DeadlineDays,
		HrProgramID:    t.HrProgramID.String(),
		MappedCategory: t.MappedCategory,
		ModeOfDelivery: t.ModeOfDelivery,
		InstructorName: t.InstructorName,
		FacilityID:     t.FacilityID.String(),
		IsActive:       t.IsActive,
		CreatedAt:      t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      t.UpdatedAt.Format(time.RFC3339),
	}

	if t.Description.Valid {
		resp.Description = &t.Description.String
	}
	if t.Location.Valid {
		resp.Location = &t.Location.String
	}
	if t.VirtualLink.Valid {
		resp.VirtualLink = &t.VirtualLink.String
	}
	if t.PreReadUri.Valid {
		resp.PreReadURI = &t.PreReadUri.String
	}
	if t.InstitutePartnerName.Valid {
		resp.InstitutePartnerName = &t.InstitutePartnerName.String
	}
	if t.ProcessOwnerName.Valid {
		resp.ProcessOwnerName = &t.ProcessOwnerName.String
	}
	if t.ProcessOwnerEmail.Valid {
		resp.ProcessOwnerEmail = &t.ProcessOwnerEmail.String
	}
	if t.DurationManhours.Valid {
		resp.DurationManhours = &t.DurationManhours.Float64
	}
	if t.TrainingMandays.Valid {
		resp.TrainingMandays = &t.TrainingMandays.Float64
	}

	return resp
}

// MapNominationToResponse converts db.Nomination to models.NominationResponse.
func MapNominationToResponse(n db.Nomination) models.NominationResponse {
	resp := models.NominationResponse{
		ID:            n.ID.String(),
		Status:        n.Status,
		UserID:        n.UserID.String(),
		TrainingID:    n.TrainingID.String(),
		NominatedByID: n.NominatedByID.String(),
		CreatedAt:     n.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     n.UpdatedAt.Format(time.RFC3339),
	}

	if n.HrCompletionStatus.Valid {
		resp.HrCompletionStatus = &n.HrCompletionStatus.String
	}
	if n.ProfFees.Valid {
		resp.ProfFees = &n.ProfFees.Float64
	}
	if n.VenueCost.Valid {
		resp.VenueCost = &n.VenueCost.Float64
	}
	if n.OtherCost.Valid {
		resp.OtherCost = &n.OtherCost.Float64
	}
	if n.NonTemsTravel.Valid {
		resp.NonTemsTravel = &n.NonTemsTravel.Float64
	}
	if n.NonTemsAccommodation.Valid {
		resp.NonTemsAccommodation = &n.NonTemsAccommodation.Float64
	}
	if n.TotalCost.Valid {
		resp.TotalCost = &n.TotalCost.Float64
	}

	return resp
}
