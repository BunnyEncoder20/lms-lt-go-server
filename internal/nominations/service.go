// Package nominations contains all the handlers and services related to nomination management.
package nominations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go-server/internal/database"
	"go-server/internal/database/db"
	"go-server/internal/models"

	"github.com/google/uuid"
)

// Service defines the interface for nomination-related business logic.
type Service interface {
	// Manager: Nominate employees for training
	NominateEmployees(ctx context.Context, managerID string, req models.CreateNominationRequest) ([]models.NominationResponse, error)
	// Employee: Self-nominate for training
	SelfNomination(ctx context.Context, userID string, req models.SelfNominationRequest) (models.NominationResponse, error)
	// Employee: respond to nomination (accept/decline)
	RespondToNomination(ctx context.Context, employeeID string, nominationID string, req models.NominationResponseRequest) (models.NominationResponse, error)
	// Manager: Approve/Reject self-nomination
	RespondToSelfNomination(ctx context.Context, managerID string, nominationID string, status models.NominationStatus) (models.NominationResponse, error)
	// Employee: my nominations
	GetMyNominations(ctx context.Context, userID string) ([]models.NominationResponse, error)
	// Manager: view team nominations
	GetTeamNominations(ctx context.Context, managerID string) ([]models.NominationResponse, error)
	// Admin: view all nominations with filters
	GetAllNominations(ctx context.Context, filters models.NominationFilters) ([]models.NominationResponse, error)
	// Admin: override status
	UpdateNominationStatus(ctx context.Context, nominationID string, status models.NominationStatus) (models.NominationResponse, error)
	// Manager: dashboard KPIs
	GetManagerDashboard(ctx context.Context, managerID string) (models.ManagerDashboardResponse, error)
	// Employee: dashboard KPIs
	GetEmployeeDashboard(ctx context.Context, employeeID string) (models.EmployeeDashboardResponse, error)
	// Get all published/upcoming courses/trainings
	GetAllPublishedCourses(ctx context.Context) ([]models.TrainingResponse, error)
}

type service struct {
	db database.Service
}

// NewService creates a new nomination service instance.
func NewService(db database.Service) Service {
	return &service{
		db: db,
	}
}

// NominateEmployees allows a manager to nominate employees for training.
func (s *service) NominateEmployees(ctx context.Context, managerID string, req models.CreateNominationRequest) ([]models.NominationResponse, error) {
	// Parse manager ID
	parsedManagerID, err := uuid.Parse(managerID)
	if err != nil {
		return nil, errors.New("invalid manager ID format")
	}

	// Parse training ID
	parsedTrainingID, err := uuid.Parse(req.TrainingID)
	if err != nil {
		return nil, errors.New("invalid training ID format")
	}

	// Verify training exists
	_, err = s.db.Read().GetTrainingByID(ctx, parsedTrainingID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("training not found")
		}
		return nil, fmt.Errorf("failed to get training: %w", err)
	}

	// We declare the response slice outside the transactions block
	responses := make([]models.NominationResponse, 0, len(req.UserIDs))

	// Opening the transactions block
	err = s.db.ExecTx(ctx, func(qtx *db.Queries) error {
		// everthing in here is running inside the transaction
		for _, userID := range req.UserIDs {
			parsedUserID, err := uuid.Parse(userID)
			if err != nil {
				return fmt.Errorf("invalid user ID format: %w", err)
			}

			// Create nomination with PENDING_EMPLOYEE_APPROVAL status
			params := db.CreateNominationParams{
				ID:            uuid.New(),
				Status:        models.NomPendingEmployeeApproval,
				UserID:        parsedUserID,
				TrainingID:    uuid.NullUUID{UUID: parsedTrainingID, Valid: true},
				NominatedByID: parsedManagerID,
			}

			// NOTE: inside the transaction block, we have to use the provided qtx (which implements the same Queries interface) instead of s.db.Read()/s.db.Write()
			nomination, err := qtx.CreateNomination(ctx, params)
			if err != nil {
				// If unique constraint violation, skip this user (already nominated)
				continue
			}

			responses = append(responses, MapNominationToResponse(nomination))
		}

		return nil // returning nil tells ExecTx to commit the transaction
	})
	// 3. check if the entire transaction failed
	if err != nil {
		return nil, fmt.Errorf("transaction failed during nomination creation: %w", err)
	}

	return responses, nil
}

// SelfNomination allows an employee to self-nominate for training.
func (s *service) SelfNomination(ctx context.Context, userID string, req models.SelfNominationRequest) (models.NominationResponse, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return models.NominationResponse{}, errors.New("invalid user ID format")
	}

	parsedTrainingID, err := uuid.Parse(req.TrainingID)
	if err != nil {
		return models.NominationResponse{}, errors.New("invalid training ID format")
	}

	var nomination db.Nomination

	// start of transaction (cause there are 3 reads before the final write, we wrap all of hem in a transaction to ensure data consistency
	err = s.db.ExecTx(ctx, func(qtx *db.Queries) error {
		_, err = qtx.GetTrainingByID(ctx, parsedTrainingID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("training not found")
			}
			return fmt.Errorf("failed to get training: %w", err)
		}

		user, err := qtx.GetUserByID(ctx, parsedUserID)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		if !user.IsID.Valid {
			return errors.New("cannot self-nominate without a manager assigned")
		}

		params := db.CreateNominationParams{
			ID:            uuid.New(),
			Status:        models.NomPendingManagerApproval,
			UserID:        parsedUserID,
			TrainingID:    uuid.NullUUID{UUID: parsedTrainingID, Valid: true},
			NominatedByID: user.IsID.UUID,
		}

		nomination, err = qtx.CreateNomination(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to create nomination: %w", err)
		}

		return nil
	})
	if err != nil {
		return models.NominationResponse{}, err
	}

	return MapNominationToResponse(nomination), nil
}

// RespondToNomination allows an employee to accept or decline a nomination.
func (s *service) RespondToNomination(ctx context.Context, employeeID string, nominationID string, req models.NominationResponseRequest) (models.NominationResponse, error) {
	parsedEmployeeID, err := uuid.Parse(employeeID)
	if err != nil {
		return models.NominationResponse{}, errors.New("invalid employee ID format")
	}

	parsedNominationID, err := uuid.Parse(nominationID)
	if err != nil {
		return models.NominationResponse{}, errors.New("invalid nomination ID format")
	}

	var updatedNomination db.Nomination

	// Start of transaction (cause there are 2 reads before the final write, we wrap all of hem in a transaction to ensure data consistency
	err = s.db.ExecTx(ctx, func(qtx *db.Queries) error {
		nomination, err := qtx.GetNominationByID(ctx, parsedNominationID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("nomination not found")
			}
			return fmt.Errorf("failed to get nomination: %w", err)
		}

		if nomination.UserID != parsedEmployeeID {
			return errors.New("unauthorized: nomination does not belong to this employee")
		}

		var newStatus models.NominationStatus
		switch req.Status {
		case "ACCEPTED":
			newStatus = models.NomEnrolled
		case "DECLINED":
			newStatus = models.NomDeclined
		default:
			return errors.New("invalid status: must be ACCEPTED or DECLINED")
		}

		params := db.UpdateNominationStatusParams{
			Status: newStatus,
			ID:     parsedNominationID,
		}

		updatedNomination, err = qtx.UpdateNominationStatus(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to update nomination status: %w", err)
		}

		return nil
	})
	if err != nil {
		return models.NominationResponse{}, err
	}

	return MapNominationToResponse(updatedNomination), nil
}

// RespondToSelfNomination allows a manager to approve or reject a self-nomination.
func (s *service) RespondToSelfNomination(ctx context.Context, managerID string, nominationID string, status models.NominationStatus) (models.NominationResponse, error) {
	parsedManagerID, err := uuid.Parse(managerID)
	if err != nil {
		return models.NominationResponse{}, errors.New("invalid manager ID format")
	}

	parsedNominationID, err := uuid.Parse(nominationID)
	if err != nil {
		return models.NominationResponse{}, errors.New("invalid nomination ID format")
	}

	if status != models.NomEnrolled && status != models.NomRejected {
		return models.NominationResponse{}, errors.New("invalid status: must be ENROLLED or REJECTED")
	}

	var updatedNomination db.Nomination

	// start of transaction (cause there are 2 reads before the final write, we wrap all of hem in a transaction to ensure data consistency
	err = s.db.ExecTx(ctx, func(qtx *db.Queries) error {
		nomination, err := qtx.GetNominationByID(ctx, parsedNominationID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("nomination not found")
			}
			return fmt.Errorf("failed to get nomination: %w", err)
		}

		if nomination.NominatedByID != parsedManagerID {
			return errors.New("unauthorized: can only respond to nominations from your team")
		}

		if nomination.Status != models.NomPendingManagerApproval {
			return errors.New("can only respond to nominations that are in PENDING status. Current status: " + string(nomination.Status))
		}

		params := db.UpdateNominationStatusParams{
			Status: status,
			ID:     parsedNominationID,
		}

		updatedNomination, err = qtx.UpdateNominationStatus(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to update nomination status: %w", err)
		}

		return nil
	})
	if err != nil {
		return models.NominationResponse{}, err
	}

	return MapNominationToResponse(updatedNomination), nil
}

// GetMyNominations retrieves all nominations for an employee.
func (s *service) GetMyNominations(ctx context.Context, userID string) ([]models.NominationResponse, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	nominations, err := s.db.Read().GetNominationsByUserID(ctx, parsedUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get nominations: %w", err)
	}

	responses := make([]models.NominationResponse, len(nominations))
	for i, n := range nominations {
		responses[i] = MapNominationToResponse(n)
	}

	return responses, nil
}

// GetTeamNominations retrieves all nominations for a manager's team.
func (s *service) GetTeamNominations(ctx context.Context, managerID string) ([]models.NominationResponse, error) {
	parsedManagerID, err := uuid.Parse(managerID)
	if err != nil {
		return nil, errors.New("invalid manager ID format")
	}

	// Wrap in NullUUID for the query
	managerNullUUID := uuid.NullUUID{
		UUID:  parsedManagerID,
		Valid: true,
	}

	nominations, err := s.db.Read().GetNominationsByManagerID(ctx, managerNullUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team nominations: %w", err)
	}

	responses := make([]models.NominationResponse, len(nominations))
	for i, n := range nominations {
		responses[i] = MapNominationToResponse(n)
	}

	return responses, nil
}

// GetAllNominations retrieves all nominations with optional filters (admin only).
func (s *service) GetAllNominations(ctx context.Context, filters models.NominationFilters) ([]models.NominationResponse, error) {
	params := db.ListNominationsByFiltersParams{}

	// Apply filters if provided
	if filters.Status != nil && *filters.Status != "" {
		params.Status = sql.NullString{String: *filters.Status, Valid: true}
	} else {
		params.Status = sql.NullString{Valid: false}
	}

	if filters.TrainingID != nil && *filters.TrainingID != "" {
		params.TrainingID = sql.NullString{String: *filters.TrainingID, Valid: true}
	} else {
		params.TrainingID = sql.NullString{Valid: false}
	}

	if filters.UserID != nil && *filters.UserID != "" {
		params.UserID = sql.NullString{String: *filters.UserID, Valid: true}
	} else {
		params.UserID = sql.NullString{Valid: false}
	}

	if filters.ManagerID != nil && *filters.ManagerID != "" {
		params.NominatedByID = sql.NullString{String: *filters.ManagerID, Valid: true}
	} else {
		params.NominatedByID = sql.NullString{Valid: false}
	}

	nominations, err := s.db.Read().ListNominationsByFilters(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get nominations: %w", err)
	}

	responses := make([]models.NominationResponse, len(nominations))
	for i, n := range nominations {
		responses[i] = MapNominationToResponse(n)
	}

	return responses, nil
}

// UpdateNominationStatus allows admin to override nomination status.
func (s *service) UpdateNominationStatus(ctx context.Context, nominationID string, status models.NominationStatus) (models.NominationResponse, error) {
	parsedNominationID, err := uuid.Parse(nominationID)
	if err != nil {
		return models.NominationResponse{}, errors.New("invalid nomination ID format")
	}

	if !models.IsValidNominationStatus(status) {
		return models.NominationResponse{}, errors.New("invalid nomination status")
	}

	params := db.UpdateNominationStatusParams{
		Status: status,
		ID:     parsedNominationID,
	}

	nomination, err := s.db.Write().UpdateNominationStatus(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.NominationResponse{}, errors.New("nomination not found")
		}
		return models.NominationResponse{}, fmt.Errorf("failed to update nomination status: %w", err)
	}

	return MapNominationToResponse(nomination), nil
}

// GetManagerDashboard retrieves dashboard KPIs for a manager.
func (s *service) GetManagerDashboard(ctx context.Context, managerID string) (models.ManagerDashboardResponse, error) {
	parsedManagerID, err := uuid.Parse(managerID)
	if err != nil {
		return models.ManagerDashboardResponse{}, errors.New("invalid manager ID format")
	}

	// Wrap in NullUUID for the queries
	managerNullUUID := uuid.NullUUID{
		UUID:  parsedManagerID,
		Valid: true,
	}

	// Get nomination counts
	counts, err := s.db.Read().CountTeamNominationsByManager(ctx, managerNullUUID)
	if err != nil {
		return models.ManagerDashboardResponse{}, fmt.Errorf("failed to get nomination counts: %w", err)
	}

	// Get team size
	teamSize, err := s.db.Read().CountTeamMembers(ctx, managerNullUUID)
	if err != nil {
		return models.ManagerDashboardResponse{}, fmt.Errorf("failed to get team size: %w", err)
	}

	dashboard := models.ManagerDashboardResponse{
		TotalNominations: counts.TotalCount,
		TeamSize:         teamSize,
	}

	// Safely extract counts from NullFloat64
	if counts.PendingCount.Valid {
		dashboard.PendingNominations = int64(counts.PendingCount.Float64)
	}
	if counts.ApprovedCount.Valid {
		dashboard.ApprovedNominations = int64(counts.ApprovedCount.Float64)
	}
	if counts.CompletedCount.Valid {
		dashboard.CompletedNominations = int64(counts.CompletedCount.Float64)
	}

	return dashboard, nil
}

// GetEmployeeDashboard retrieves dashboard KPIs for an employee.
func (s *service) GetEmployeeDashboard(ctx context.Context, employeeID string) (models.EmployeeDashboardResponse, error) {
	parsedEmployeeID, err := uuid.Parse(employeeID)
	if err != nil {
		return models.EmployeeDashboardResponse{}, errors.New("invalid employee ID format")
	}

	// Get nomination counts
	counts, err := s.db.Read().CountNominationsByUserID(ctx, parsedEmployeeID)
	if err != nil {
		return models.EmployeeDashboardResponse{}, fmt.Errorf("failed to get nomination counts: %w", err)
	}

	// Get available courses (upcoming trainings)
	upcomingTrainings, err := s.db.Read().ListUpcomingTrainings(ctx)
	if err != nil {
		return models.EmployeeDashboardResponse{}, fmt.Errorf("failed to get upcoming trainings: %w", err)
	}

	dashboard := models.EmployeeDashboardResponse{
		TotalNominations: counts.TotalCount,
		AvailableCourses: int64(len(upcomingTrainings)),
	}

	// Safely extract counts from NullFloat64
	if counts.PendingCount.Valid {
		dashboard.PendingNominations = int64(counts.PendingCount.Float64)
	}
	if counts.ApprovedCount.Valid {
		dashboard.ApprovedNominations = int64(counts.ApprovedCount.Float64)
	}
	if counts.CompletedCount.Valid {
		dashboard.CompletedNominations = int64(counts.CompletedCount.Float64)
	}

	return dashboard, nil
}

// GetAllPublishedCourses retrieves all published/upcoming courses.
func (s *service) GetAllPublishedCourses(ctx context.Context) ([]models.TrainingResponse, error) {
	trainings, err := s.db.Read().GetAllActiveAndUpcomingTrainings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming trainings: %w", err)
	}

	responses := make([]models.TrainingResponse, 0, len(trainings))
	for _, t := range trainings {
		responses = append(responses, MapTrainingToResponse(t))
	}

	return responses, nil
}

// MapNominationToResponse converts db.Nomination to models.NominationResponse.
func MapNominationToResponse(n db.Nomination) models.NominationResponse {
	resp := models.NominationResponse{
		ID:            n.ID.String(),
		Status:        n.Status,
		UserID:        n.UserID.String(),
		NominatedByID: n.NominatedByID.String(),
		CreatedAt:     n.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     n.UpdatedAt.Format(time.RFC3339),
	}

	if n.TrainingID.Valid {
		resp.TrainingID = n.TrainingID.UUID.String()
	}

	if n.HrCompletionStatus.Valid {
		resp.HrCompletionStatus = &n.HrCompletionStatus.String
	}
	if n.ProfFees.Valid {
		val := n.ProfFees.Int64
		resp.ProfFees = &val
	}
	if n.VenueCost.Valid {
		val := n.VenueCost.Int64
		resp.VenueCost = &val
	}
	if n.OtherCost.Valid {
		val := n.OtherCost.Int64
		resp.OtherCost = &val
	}
	if n.NonTemsTravel.Valid {
		val := n.NonTemsTravel.Int64
		resp.NonTemsTravel = &val
	}
	if n.NonTemsAccommodation.Valid {
		val := n.NonTemsAccommodation.Int64
		resp.NonTemsAccommodation = &val
	}
	if n.TotalCost.Valid {
		val := n.TotalCost.Int64
		resp.TotalCost = &val
	}

	return resp
}

// MapTrainingToResponse converts db.Training to models.TrainingResponse.
func MapTrainingToResponse(t db.Training) models.TrainingResponse {
	resp := models.TrainingResponse{
		ID:               t.ID.String(),
		Title:            t.Title,
		Category:         t.Category,
		StartDate:        t.StartDate.Format(time.RFC3339),
		EndDate:          t.EndDate.Format(time.RFC3339),
		CreatedByID:      t.CreatedByID.String(),
		DeadlineDays:     t.DeadlineDays,
		VenueCost:        t.VenueCost.Int64,
		ProfessionalFees: t.ProfessionalFees.Int64,
		StationaryCost:   t.StationaryCost.Int64,
		Status:           t.Status,
		ModeOfDelivery:   t.ModeOfDelivery,
		IsActive:         t.IsActive,
		CreatedAt:        t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        t.UpdatedAt.Format(time.RFC3339),
	}

	if t.Description.Valid {
		resp.Description = &t.Description.String
	}
	if t.InstructorName.Valid {
		resp.InstructorName = &t.InstructorName.String
	}
	if t.LearningOutcomes.Valid {
		resp.LearningOutcomes = &t.LearningOutcomes.String
	}
	if t.MonthTag.Valid {
		resp.MonthTag = &t.MonthTag.String
	}
	if t.StartTime.Valid {
		resp.StartTime = &t.StartTime.String
	}
	if t.EndTime.Valid {
		resp.EndTime = &t.EndTime.String
	}
	if t.Timezone.Valid {
		resp.Timezone = &t.Timezone.String
	}
	if t.Format.Valid {
		resp.Format = &t.Format.String
	}
	if t.RegistrationDeadline.Valid {
		rd := t.RegistrationDeadline.Time.Format(time.RFC3339)
		resp.RegistrationDeadline = &rd
	}
	if t.MaxCapacity.Valid {
		resp.MaxCapacity = &t.MaxCapacity.Int64
	}
	if t.TargetClusters.Valid {
		resp.TargetClusters = &t.TargetClusters.String
	}
	if t.PrerequisitesUrl.Valid {
		resp.PrerequisitesUrl = &t.PrerequisitesUrl.String
	}
	if t.Location.Valid {
		resp.Location = &t.Location.String
	}
	if t.VirtualLink.Valid {
		resp.VirtualLink = &t.VirtualLink.String
	}
	if t.PreReadUrl.Valid {
		resp.PreReadUrl = &t.PreReadUrl.String
	}
	if t.HrProgramID != uuid.Nil {
		id := t.HrProgramID.String()
		resp.HrProgramID = &id
	}
	if t.MappedCategory.Valid {
		resp.MappedCategory = &t.MappedCategory.String
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
	if t.FacilityID != uuid.Nil {
		id := t.FacilityID.String()
		resp.FacilityID = &id
	}

	return resp
}
