// Package Admin contains all the handlers and services related to admin management.
package admin

import (
	"context"

	"go-server/internal/database"
	"go-server/internal/models"
)

type Service interface {
	GetKpis(ctx context.Context) (models.AdminKpisResponse, error)
	GetMonthlyStats(ctx context.Context) ([]models.MonthlyStatsResponse, error)
	GetCategoryDistribution(ctx context.Context) ([]models.CategoryDistributionResponse, error)
	GetClusterStats(ctx context.Context) ([]models.ClusterStatsResponse, error)
}

type service struct {
	db database.Service
}

func NewService(db database.Service) Service {
	return &service{
		db: db,
	}
}

func (s *service) GetKpis(ctx context.Context) (models.AdminKpisResponse, error) {
	data, err := s.db.Read().GetAdminKpis(ctx)
	if err != nil {
		return models.AdminKpisResponse{}, err
	}

	var totalManDays float64
	if data.TotalManDays != nil {
		totalManDays = data.TotalManDays.(float64)
	}

	return models.AdminKpisResponse{
		TotalTrainings:    data.TotalTrainings,
		TotalParticipants: data.TotalParticipants,
		CompletedCount:    data.CompletedCount,
		EnrolledCount:     data.EnrolledCount,
		TotalManDays:      totalManDays,
	}, nil
}

func (s *service) GetMonthlyStats(ctx context.Context) ([]models.MonthlyStatsResponse, error) {
	data, err := s.db.Read().GetMonthlyStats(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]models.MonthlyStatsResponse, len(data))

	for i, d := range data {
		// fmt.Printf("%+v\n", d)
		var monthKey, monthLabel string
		if d.MonthKey != nil {
			monthKey = d.MonthKey.(string)
		}
		if d.MonthLabel != nil {
			monthLabel = d.MonthLabel.(string)
		}

		results[i] = models.MonthlyStatsResponse{
			MonthKey:     monthKey,
			MonthLabel:   monthLabel,
			Participants: d.ParticipantCount,
			Trainings:    d.TrainingCount,
		}
	}

	return results, nil
}

func (s *service) GetCategoryDistribution(ctx context.Context) ([]models.CategoryDistributionResponse, error) {
	data, err := s.db.Read().GetCategoryDistribution(ctx)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("%v+", data)
	results := make([]models.CategoryDistributionResponse, len(data))

	for i, cat := range data {
		results[i] = models.CategoryDistributionResponse{
			Name:  cat.Name,
			Value: cat.Value,
		}
	}

	return results, nil
}

func (s *service) GetClusterStats(ctx context.Context) ([]models.ClusterStatsResponse, error) {
	data, err := s.db.Read().GetClusterStats(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]models.ClusterStatsResponse, len(data))
	for i, c := range data {
		var untrained int64
		if c.Untrained != nil {
			untrained = c.Untrained.(int64)
		}

		results[i] = models.ClusterStatsResponse{
			Cluster:        c.Cluster,
			TotalEmployees: c.TotalEmployees,
			Trained:        c.Trained,
			Untrained:      untrained,
		}
	}

	return results, nil
}
