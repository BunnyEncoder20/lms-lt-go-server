package courses

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go-server/internal/database"
	"go-server/internal/database/db"
	"go-server/internal/models"

	"github.com/google/uuid"
)

var (
	ErrCourseNotFound      = errors.New("course not found")
	ErrModuleNotFound      = errors.New("module not found")
	ErrLessonNotFound      = errors.New("lesson not found")
	ErrArchivedCourse      = errors.New("cannot modify an archived course")
	ErrCourseAlreadyPub    = errors.New("course is already published")
	ErrRestoreOnlyArchived = errors.New("only archived courses can be restored")
	ErrCourseNoModules     = errors.New("course must have at least one module to publish")
	ErrCourseNoLessons     = errors.New("course must have at least one lesson to publish")
	ErrInvalidCourseID     = errors.New("invalid course ID format")
	ErrInvalidModuleID     = errors.New("invalid module ID format")
	ErrInvalidLessonID     = errors.New("invalid lesson ID format")
)

type Service interface {
	GetDashboardStats(ctx context.Context) (models.CourseDashboardStatsResponse, error)
	FindAll(ctx context.Context) ([]models.CourseListResponse, error)
	FindPublished(ctx context.Context) ([]models.CourseListResponse, error)
	FindOne(ctx context.Context, id string) (models.CourseDetailResponse, error)
	FindOnePublished(ctx context.Context, id string) (models.CourseDetailResponse, error)
	Create(ctx context.Context, dto models.CreateCourseRequest, authorID string) (models.CourseListResponse, error)
	Update(ctx context.Context, id string, dto models.UpdateCourseRequestV2) (models.CourseListResponse, error)
	Publish(ctx context.Context, id string) (models.CourseListResponse, error)
	Archive(ctx context.Context, id string) (models.CourseListResponse, error)
	Restore(ctx context.Context, id string) (models.CourseListResponse, error)
	Remove(ctx context.Context, id string) error
	CreateModule(ctx context.Context, courseID string, dto models.CreateCourseModuleRequest) (models.CourseModuleResponse, error)
	UpdateModule(ctx context.Context, moduleID string, dto models.UpdateCourseModuleRequest) (models.CourseModuleResponse, error)
	RemoveModule(ctx context.Context, moduleID string) error
	ReorderModules(ctx context.Context, dto models.ReorderCourseModulesRequest) error
	CreateLesson(ctx context.Context, moduleID string, dto models.CreateLessonRequest) (models.LessonResponse, error)
	UpdateLesson(ctx context.Context, lessonID string, dto models.UpdateLessonRequest) (models.LessonResponse, error)
	RemoveLesson(ctx context.Context, lessonID string) error
	ReorderLessons(ctx context.Context, dto models.ReorderLessonsRequest) error
}

type service struct {
	db database.Service
}

func NewService(db database.Service) Service {
	return &service{db: db}
}

func (s *service) GetDashboardStats(ctx context.Context) (models.CourseDashboardStatsResponse, error) {
	stats, err := s.db.Read().GetCourseDashboardStats(ctx)
	if err != nil {
		return models.CourseDashboardStatsResponse{}, err
	}

	published := int64(0)
	if stats.Published.Valid {
		published = int64(stats.Published.Float64)
	}
	draft := int64(0)
	if stats.Draft.Valid {
		draft = int64(stats.Draft.Float64)
	}
	archived := int64(0)
	if stats.Archived.Valid {
		archived = int64(stats.Archived.Float64)
	}

	completionRate := int64(0)
	if stats.TotalAssignments > 0 {
		completionRate = int64((stats.CompletedAssignments * 100) / stats.TotalAssignments)
	}

	return models.CourseDashboardStatsResponse{
		TotalCourses:         stats.TotalCourses,
		Published:            published,
		Draft:                draft,
		Archived:             archived,
		TotalLessons:         stats.TotalLessons,
		TotalAssignments:     stats.TotalAssignments,
		CompletedAssignments: stats.CompletedAssignments,
		CompletionRate:       completionRate,
	}, nil
}

func (s *service) FindAll(ctx context.Context) ([]models.CourseListResponse, error) {
	rows, err := s.db.Read().ListCourses(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]models.CourseListResponse, len(rows))
	for i, row := range rows {
		res[i] = mapCourseListRow(row)
	}

	return res, nil
}

func (s *service) FindPublished(ctx context.Context) ([]models.CourseListResponse, error) {
	rows, err := s.db.Read().ListPublishedCourses(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]models.CourseListResponse, len(rows))
	for i, row := range rows {
		res[i] = mapPublishedCourseListRow(row)
	}

	return res, nil
}

func (s *service) FindOne(ctx context.Context, id string) (models.CourseDetailResponse, error) {
	courseID, err := uuid.Parse(id)
	if err != nil {
		return models.CourseDetailResponse{}, ErrInvalidCourseID
	}

	return s.getCourseDetail(ctx, courseID, false)
}

func (s *service) FindOnePublished(ctx context.Context, id string) (models.CourseDetailResponse, error) {
	courseID, err := uuid.Parse(id)
	if err != nil {
		return models.CourseDetailResponse{}, ErrInvalidCourseID
	}

	return s.getCourseDetail(ctx, courseID, true)
}

func (s *service) Create(ctx context.Context, dto models.CreateCourseRequest, authorID string) (models.CourseListResponse, error) {
	authorUUID, err := uuid.Parse(authorID)
	if err != nil {
		return models.CourseListResponse{}, errors.New("invalid author ID format")
	}

	learningOutcomes, err := json.Marshal(dto.LearningOutcomes)
	if err != nil {
		return models.CourseListResponse{}, fmt.Errorf("failed to marshal learning outcomes: %w", err)
	}

	isStrict := false
	if dto.IsStrictSequencing != nil {
		isStrict = *dto.IsStrictSequencing
	}

	params := db.CreateCourseParams{
		ID:                 uuid.New(),
		Title:              dto.Title,
		AuthorID:           uuid.NullUUID{UUID: authorUUID, Valid: true},
		Status:             models.CourseDraft,
		Category:           models.TrainingCategory(dto.Category),
		LearningOutcomes:   string(learningOutcomes),
		IsStrictSequencing: isStrict,
		Version:            1,
	}

	if dto.Description != nil {
		params.Description = sql.NullString{String: *dto.Description, Valid: true}
	}
	if dto.CoverImageURL != nil {
		params.CoverImageUrl = sql.NullString{String: *dto.CoverImageURL, Valid: true}
	}
	if dto.EstimatedDuration != nil {
		params.EstimatedDuration = sql.NullInt64{Int64: *dto.EstimatedDuration, Valid: true}
	}

	course, err := s.db.Write().CreateCourse(ctx, params)
	if err != nil {
		return models.CourseListResponse{}, err
	}

	return s.mapCourseToListResponse(ctx, course)
}

func (s *service) Update(ctx context.Context, id string, dto models.UpdateCourseRequestV2) (models.CourseListResponse, error) {
	courseID, err := uuid.Parse(id)
	if err != nil {
		return models.CourseListResponse{}, ErrInvalidCourseID
	}

	existing, err := s.db.Read().GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.CourseListResponse{}, ErrCourseNotFound
		}
		return models.CourseListResponse{}, err
	}

	if existing.Status == models.CourseArchived {
		return models.CourseListResponse{}, ErrArchivedCourse
	}

	params := db.UpdateCourseParams{
		ID:                 existing.ID,
		Title:              existing.Title,
		Description:        existing.Description,
		CoverImageUrl:      existing.CoverImageUrl,
		Status:             existing.Status,
		Category:           existing.Category,
		EstimatedDuration:  existing.EstimatedDuration,
		LearningOutcomes:   existing.LearningOutcomes,
		IsStrictSequencing: existing.IsStrictSequencing,
		PublishedAt:        existing.PublishedAt,
	}

	if dto.Title != nil {
		params.Title = *dto.Title
	}
	if dto.Description != nil {
		params.Description = sql.NullString{String: *dto.Description, Valid: true}
	}
	if dto.CoverImageURL != nil {
		params.CoverImageUrl = sql.NullString{String: *dto.CoverImageURL, Valid: true}
	}
	if dto.Category != nil {
		params.Category = models.TrainingCategory(*dto.Category)
	}
	if dto.EstimatedDuration != nil {
		params.EstimatedDuration = sql.NullInt64{Int64: *dto.EstimatedDuration, Valid: true}
	}
	if dto.LearningOutcomes != nil {
		encoded, marshalErr := json.Marshal(dto.LearningOutcomes)
		if marshalErr != nil {
			return models.CourseListResponse{}, fmt.Errorf("failed to marshal learning outcomes: %w", marshalErr)
		}
		params.LearningOutcomes = string(encoded)
	}
	if dto.IsStrictSequencing != nil {
		params.IsStrictSequencing = *dto.IsStrictSequencing
	}

	course, err := s.db.Write().UpdateCourse(ctx, params)
	if err != nil {
		return models.CourseListResponse{}, err
	}

	return s.mapCourseToListResponse(ctx, course)
}

func (s *service) Publish(ctx context.Context, id string) (models.CourseListResponse, error) {
	courseID, err := uuid.Parse(id)
	if err != nil {
		return models.CourseListResponse{}, ErrInvalidCourseID
	}

	course, err := s.db.Read().GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.CourseListResponse{}, ErrCourseNotFound
		}
		return models.CourseListResponse{}, err
	}

	if course.Status == models.CoursePublished {
		return models.CourseListResponse{}, ErrCourseAlreadyPub
	}
	if course.Status == models.CourseArchived {
		return models.CourseListResponse{}, ErrArchivedCourse
	}

	moduleCount, err := s.db.Read().CountCourseModules(ctx, courseID)
	if err != nil {
		return models.CourseListResponse{}, err
	}
	if moduleCount == 0 {
		return models.CourseListResponse{}, ErrCourseNoModules
	}

	lessonCount, err := s.db.Read().CountCourseLessons(ctx, courseID)
	if err != nil {
		return models.CourseListResponse{}, err
	}
	if lessonCount == 0 {
		return models.CourseListResponse{}, ErrCourseNoLessons
	}

	updated, err := s.db.Write().PublishCourse(ctx, courseID)
	if err != nil {
		return models.CourseListResponse{}, err
	}

	return s.mapCourseToListResponse(ctx, updated)
}

func (s *service) Archive(ctx context.Context, id string) (models.CourseListResponse, error) {
	courseID, err := uuid.Parse(id)
	if err != nil {
		return models.CourseListResponse{}, ErrInvalidCourseID
	}

	updated, err := s.db.Write().ArchiveCourse(ctx, courseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.CourseListResponse{}, ErrCourseNotFound
		}
		return models.CourseListResponse{}, err
	}

	return s.mapCourseToListResponse(ctx, updated)
}

func (s *service) Restore(ctx context.Context, id string) (models.CourseListResponse, error) {
	courseID, err := uuid.Parse(id)
	if err != nil {
		return models.CourseListResponse{}, ErrInvalidCourseID
	}

	existing, err := s.db.Read().GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.CourseListResponse{}, ErrCourseNotFound
		}
		return models.CourseListResponse{}, err
	}
	if existing.Status != models.CourseArchived {
		return models.CourseListResponse{}, ErrRestoreOnlyArchived
	}

	updated, err := s.db.Write().RestoreCourse(ctx, courseID)
	if err != nil {
		return models.CourseListResponse{}, err
	}

	return s.mapCourseToListResponse(ctx, updated)
}

func (s *service) Remove(ctx context.Context, id string) error {
	courseID, err := uuid.Parse(id)
	if err != nil {
		return ErrInvalidCourseID
	}

	_, err = s.db.Read().GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCourseNotFound
		}
		return err
	}

	return s.db.ExecTx(ctx, func(qtx *db.Queries) error {
		if err := qtx.DeleteAttendanceDispatchesByCourseID(ctx, courseID); err != nil {
			return err
		}
		if err := qtx.DeleteNominationsByCourseID(ctx, uuid.NullUUID{UUID: courseID, Valid: true}); err != nil {
			return err
		}
		if err := qtx.DeleteManagerAllocationsByCourseID(ctx, courseID); err != nil {
			return err
		}
		if err := qtx.DeleteCourseAssignmentsByCourseID(ctx, courseID); err != nil {
			return err
		}
		if err := qtx.DeleteCourseByID(ctx, courseID); err != nil {
			return err
		}

		return nil
	})
}

func (s *service) CreateModule(ctx context.Context, courseID string, dto models.CreateCourseModuleRequest) (models.CourseModuleResponse, error) {
	parsedCourseID, err := uuid.Parse(courseID)
	if err != nil {
		return models.CourseModuleResponse{}, ErrInvalidCourseID
	}

	course, err := s.db.Read().GetCourseByID(ctx, parsedCourseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.CourseModuleResponse{}, ErrCourseNotFound
		}
		return models.CourseModuleResponse{}, err
	}
	if course.Status == models.CourseArchived {
		return models.CourseModuleResponse{}, ErrArchivedCourse
	}

	sequenceOrder := int64(0)
	if dto.SequenceOrder != nil {
		sequenceOrder = *dto.SequenceOrder
	} else {
		maxVal, maxErr := s.db.Read().GetMaxCourseModuleOrder(ctx, parsedCourseID)
		if maxErr != nil {
			return models.CourseModuleResponse{}, maxErr
		}
		sequenceOrder = parseInt64FromAny(maxVal) + 1
	}

	params := db.CreateCourseModuleParams{
		ID:            uuid.New(),
		Title:         dto.Title,
		CourseID:      parsedCourseID,
		SequenceOrder: sequenceOrder,
	}
	if dto.Description != nil {
		params.Description = sql.NullString{String: *dto.Description, Valid: true}
	}

	created, err := s.db.Write().CreateCourseModule(ctx, params)
	if err != nil {
		return models.CourseModuleResponse{}, err
	}

	return s.getModuleResponse(ctx, created.ID)
}

func (s *service) UpdateModule(ctx context.Context, moduleID string, dto models.UpdateCourseModuleRequest) (models.CourseModuleResponse, error) {
	parsedModuleID, err := uuid.Parse(moduleID)
	if err != nil {
		return models.CourseModuleResponse{}, ErrInvalidModuleID
	}

	module, err := s.db.Read().GetCourseModuleByID(ctx, parsedModuleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.CourseModuleResponse{}, ErrModuleNotFound
		}
		return models.CourseModuleResponse{}, err
	}

	course, err := s.db.Read().GetCourseByID(ctx, module.CourseID)
	if err != nil {
		return models.CourseModuleResponse{}, err
	}
	if course.Status == models.CourseArchived {
		return models.CourseModuleResponse{}, ErrArchivedCourse
	}

	params := db.UpdateCourseModuleParams{
		ID:            parsedModuleID,
		Title:         module.Title,
		Description:   module.Description,
		SequenceOrder: module.SequenceOrder,
	}
	if dto.Title != nil {
		params.Title = *dto.Title
	}
	if dto.Description != nil {
		params.Description = sql.NullString{String: *dto.Description, Valid: true}
	}
	if dto.SequenceOrder != nil {
		params.SequenceOrder = *dto.SequenceOrder
	}

	updated, err := s.db.Write().UpdateCourseModule(ctx, params)
	if err != nil {
		return models.CourseModuleResponse{}, err
	}

	return s.getModuleResponse(ctx, updated.ID)
}

func (s *service) RemoveModule(ctx context.Context, moduleID string) error {
	parsedModuleID, err := uuid.Parse(moduleID)
	if err != nil {
		return ErrInvalidModuleID
	}

	module, err := s.db.Read().GetCourseModuleByID(ctx, parsedModuleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrModuleNotFound
		}
		return err
	}

	course, err := s.db.Read().GetCourseByID(ctx, module.CourseID)
	if err != nil {
		return err
	}
	if course.Status == models.CourseArchived {
		return ErrArchivedCourse
	}

	return s.db.Write().DeleteCourseModuleByID(ctx, parsedModuleID)
}

func (s *service) ReorderModules(ctx context.Context, dto models.ReorderCourseModulesRequest) error {
	if len(dto.Modules) == 0 {
		return nil
	}

	return s.db.ExecTx(ctx, func(qtx *db.Queries) error {
		for idx, item := range dto.Modules {
			moduleID, err := uuid.Parse(item.ID)
			if err != nil {
				return ErrInvalidModuleID
			}

			if err := qtx.ReorderCourseModule(ctx, db.ReorderCourseModuleParams{
				ID:            moduleID,
				SequenceOrder: int64(idx + 1),
			}); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *service) CreateLesson(ctx context.Context, moduleID string, dto models.CreateLessonRequest) (models.LessonResponse, error) {
	parsedModuleID, err := uuid.Parse(moduleID)
	if err != nil {
		return models.LessonResponse{}, ErrInvalidModuleID
	}

	module, err := s.db.Read().GetCourseModuleByID(ctx, parsedModuleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.LessonResponse{}, ErrModuleNotFound
		}
		return models.LessonResponse{}, err
	}

	course, err := s.db.Read().GetCourseByID(ctx, module.CourseID)
	if err != nil {
		return models.LessonResponse{}, err
	}
	if course.Status == models.CourseArchived {
		return models.LessonResponse{}, ErrArchivedCourse
	}

	sequenceOrder := int64(0)
	if dto.SequenceOrder != nil {
		sequenceOrder = *dto.SequenceOrder
	} else {
		maxVal, maxErr := s.db.Read().GetMaxLessonOrder(ctx, parsedModuleID)
		if maxErr != nil {
			return models.LessonResponse{}, maxErr
		}
		sequenceOrder = parseInt64FromAny(maxVal) + 1
	}

	params := db.CreateLessonParams{
		ID:            uuid.New(),
		Title:         dto.Title,
		ContentType:   models.LessonContentType(dto.ContentType),
		ModuleID:      parsedModuleID,
		SequenceOrder: sequenceOrder,
	}
	if dto.AssetURL != nil {
		params.AssetUrl = sql.NullString{String: *dto.AssetURL, Valid: true}
	}
	if dto.RichTextContent != nil {
		params.RichTextContent = sql.NullString{String: *dto.RichTextContent, Valid: true}
	}
	if dto.DurationMinutes != nil {
		params.DurationMinutes = sql.NullInt64{Int64: *dto.DurationMinutes, Valid: true}
	}

	lesson, err := s.db.Write().CreateLesson(ctx, params)
	if err != nil {
		return models.LessonResponse{}, err
	}

	return mapLessonToResponse(lesson), nil
}

func (s *service) UpdateLesson(ctx context.Context, lessonID string, dto models.UpdateLessonRequest) (models.LessonResponse, error) {
	parsedLessonID, err := uuid.Parse(lessonID)
	if err != nil {
		return models.LessonResponse{}, ErrInvalidLessonID
	}

	lesson, err := s.db.Read().GetLessonByID(ctx, parsedLessonID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.LessonResponse{}, ErrLessonNotFound
		}
		return models.LessonResponse{}, err
	}

	module, err := s.db.Read().GetCourseModuleByID(ctx, lesson.ModuleID)
	if err != nil {
		return models.LessonResponse{}, err
	}
	course, err := s.db.Read().GetCourseByID(ctx, module.CourseID)
	if err != nil {
		return models.LessonResponse{}, err
	}
	if course.Status == models.CourseArchived {
		return models.LessonResponse{}, ErrArchivedCourse
	}

	params := db.UpdateLessonParams{
		ID:              parsedLessonID,
		Title:           lesson.Title,
		ContentType:     lesson.ContentType,
		AssetUrl:        lesson.AssetUrl,
		RichTextContent: lesson.RichTextContent,
		DurationMinutes: lesson.DurationMinutes,
		SequenceOrder:   lesson.SequenceOrder,
	}

	if dto.Title != nil {
		params.Title = *dto.Title
	}
	if dto.ContentType != nil {
		params.ContentType = models.LessonContentType(*dto.ContentType)
	}
	if dto.AssetURL != nil {
		params.AssetUrl = sql.NullString{String: *dto.AssetURL, Valid: true}
	}
	if dto.RichTextContent != nil {
		params.RichTextContent = sql.NullString{String: *dto.RichTextContent, Valid: true}
	}
	if dto.DurationMinutes != nil {
		params.DurationMinutes = sql.NullInt64{Int64: *dto.DurationMinutes, Valid: true}
	}
	if dto.SequenceOrder != nil {
		params.SequenceOrder = *dto.SequenceOrder
	}

	updated, err := s.db.Write().UpdateLesson(ctx, params)
	if err != nil {
		return models.LessonResponse{}, err
	}

	return mapLessonToResponse(updated), nil
}

func (s *service) RemoveLesson(ctx context.Context, lessonID string) error {
	parsedLessonID, err := uuid.Parse(lessonID)
	if err != nil {
		return ErrInvalidLessonID
	}

	lesson, err := s.db.Read().GetLessonByID(ctx, parsedLessonID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrLessonNotFound
		}
		return err
	}

	module, err := s.db.Read().GetCourseModuleByID(ctx, lesson.ModuleID)
	if err != nil {
		return err
	}
	course, err := s.db.Read().GetCourseByID(ctx, module.CourseID)
	if err != nil {
		return err
	}
	if course.Status == models.CourseArchived {
		return ErrArchivedCourse
	}

	return s.db.Write().DeleteLessonByID(ctx, parsedLessonID)
}

func (s *service) ReorderLessons(ctx context.Context, dto models.ReorderLessonsRequest) error {
	if len(dto.Lessons) == 0 {
		return nil
	}

	return s.db.ExecTx(ctx, func(qtx *db.Queries) error {
		for idx, item := range dto.Lessons {
			lessonID, err := uuid.Parse(item.ID)
			if err != nil {
				return ErrInvalidLessonID
			}

			if err := qtx.ReorderLesson(ctx, db.ReorderLessonParams{
				ID:            lessonID,
				SequenceOrder: int64(idx + 1),
			}); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *service) getCourseDetail(ctx context.Context, id uuid.UUID, publishedOnly bool) (models.CourseDetailResponse, error) {
	course, err := s.db.Read().GetCourseWithAuthor(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.CourseDetailResponse{}, ErrCourseNotFound
		}
		return models.CourseDetailResponse{}, err
	}

	if publishedOnly && course.Status != models.CoursePublished {
		return models.CourseDetailResponse{}, ErrCourseNotFound
	}

	modules, err := s.db.Read().ListCourseModulesByCourseID(ctx, id)
	if err != nil {
		return models.CourseDetailResponse{}, err
	}

	moduleResponses := make([]models.CourseModuleResponse, len(modules))
	for i, mod := range modules {
		lessons, lessonsErr := s.db.Read().ListLessonsByModuleID(ctx, mod.ID)
		if lessonsErr != nil {
			return models.CourseDetailResponse{}, lessonsErr
		}

		lessonResponses := make([]models.LessonResponse, len(lessons))
		for j, lesson := range lessons {
			lessonResponses[j] = mapLessonToResponse(lesson)
		}

		moduleResponses[i] = models.CourseModuleResponse{
			ID:            mod.ID.String(),
			Title:         mod.Title,
			Description:   toStringPointer(mod.Description),
			SequenceOrder: mod.SequenceOrder,
			Lessons:       lessonResponses,
		}
	}

	assignmentsCount, err := s.db.Read().CountCourseAssignmentsByCourseID(ctx, id)
	if err != nil {
		return models.CourseDetailResponse{}, err
	}

	resp := models.CourseDetailResponse{
		ID:                 course.ID.String(),
		Title:              course.Title,
		Description:        toStringPointer(course.Description),
		CoverImageURL:      toStringPointer(course.CoverImageUrl),
		Status:             course.Status,
		Category:           course.Category,
		EstimatedDuration:  toInt64Pointer(course.EstimatedDuration),
		LearningOutcomes:   parseLearningOutcomes(course.LearningOutcomes),
		IsStrictSequencing: course.IsStrictSequencing,
		Version:            course.Version,
		PublishedAt:        toTimePointer(course.PublishedAt),
		CreatedAt:          course.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          course.UpdatedAt.Format(time.RFC3339),
		Modules:            moduleResponses,
		Count: models.CourseCountResponse{
			Modules:     int64(len(moduleResponses)),
			Assignments: assignmentsCount,
		},
	}

	if course.AuthorID.Valid {
		aid := course.AuthorID.UUID.String()
		resp.AuthorID = &aid
	}
	if course.AuthorFirstName.Valid || course.AuthorLastName.Valid {
		author := &models.CourseAuthorResponse{}
		if course.AuthorID.Valid {
			author.ID = course.AuthorID.UUID.String()
		}
		if course.AuthorFirstName.Valid {
			author.FirstName = course.AuthorFirstName.String
		}
		if course.AuthorLastName.Valid {
			author.LastName = course.AuthorLastName.String
		}
		resp.Author = author
	}

	return resp, nil
}

func (s *service) mapCourseToListResponse(ctx context.Context, course db.Course) (models.CourseListResponse, error) {
	listRows, err := s.db.Read().ListCourses(ctx)
	if err != nil {
		return models.CourseListResponse{}, err
	}

	for _, row := range listRows {
		if row.ID == course.ID {
			return mapCourseListRow(row), nil
		}
	}

	// Fallback for edge cases where the course row isn't present in the list query result yet.
	return models.CourseListResponse{
		ID:                 course.ID.String(),
		Title:              course.Title,
		Description:        toStringPointer(course.Description),
		CoverImageURL:      toStringPointer(course.CoverImageUrl),
		Status:             course.Status,
		Category:           course.Category,
		EstimatedDuration:  toInt64Pointer(course.EstimatedDuration),
		LearningOutcomes:   parseLearningOutcomes(course.LearningOutcomes),
		IsStrictSequencing: course.IsStrictSequencing,
		Version:            course.Version,
		PublishedAt:        toTimePointer(course.PublishedAt),
		CreatedAt:          course.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          course.UpdatedAt.Format(time.RFC3339),
		Count:              models.CourseCountResponse{},
	}, nil
}

func (s *service) getModuleResponse(ctx context.Context, moduleID uuid.UUID) (models.CourseModuleResponse, error) {
	mod, err := s.db.Read().GetCourseModuleByID(ctx, moduleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.CourseModuleResponse{}, ErrModuleNotFound
		}
		return models.CourseModuleResponse{}, err
	}

	lessons, err := s.db.Read().ListLessonsByModuleID(ctx, moduleID)
	if err != nil {
		return models.CourseModuleResponse{}, err
	}

	resp := models.CourseModuleResponse{
		ID:            mod.ID.String(),
		Title:         mod.Title,
		Description:   toStringPointer(mod.Description),
		SequenceOrder: mod.SequenceOrder,
		Lessons:       make([]models.LessonResponse, len(lessons)),
	}

	for i, lesson := range lessons {
		resp.Lessons[i] = mapLessonToResponse(lesson)
	}

	return resp, nil
}

func mapCourseListRow(row db.ListCoursesRow) models.CourseListResponse {
	resp := models.CourseListResponse{
		ID:                 row.ID.String(),
		Title:              row.Title,
		Description:        toStringPointer(row.Description),
		CoverImageURL:      toStringPointer(row.CoverImageUrl),
		Status:             row.Status,
		Category:           row.Category,
		EstimatedDuration:  toInt64Pointer(row.EstimatedDuration),
		LearningOutcomes:   parseLearningOutcomes(row.LearningOutcomes),
		IsStrictSequencing: row.IsStrictSequencing,
		Version:            row.Version,
		PublishedAt:        toTimePointer(row.PublishedAt),
		CreatedAt:          row.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          row.UpdatedAt.Format(time.RFC3339),
		Count: models.CourseCountResponse{
			Modules:     row.ModuleCount,
			Assignments: row.AssignmentCount,
		},
	}

	if row.AuthorID.Valid || row.AuthorFirstName.Valid || row.AuthorLastName.Valid {
		author := &models.CourseAuthorResponse{}
		if row.AuthorID.Valid {
			author.ID = row.AuthorID.UUID.String()
		}
		if row.AuthorFirstName.Valid {
			author.FirstName = row.AuthorFirstName.String
		}
		if row.AuthorLastName.Valid {
			author.LastName = row.AuthorLastName.String
		}
		resp.Author = author
	}

	return resp
}

func mapPublishedCourseListRow(row db.ListPublishedCoursesRow) models.CourseListResponse {
	resp := models.CourseListResponse{
		ID:                 row.ID.String(),
		Title:              row.Title,
		Description:        toStringPointer(row.Description),
		CoverImageURL:      toStringPointer(row.CoverImageUrl),
		Status:             row.Status,
		Category:           row.Category,
		EstimatedDuration:  toInt64Pointer(row.EstimatedDuration),
		LearningOutcomes:   parseLearningOutcomes(row.LearningOutcomes),
		IsStrictSequencing: row.IsStrictSequencing,
		Version:            row.Version,
		PublishedAt:        toTimePointer(row.PublishedAt),
		CreatedAt:          row.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          row.UpdatedAt.Format(time.RFC3339),
		Count: models.CourseCountResponse{
			Modules:     row.ModuleCount,
			Assignments: row.AssignmentCount,
		},
	}

	if row.AuthorID.Valid || row.AuthorFirstName.Valid || row.AuthorLastName.Valid {
		author := &models.CourseAuthorResponse{}
		if row.AuthorID.Valid {
			author.ID = row.AuthorID.UUID.String()
		}
		if row.AuthorFirstName.Valid {
			author.FirstName = row.AuthorFirstName.String
		}
		if row.AuthorLastName.Valid {
			author.LastName = row.AuthorLastName.String
		}
		resp.Author = author
	}

	return resp
}

func mapLessonToResponse(lesson db.Lesson) models.LessonResponse {
	return models.LessonResponse{
		ID:              lesson.ID.String(),
		Title:           lesson.Title,
		ContentType:     lesson.ContentType,
		AssetURL:        toStringPointer(lesson.AssetUrl),
		RichTextContent: toStringPointer(lesson.RichTextContent),
		DurationMinutes: toInt64Pointer(lesson.DurationMinutes),
		SequenceOrder:   lesson.SequenceOrder,
	}
}

func toStringPointer(val sql.NullString) *string {
	if !val.Valid {
		return nil
	}
	v := val.String
	return &v
}

func toInt64Pointer(val sql.NullInt64) *int64 {
	if !val.Valid {
		return nil
	}
	v := val.Int64
	return &v
}

func toTimePointer(val sql.NullTime) *string {
	if !val.Valid {
		return nil
	}
	v := val.Time.Format(time.RFC3339)
	return &v
}

func parseLearningOutcomes(encoded string) []string {
	if encoded == "" {
		return []string{}
	}

	var outcomes []string
	if err := json.Unmarshal([]byte(encoded), &outcomes); err != nil {
		return []string{}
	}

	return outcomes
}

func parseInt64FromAny(value interface{}) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case float32:
		return int64(v)
	default:
		return 0
	}
}
