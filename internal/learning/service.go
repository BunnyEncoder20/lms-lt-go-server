package learning

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"

	"go-server/internal/database"
	"go-server/internal/database/db"
	"go-server/internal/models"
	"go-server/internal/notifications"

	"github.com/google/uuid"
)

var (
	ErrLearningCourseNotFound    = errors.New("course not found")
	ErrLearningLessonNotFound    = errors.New("lesson not found")
	ErrLearningAssignmentMissing = errors.New("you are not assigned to this course")
	ErrLearningAlreadyCompleted  = errors.New("lesson is already marked as completed")
	ErrLearningThresholdFailed   = errors.New("lesson completion requirements are not satisfied yet")
	ErrLearningInvalidDirection  = errors.New("direction must be next or previous")
)

type Service interface {
	FindPublishedCourses(ctx context.Context) ([]models.AssignmentCourseSummaryResponse, error)
	BulkAssign(ctx context.Context, dto models.BulkAssignRequest, assignedByID string) ([]any, error)
	FindAllAssignments(ctx context.Context) ([]models.AssignmentSummaryResponse, error)
	FindMyAssignments(ctx context.Context, userID string) ([]models.AssignmentSummaryResponse, error)
	GetCoursePlayerData(ctx context.Context, courseID string, userID string) (models.CoursePlayerResponse, error)
	GetCourseNavigation(ctx context.Context, courseID string, userID string) (models.CourseNavigationResponse, error)
	GetAdjacentLesson(ctx context.Context, lessonID string, userID string, direction string) (models.AdjacentLessonResponse, error)
	UpdateLessonProgress(ctx context.Context, lessonID string, userID string, dto models.UpdateProgressRequest) (models.LessonProgressMutationResponse, error)
	HeartbeatLessonProgress(ctx context.Context, lessonID string, userID string, dto models.HeartbeatProgressRequest) (models.LessonProgressMutationResponse, error)
	CompleteLesson(ctx context.Context, lessonID string, userID string, dto models.CompleteLessonRequest) (models.LessonProgressMutationResponse, error)
}

type service struct {
	db            database.Service
	notifications notifications.Service
}

type completionThresholds struct {
	videoMinPercent          float64
	audioMinPercent          float64
	richTextMinScrollPercent float64
	richTextMinSeconds       int64
	pdfMinSeconds            int64
	imageMinSeconds          int64
	presentationMinSeconds   int64
}

var thresholds = completionThresholds{
	videoMinPercent:          90,
	audioMinPercent:          90,
	richTextMinScrollPercent: 90,
	richTextMinSeconds:       15,
	pdfMinSeconds:            20,
	imageMinSeconds:          12,
	presentationMinSeconds:   20,
}

func NewService(dbSvc database.Service, notifier notifications.Service) Service {
	if notifier == nil {
		notifier = notifications.NewNoopService(nil)
	}

	return &service{db: dbSvc, notifications: notifier}
}

func (s *service) FindPublishedCourses(ctx context.Context) ([]models.AssignmentCourseSummaryResponse, error) {
	rows, err := s.db.Read().ListPublishedCourses(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]models.AssignmentCourseSummaryResponse, len(rows))
	for i, row := range rows {
		items[i] = mapPublishedCourseSummary(row)
	}

	return items, nil
}

func (s *service) BulkAssign(ctx context.Context, dto models.BulkAssignRequest, assignedByID string) ([]any, error) {
	courseID, err := uuid.Parse(dto.CourseID)
	if err != nil {
		return nil, errors.New("invalid course ID format")
	}

	assignerID, err := uuid.Parse(assignedByID)
	if err != nil {
		return nil, errors.New("invalid assigner ID format")
	}

	course, err := s.db.Read().GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLearningCourseNotFound
		}
		return nil, err
	}
	if course.Status != models.CoursePublished {
		return nil, errors.New("can only assign published courses")
	}

	var dueDate sql.NullTime
	if dto.DueDate != nil && *dto.DueDate != "" {
		parsedDue, parseErr := time.Parse(time.RFC3339, *dto.DueDate)
		if parseErr != nil {
			return nil, errors.New("invalid due_date format, expected RFC3339")
		}
		dueDate = sql.NullTime{Time: parsedDue, Valid: true}
	}

	results := make([]any, 0, len(dto.UserIDs))
	successUserIDs := make([]string, 0, len(dto.UserIDs))

	err = s.db.ExecTx(ctx, func(qtx *db.Queries) error {
		for _, userID := range dto.UserIDs {
			uid, parseErr := uuid.Parse(userID)
			if parseErr != nil {
				results = append(results, map[string]any{"userId": userID, "status": "invalid_user_id"})
				continue
			}

			_, getErr := qtx.GetCourseAssignmentByUserCourse(ctx, db.GetCourseAssignmentByUserCourseParams{
				UserID:   uuid.NullUUID{UUID: uid, Valid: true},
				CourseID: courseID,
			})
			if getErr == nil {
				results = append(results, map[string]any{"userId": userID, "status": "already_assigned"})
				continue
			}
			if !errors.Is(getErr, sql.ErrNoRows) {
				return getErr
			}

			assignment, createErr := qtx.CreateCourseAssignment(ctx, db.CreateCourseAssignmentParams{
				ID:                 uuid.New(),
				Status:             models.AssignmentNotStarted,
				ProgressPercentage: 0,
				CourseVersion:      course.Version,
				DueDate:            dueDate,
				CourseID:           courseID,
				UserID:             uuid.NullUUID{UUID: uid, Valid: true},
				AssignedByID:       uuid.NullUUID{UUID: assignerID, Valid: true},
			})
			if createErr != nil {
				results = append(results, map[string]any{"userId": userID, "status": "failed", "reason": createErr.Error()})
				continue
			}

			summary, mapErr := s.mapAssignmentSummary(ctx, assignment)
			if mapErr != nil {
				return mapErr
			}
			results = append(results, summary)
			successUserIDs = append(successUserIDs, userID)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(successUserIDs) > 0 {
		usersCopy := append([]string(nil), successUserIDs...)
		courseIDCopy := courseID.String()
		go func(users []string, cid string) {
			for _, uid := range users {
				_, publishErr := s.notifications.PublishHrEvent(context.Background(), notifications.HrEvent{
					Type:    models.HrNotificationCourseAssigned,
					Title:   "Course Assigned",
					Message: "You have a new mandatory course.",
					Payload: map[string]any{
						"courseId": cid,
						"userId":   uid,
					},
				})
				if publishErr != nil {
					slog.Default().Error("failed publishing COURSE_ASSIGNED notification", "user_id", uid, "course_id", cid, "error", publishErr)
					continue
				}

				s.notifications.PublishUserEvent(context.Background(), uid, "hr.notification", map[string]any{
					"type":     models.HrNotificationCourseAssigned,
					"message":  "You have a new mandatory course.",
					"courseId": cid,
				})
			}
		}(usersCopy, courseIDCopy)
	}

	return results, nil
}

func (s *service) FindAllAssignments(ctx context.Context) ([]models.AssignmentSummaryResponse, error) {
	rows, err := s.db.Read().ListCourseAssignments(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]models.AssignmentSummaryResponse, len(rows))
	for i, row := range rows {
		items[i], err = s.mapAssignmentSummary(ctx, row)
		if err != nil {
			return nil, err
		}
	}

	return items, nil
}

func (s *service) FindMyAssignments(ctx context.Context, userID string) ([]models.AssignmentSummaryResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	rows, err := s.db.Read().ListCourseAssignmentsByUserID(ctx, uuid.NullUUID{UUID: uid, Valid: true})
	if err != nil {
		return nil, err
	}

	items := make([]models.AssignmentSummaryResponse, len(rows))
	for i, row := range rows {
		items[i], err = s.mapAssignmentSummary(ctx, row)
		if err != nil {
			return nil, err
		}
	}

	return items, nil
}

func (s *service) GetCoursePlayerData(ctx context.Context, courseID string, userID string) (models.CoursePlayerResponse, error) {
	cid, err := uuid.Parse(courseID)
	if err != nil {
		return models.CoursePlayerResponse{}, errors.New("invalid course ID format")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return models.CoursePlayerResponse{}, errors.New("invalid user ID format")
	}

	assignment, err := s.db.Read().GetCourseAssignmentByUserCourse(ctx, db.GetCourseAssignmentByUserCourseParams{
		UserID:   uuid.NullUUID{UUID: uid, Valid: true},
		CourseID: cid,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.CoursePlayerResponse{}, ErrLearningAssignmentMissing
		}
		return models.CoursePlayerResponse{}, err
	}

	course, err := s.db.Read().GetCourseByID(ctx, cid)
	if err != nil {
		return models.CoursePlayerResponse{}, err
	}

	mods, err := s.db.Read().ListCourseModulesByCourseID(ctx, cid)
	if err != nil {
		return models.CoursePlayerResponse{}, err
	}

	progressList, err := s.db.Read().ListLessonProgressByAssignment(ctx, assignment.ID)
	if err != nil {
		return models.CoursePlayerResponse{}, err
	}
	progressByLesson := map[uuid.UUID]db.LessonProgress{}
	for _, p := range progressList {
		progressByLesson[p.LessonID] = p
	}

	resp := models.CoursePlayerResponse{
		AssignmentID:       assignment.ID.String(),
		Status:             assignment.Status,
		ProgressPercentage: assignment.ProgressPercentage,
		EnrolledAt:         assignment.EnrolledAt.Format(time.RFC3339),
		CompletedAt:        nullTimePtr(assignment.CompletedAt),
		DueDate:            nullTimePtr(assignment.DueDate),
	}
	resp.Course.ID = course.ID.String()
	resp.Course.Title = course.Title
	resp.Course.Description = nullStringPtr(course.Description)
	resp.Course.CoverImageURL = nullStringPtr(course.CoverImageUrl)
	resp.Course.Category = course.Category
	resp.Course.EstimatedDuration = nullInt64Ptr(course.EstimatedDuration)
	resp.Course.LearningOutcomes = parseOutcomes(course.LearningOutcomes)
	resp.Course.IsStrictSequencing = course.IsStrictSequencing

	if course.AuthorID.Valid {
		author, err := s.db.Read().GetUserByID(ctx, course.AuthorID.UUID)
		if err == nil {
			resp.Course.Author = &models.CourseAuthorResponse{
				ID:        author.ID.String(),
				FirstName: author.FirstName,
				LastName:  author.LastName,
			}
		}
	}

	modules := make([]models.PlayerModuleResponse, len(mods))
	for i, mod := range mods {
		lessons, err := s.db.Read().ListLessonsByModuleID(ctx, mod.ID)
		if err != nil {
			return models.CoursePlayerResponse{}, err
		}

		playerLessons := make([]models.PlayerLessonResponse, len(lessons))
		for j, l := range lessons {
			p, ok := progressByLesson[l.ID]
			progress := models.LessonProgressStateResponse{IsCompleted: false, LastPlaybackPosition: 0}
			if ok {
				progress.IsCompleted = p.IsCompleted
				progress.LastPlaybackPosition = p.LastPlaybackPosition
				progress.CompletedAt = nullTimePtr(p.CompletedAt)
			}
			playerLessons[j] = models.PlayerLessonResponse{
				ID:              l.ID.String(),
				Title:           l.Title,
				ContentType:     l.ContentType,
				AssetURL:        nullStringPtr(l.AssetUrl),
				RichTextContent: nullStringPtr(l.RichTextContent),
				DurationMinutes: nullInt64Ptr(l.DurationMinutes),
				SequenceOrder:   l.SequenceOrder,
				Progress:        progress,
			}
		}

		modules[i] = models.PlayerModuleResponse{
			ID:            mod.ID.String(),
			Title:         mod.Title,
			SequenceOrder: mod.SequenceOrder,
			Lessons:       playerLessons,
		}
	}
	resp.Course.Modules = modules

	return resp, nil
}

func (s *service) GetCourseNavigation(ctx context.Context, courseID string, userID string) (models.CourseNavigationResponse, error) {
	cid, err := uuid.Parse(courseID)
	if err != nil {
		return models.CourseNavigationResponse{}, errors.New("invalid course ID format")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return models.CourseNavigationResponse{}, errors.New("invalid user ID format")
	}

	assignment, err := s.db.Read().GetCourseAssignmentByUserCourse(ctx, db.GetCourseAssignmentByUserCourseParams{
		UserID:   uuid.NullUUID{UUID: uid, Valid: true},
		CourseID: cid,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.CourseNavigationResponse{}, ErrLearningAssignmentMissing
		}
		return models.CourseNavigationResponse{}, err
	}

	course, err := s.db.Read().GetCourseByID(ctx, cid)
	if err != nil {
		return models.CourseNavigationResponse{}, err
	}

	mods, err := s.db.Read().ListCourseModulesByCourseID(ctx, cid)
	if err != nil {
		return models.CourseNavigationResponse{}, err
	}

	progressList, err := s.db.Read().ListLessonProgressByAssignment(ctx, assignment.ID)
	if err != nil {
		return models.CourseNavigationResponse{}, err
	}
	progressByLesson := map[uuid.UUID]db.LessonProgress{}
	for _, p := range progressList {
		progressByLesson[p.LessonID] = p
	}

	type flatLesson struct {
		ModuleID            uuid.UUID
		ModuleTitle         string
		ModuleSequenceOrder int64
		LessonID            uuid.UUID
		LessonTitle         string
		LessonSequenceOrder int64
		IsCompleted         bool
		LastPlayback        int64
		CompletedAt         *string
	}

	flat := make([]flatLesson, 0)
	moduleLessons := map[uuid.UUID][]db.Lesson{}
	for _, mod := range mods {
		lessons, err := s.db.Read().ListLessonsByModuleID(ctx, mod.ID)
		if err != nil {
			return models.CourseNavigationResponse{}, err
		}
		moduleLessons[mod.ID] = lessons
		for _, l := range lessons {
			p, ok := progressByLesson[l.ID]
			isCompleted := false
			playback := int64(0)
			var completedAt *string
			if ok {
				isCompleted = p.IsCompleted
				playback = p.LastPlaybackPosition
				completedAt = nullTimePtr(p.CompletedAt)
			}
			flat = append(flat, flatLesson{
				ModuleID:            mod.ID,
				ModuleTitle:         mod.Title,
				ModuleSequenceOrder: mod.SequenceOrder,
				LessonID:            l.ID,
				LessonTitle:         l.Title,
				LessonSequenceOrder: l.SequenceOrder,
				IsCompleted:         isCompleted,
				LastPlayback:        playback,
				CompletedAt:         completedAt,
			})
		}
	}

	sort.Slice(flat, func(i, j int) bool {
		if flat[i].ModuleSequenceOrder != flat[j].ModuleSequenceOrder {
			return flat[i].ModuleSequenceOrder < flat[j].ModuleSequenceOrder
		}
		return flat[i].LessonSequenceOrder < flat[j].LessonSequenceOrder
	})

	indexByLesson := map[uuid.UUID]int{}
	for i, item := range flat {
		indexByLesson[item.LessonID] = i
	}

	var resp models.CourseNavigationResponse
	resp.Course.ID = course.ID.String()
	resp.Course.Title = course.Title
	resp.Course.IsStrictSequencing = course.IsStrictSequencing
	resp.Assignment.ID = assignment.ID.String()
	resp.Assignment.Status = assignment.Status
	resp.Assignment.ProgressPercentage = assignment.ProgressPercentage

	resp.Modules = make([]models.NavigationModuleResponse, len(mods))
	for i, mod := range mods {
		lessons := moduleLessons[mod.ID]
		navLessons := make([]models.NavigationLessonResponse, len(lessons))
		for j, l := range lessons {
			idx := indexByLesson[l.ID]
			current := flat[idx]
			var previous *flatLesson
			if idx > 0 {
				previous = &flat[idx-1]
			}

			locked := course.IsStrictSequencing && previous != nil && !previous.IsCompleted
			status := "NOT_STARTED"
			if current.IsCompleted {
				status = "COMPLETED"
			} else if current.LastPlayback > 0 {
				status = "IN_PROGRESS"
			}

			navLessons[j] = models.NavigationLessonResponse{
				ID:            l.ID.String(),
				Title:         l.Title,
				SequenceOrder: l.SequenceOrder,
				Status:        status,
				Locked:        locked,
				Progress: models.LessonProgressStateResponse{
					IsCompleted:          current.IsCompleted,
					LastPlaybackPosition: current.LastPlayback,
					CompletedAt:          current.CompletedAt,
				},
			}
		}

		resp.Modules[i] = models.NavigationModuleResponse{
			ID:            mod.ID.String(),
			Title:         mod.Title,
			SequenceOrder: mod.SequenceOrder,
			Lessons:       navLessons,
		}
	}

	return resp, nil
}

func (s *service) GetAdjacentLesson(ctx context.Context, lessonID string, userID string, direction string) (models.AdjacentLessonResponse, error) {
	if direction != "next" && direction != "previous" {
		return models.AdjacentLessonResponse{}, ErrLearningInvalidDirection
	}

	lessonUUID, err := uuid.Parse(lessonID)
	if err != nil {
		return models.AdjacentLessonResponse{}, errors.New("invalid lesson ID format")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return models.AdjacentLessonResponse{}, errors.New("invalid user ID format")
	}

	lesson, err := s.db.Read().GetLessonByID(ctx, lessonUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.AdjacentLessonResponse{}, ErrLearningLessonNotFound
		}
		return models.AdjacentLessonResponse{}, err
	}

	module, err := s.db.Read().GetCourseModuleByID(ctx, lesson.ModuleID)
	if err != nil {
		return models.AdjacentLessonResponse{}, err
	}

	_, err = s.db.Read().GetCourseAssignmentByUserCourse(ctx, db.GetCourseAssignmentByUserCourseParams{
		UserID:   uuid.NullUUID{UUID: uid, Valid: true},
		CourseID: module.CourseID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.AdjacentLessonResponse{}, ErrLearningAssignmentMissing
		}
		return models.AdjacentLessonResponse{}, err
	}

	type orderedLesson struct {
		LessonID            uuid.UUID
		LessonTitle         string
		LessonSequenceOrder int64
		ModuleID            uuid.UUID
		ModuleTitle         string
		ModuleSequenceOrder int64
	}

	mods, err := s.db.Read().ListCourseModulesByCourseID(ctx, module.CourseID)
	if err != nil {
		return models.AdjacentLessonResponse{}, err
	}

	ordered := make([]orderedLesson, 0)
	for _, mod := range mods {
		lessons, err := s.db.Read().ListLessonsByModuleID(ctx, mod.ID)
		if err != nil {
			return models.AdjacentLessonResponse{}, err
		}
		for _, l := range lessons {
			ordered = append(ordered, orderedLesson{
				LessonID:            l.ID,
				LessonTitle:         l.Title,
				LessonSequenceOrder: l.SequenceOrder,
				ModuleID:            mod.ID,
				ModuleTitle:         mod.Title,
				ModuleSequenceOrder: mod.SequenceOrder,
			})
		}
	}

	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].ModuleSequenceOrder != ordered[j].ModuleSequenceOrder {
			return ordered[i].ModuleSequenceOrder < ordered[j].ModuleSequenceOrder
		}
		return ordered[i].LessonSequenceOrder < ordered[j].LessonSequenceOrder
	})

	idx := -1
	for i, item := range ordered {
		if item.LessonID == lessonUUID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return models.AdjacentLessonResponse{}, ErrLearningLessonNotFound
	}

	target := idx + 1
	if direction == "previous" {
		target = idx - 1
	}

	resp := models.AdjacentLessonResponse{CurrentLessonID: lessonID, Direction: direction}
	if target < 0 || target >= len(ordered) {
		return resp, nil
	}

	t := ordered[target]
	resp.Lesson = &struct {
		ID            string `json:"id"`
		Title         string `json:"title"`
		SequenceOrder int64  `json:"sequence_order"`
		Module        struct {
			ID            string `json:"id"`
			Title         string `json:"title"`
			SequenceOrder int64  `json:"sequence_order"`
		} `json:"module"`
	}{
		ID:            t.LessonID.String(),
		Title:         t.LessonTitle,
		SequenceOrder: t.LessonSequenceOrder,
	}
	resp.Lesson.Module.ID = t.ModuleID.String()
	resp.Lesson.Module.Title = t.ModuleTitle
	resp.Lesson.Module.SequenceOrder = t.ModuleSequenceOrder

	return resp, nil
}

func (s *service) UpdateLessonProgress(ctx context.Context, lessonID string, userID string, dto models.UpdateProgressRequest) (models.LessonProgressMutationResponse, error) {
	lessonUUID, err := uuid.Parse(lessonID)
	if err != nil {
		return models.LessonProgressMutationResponse{}, errors.New("invalid lesson ID format")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return models.LessonProgressMutationResponse{}, errors.New("invalid user ID format")
	}

	lesson, err := s.db.Read().GetLessonByID(ctx, lessonUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.LessonProgressMutationResponse{}, ErrLearningLessonNotFound
		}
		return models.LessonProgressMutationResponse{}, err
	}
	module, err := s.db.Read().GetCourseModuleByID(ctx, lesson.ModuleID)
	if err != nil {
		return models.LessonProgressMutationResponse{}, err
	}

	assignment, err := s.db.Read().GetCourseAssignmentByUserCourse(ctx, db.GetCourseAssignmentByUserCourseParams{
		UserID:   uuid.NullUUID{UUID: uid, Valid: true},
		CourseID: module.CourseID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.LessonProgressMutationResponse{}, ErrLearningAssignmentMissing
		}
		return models.LessonProgressMutationResponse{}, err
	}

	existing, err := s.db.Read().GetLessonProgressByAssignmentLesson(ctx, db.GetLessonProgressByAssignmentLessonParams{
		AssignmentID: assignment.ID,
		LessonID:     lessonUUID,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return models.LessonProgressMutationResponse{}, err
	}

	isCompleted := false
	lastPlayback := int64(0)
	completedAt := sql.NullTime{}
	if err == nil {
		isCompleted = existing.IsCompleted
		lastPlayback = existing.LastPlaybackPosition
		completedAt = existing.CompletedAt
	}
	if dto.IsCompleted != nil {
		isCompleted = *dto.IsCompleted
		if isCompleted {
			completedAt = sql.NullTime{Time: time.Now(), Valid: true}
		} else {
			completedAt = sql.NullTime{}
		}
	}
	if dto.LastPlaybackPosition != nil {
		lastPlayback = *dto.LastPlaybackPosition
	}

	progress, err := s.db.Write().UpsertLessonProgress(ctx, db.UpsertLessonProgressParams{
		ID:                   uuid.New(),
		IsCompleted:          isCompleted,
		LastPlaybackPosition: lastPlayback,
		CompletedAt:          completedAt,
		AssignmentID:         assignment.ID,
		LessonID:             lessonUUID,
	})
	if err != nil {
		return models.LessonProgressMutationResponse{}, err
	}

	updatedAssignment, err := s.recomputeAssignmentProgress(ctx, assignment.ID, module.CourseID)
	if err != nil {
		return models.LessonProgressMutationResponse{}, err
	}

	return toProgressMutationResponse(progress, updatedAssignment), nil
}

func (s *service) HeartbeatLessonProgress(ctx context.Context, lessonID string, userID string, dto models.HeartbeatProgressRequest) (models.LessonProgressMutationResponse, error) {
	lessonUUID, err := uuid.Parse(lessonID)
	if err != nil {
		return models.LessonProgressMutationResponse{}, errors.New("invalid lesson ID format")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return models.LessonProgressMutationResponse{}, errors.New("invalid user ID format")
	}

	lesson, err := s.db.Read().GetLessonByID(ctx, lessonUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.LessonProgressMutationResponse{}, ErrLearningLessonNotFound
		}
		return models.LessonProgressMutationResponse{}, err
	}
	module, err := s.db.Read().GetCourseModuleByID(ctx, lesson.ModuleID)
	if err != nil {
		return models.LessonProgressMutationResponse{}, err
	}

	assignment, err := s.db.Read().GetCourseAssignmentByUserCourse(ctx, db.GetCourseAssignmentByUserCourseParams{
		UserID:   uuid.NullUUID{UUID: uid, Valid: true},
		CourseID: module.CourseID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.LessonProgressMutationResponse{}, ErrLearningAssignmentMissing
		}
		return models.LessonProgressMutationResponse{}, err
	}

	progress, err := s.db.Write().UpsertLessonProgressHeartbeat(ctx, db.UpsertLessonProgressHeartbeatParams{
		ID:                   uuid.New(),
		LastPlaybackPosition: dto.LastPlaybackPosition,
		AssignmentID:         assignment.ID,
		LessonID:             lessonUUID,
	})
	if err != nil {
		return models.LessonProgressMutationResponse{}, err
	}

	updatedAssignment, err := s.recomputeAssignmentProgress(ctx, assignment.ID, module.CourseID)
	if err != nil {
		return models.LessonProgressMutationResponse{}, err
	}

	return toProgressMutationResponse(progress, updatedAssignment), nil
}

func (s *service) CompleteLesson(ctx context.Context, lessonID string, userID string, dto models.CompleteLessonRequest) (models.LessonProgressMutationResponse, error) {
	lessonUUID, err := uuid.Parse(lessonID)
	if err != nil {
		return models.LessonProgressMutationResponse{}, errors.New("invalid lesson ID format")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return models.LessonProgressMutationResponse{}, errors.New("invalid user ID format")
	}

	lesson, err := s.db.Read().GetLessonByID(ctx, lessonUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.LessonProgressMutationResponse{}, ErrLearningLessonNotFound
		}
		return models.LessonProgressMutationResponse{}, err
	}
	module, err := s.db.Read().GetCourseModuleByID(ctx, lesson.ModuleID)
	if err != nil {
		return models.LessonProgressMutationResponse{}, err
	}
	course, err := s.db.Read().GetCourseByID(ctx, module.CourseID)
	if err != nil {
		return models.LessonProgressMutationResponse{}, err
	}

	assignment, err := s.db.Read().GetCourseAssignmentByUserCourse(ctx, db.GetCourseAssignmentByUserCourseParams{
		UserID:   uuid.NullUUID{UUID: uid, Valid: true},
		CourseID: module.CourseID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.LessonProgressMutationResponse{}, ErrLearningAssignmentMissing
		}
		return models.LessonProgressMutationResponse{}, err
	}

	existing, err := s.db.Read().GetLessonProgressByAssignmentLesson(ctx, db.GetLessonProgressByAssignmentLessonParams{
		AssignmentID: assignment.ID,
		LessonID:     lessonUUID,
	})
	if err == nil && existing.IsCompleted {
		return models.LessonProgressMutationResponse{}, ErrLearningAlreadyCompleted
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return models.LessonProgressMutationResponse{}, err
	}

	if !isEligibleForCompletion(lesson.ContentType, dto) {
		return models.LessonProgressMutationResponse{}, ErrLearningThresholdFailed
	}

	result, err := s.UpdateLessonProgress(ctx, lessonID, userID, models.UpdateProgressRequest{IsCompleted: boolPtr(true)})
	if err != nil {
		return models.LessonProgressMutationResponse{}, err
	}

	user, _ := s.db.Read().GetUserByID(ctx, uid)
	learnerName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)

	_, _ = s.notifications.PublishHrEvent(ctx, notifications.HrEvent{
		Type:    models.HrNotificationLessonCompleted,
		Title:   "Lesson Completed",
		Message: fmt.Sprintf("%s completed \"%s\".", learnerName, lesson.Title),
		Payload: map[string]any{
			"lessonId":           lesson.ID.String(),
			"lessonTitle":        lesson.Title,
			"contentType":        string(lesson.ContentType),
			"moduleId":           lesson.ModuleID.String(),
			"courseId":           module.CourseID.String(),
			"assignmentId":       assignment.ID.String(),
			"progressPercentage": result.Assignment.ProgressPercentage,
		},
	})

	if result.Assignment.Status == models.AssignmentCompleted {
		_, _ = s.notifications.PublishHrEvent(ctx, notifications.HrEvent{
			Type:    models.HrNotificationCourseCompleted,
			Title:   "Course Completed",
			Message: fmt.Sprintf("%s has completed the full course \"%s\".", learnerName, course.Title),
			Payload: map[string]any{
				"courseId":      module.CourseID.String(),
				"courseTitle":   course.Title,
				"assignmentId":  assignment.ID.String(),
				"completedAt":   time.Now().Format(time.RFC3339),
				"learnerUserId": uid.String(),
			},
		})
	}

	role := "EMPLOYEE"
	if user.Role == models.RoleManager {
		role = "MANAGER"
	}
	courseIDStr := module.CourseID.String()
	s.notifications.PublishDashboardSyncForUser(ctx, uid.String(), notifications.DashboardScope(role), "learning-progress-updated", &courseIDStr)

	return result, nil
}

func (s *service) recomputeAssignmentProgress(ctx context.Context, assignmentID uuid.UUID, courseID uuid.UUID) (db.CourseAssignment, error) {
	totalLessons, err := s.db.Read().CountCourseLessons(ctx, courseID)
	if err != nil {
		return db.CourseAssignment{}, err
	}
	completedLessons, err := s.db.Read().CountCompletedLessonProgressByAssignment(ctx, assignmentID)
	if err != nil {
		return db.CourseAssignment{}, err
	}
	startedLessons, err := s.db.Read().CountStartedLessonProgressByAssignment(ctx, assignmentID)
	if err != nil {
		return db.CourseAssignment{}, err
	}

	percentage := float64(0)
	if totalLessons > 0 {
		percentage = math.Round((float64(completedLessons) / float64(totalLessons)) * 100)
	}

	status := models.AssignmentNotStarted
	if percentage >= 100 {
		status = models.AssignmentCompleted
	} else if startedLessons > 0 {
		status = models.AssignmentInProgress
	}

	completedAt := sql.NullTime{}
	if status == models.AssignmentCompleted {
		completedAt = sql.NullTime{Time: time.Now(), Valid: true}
	}

	return s.db.Write().UpdateCourseAssignmentProgress(ctx, db.UpdateCourseAssignmentProgressParams{
		ID:                 assignmentID,
		Status:             status,
		ProgressPercentage: percentage,
		CompletedAt:        completedAt,
		DueDate:            sql.NullTime{},
	})
}

func (s *service) mapAssignmentSummary(ctx context.Context, assignment db.CourseAssignment) (models.AssignmentSummaryResponse, error) {
	course, err := s.db.Read().GetCourseByID(ctx, assignment.CourseID)
	if err != nil {
		return models.AssignmentSummaryResponse{}, err
	}

	moduleCount, err := s.db.Read().CountCourseModules(ctx, assignment.CourseID)
	if err != nil {
		return models.AssignmentSummaryResponse{}, err
	}
	assignmentCount, err := s.db.Read().CountCourseAssignmentsByCourseID(ctx, assignment.CourseID)
	if err != nil {
		return models.AssignmentSummaryResponse{}, err
	}

	resp := models.AssignmentSummaryResponse{
		ID:                 assignment.ID.String(),
		Status:             assignment.Status,
		ProgressPercentage: assignment.ProgressPercentage,
		CourseVersion:      assignment.CourseVersion,
		DueDate:            nullTimePtr(assignment.DueDate),
		EnrolledAt:         assignment.EnrolledAt.Format(time.RFC3339),
		CompletedAt:        nullTimePtr(assignment.CompletedAt),
		Course: models.AssignmentCourseSummaryResponse{
			ID:                course.ID.String(),
			Title:             course.Title,
			Description:       nullStringPtr(course.Description),
			CoverImageURL:     nullStringPtr(course.CoverImageUrl),
			Category:          course.Category,
			EstimatedDuration: nullInt64Ptr(course.EstimatedDuration),
			LearningOutcomes:  parseOutcomes(course.LearningOutcomes),
			Version:           course.Version,
			Count: models.CourseCountResponse{
				Modules:     moduleCount,
				Assignments: assignmentCount,
			},
		},
	}

	if course.AuthorID.Valid {
		author, err := s.db.Read().GetUserByID(ctx, course.AuthorID.UUID)
		if err == nil {
			resp.Course.Author = &models.CourseAuthorResponse{ID: author.ID.String(), FirstName: author.FirstName, LastName: author.LastName}
		}
	}

	if assignment.UserID.Valid {
		user, err := s.db.Read().GetUserByID(ctx, assignment.UserID.UUID)
		if err == nil {
			u := models.UserResponse{
				ID:        user.ID.String(),
				PesNumber: user.PesNumber,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Email:     user.Email,
				Role:      user.Role,
			}
			resp.User = &u
		}
	}

	if assignment.AssignedByID.Valid {
		assigner, err := s.db.Read().GetUserByID(ctx, assignment.AssignedByID.UUID)
		if err == nil {
			resp.AssignedBy = &models.CourseAuthorResponse{ID: assigner.ID.String(), FirstName: assigner.FirstName, LastName: assigner.LastName}
		}
	}

	return resp, nil
}

func mapPublishedCourseSummary(row db.ListPublishedCoursesRow) models.AssignmentCourseSummaryResponse {
	resp := models.AssignmentCourseSummaryResponse{
		ID:                row.ID.String(),
		Title:             row.Title,
		Description:       nullStringPtr(row.Description),
		CoverImageURL:     nullStringPtr(row.CoverImageUrl),
		Category:          row.Category,
		EstimatedDuration: nullInt64Ptr(row.EstimatedDuration),
		LearningOutcomes:  parseOutcomes(row.LearningOutcomes),
		Version:           row.Version,
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

func toProgressMutationResponse(progress db.LessonProgress, assignment db.CourseAssignment) models.LessonProgressMutationResponse {
	resp := models.LessonProgressMutationResponse{}
	resp.Progress.ID = progress.ID.String()
	resp.Progress.IsCompleted = progress.IsCompleted
	resp.Progress.LastPlaybackPosition = progress.LastPlaybackPosition
	resp.Progress.CompletedAt = nullTimePtr(progress.CompletedAt)
	resp.Assignment.ID = assignment.ID.String()
	resp.Assignment.Status = assignment.Status
	resp.Assignment.ProgressPercentage = assignment.ProgressPercentage
	return resp
}

func isEligibleForCompletion(contentType models.LessonContentType, dto models.CompleteLessonRequest) bool {
	viewedSeconds := int64(0)
	if dto.ViewedSeconds != nil {
		viewedSeconds = *dto.ViewedSeconds
	}

	switch contentType {
	case models.LessonVideo:
		return dto.WatchedPercent != nil && *dto.WatchedPercent >= thresholds.videoMinPercent
	case models.LessonAudio:
		return dto.ListenedPercent != nil && *dto.ListenedPercent >= thresholds.audioMinPercent
	case models.LessonRichText:
		return dto.ScrolledPercent != nil && *dto.ScrolledPercent >= thresholds.richTextMinScrollPercent && viewedSeconds >= thresholds.richTextMinSeconds
	case models.LessonPdf:
		return dto.ReachedLastPage != nil && *dto.ReachedLastPage && viewedSeconds >= thresholds.pdfMinSeconds
	case models.LessonImage:
		return dto.OpenedInLightbox != nil && *dto.OpenedInLightbox && viewedSeconds >= thresholds.imageMinSeconds
	case models.LessonPresentation:
		return viewedSeconds >= thresholds.presentationMinSeconds
	default:
		return false
	}
}

func nullTimePtr(v sql.NullTime) *string {
	if !v.Valid {
		return nil
	}
	val := v.Time.Format(time.RFC3339)
	return &val
}

func nullStringPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	val := v.String
	return &val
}

func nullInt64Ptr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	val := v.Int64
	return &val
}

func boolPtr(v bool) *bool {
	return &v
}

func parseOutcomes(raw string) []string {
	if raw == "" {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return []string{}
	}
	return out
}
