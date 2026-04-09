// Package Admin contains all the handlers and services related to admin management.
package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"go-server/internal/database"
	"go-server/internal/database/db"
	"go-server/internal/models"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type Service interface {
	GetKpis(ctx context.Context) (models.AdminKpisResponse, error)
	GetMonthlyStats(ctx context.Context) ([]models.MonthlyStatsResponse, error)
	GetCategoryDistribution(ctx context.Context) ([]models.CategoryDistributionResponse, error)
	GetClusterStats(ctx context.Context) ([]models.ClusterStatsResponse, error)
	ImportHistoricalWorkbook(ctx context.Context, workbook io.Reader, originalName string) (models.ImportResponse, error)
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
		if val, ok := data.TotalManDays.(float64); ok {
			totalManDays = val
		}
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
		var monthKey, monthLabel string
		if d.MonthKey != nil {
			if val, ok := d.MonthKey.(string); ok {
				monthKey = val
			}
		}
		if d.MonthLabel != nil {
			if val, ok := d.MonthLabel.(string); ok {
				monthLabel = val
			}
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
			if val, ok := c.Untrained.(int64); ok {
				untrained = val
			}
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

func (s *service) ImportHistoricalWorkbook(ctx context.Context, workbook io.Reader, originalName string) (models.ImportResponse, error) {
	// 1. Open the Excel file direcly from the memory stream
	f, err := excelize.OpenReader(workbook)
	if err != nil {
		return models.ImportResponse{}, fmt.Errorf("failed to read excel file: %w", err)
	}
	defer f.Close()

	// 2. Get rows from the first sheet
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil || len(rows) <= 1 { // <= 1: assumes you have a header row
		return models.ImportResponse{}, errors.New("no valid historical rows found in the workbook")
	}

	// Using maps as sets to count unique values
	uniquePrograms := make(map[string]struct{})
	uniqueMonths := make(map[string]struct{})

	importedCount := 0

	// 3. The database transaction
	err = s.db.ExecTx(ctx, func(qtx *db.Queries) error {
		// A. wipe the old data
		if err := qtx.DeleteAllHistoricalRecords(ctx); err != nil {
			return err
		}

		// B. Loop through the excel rows (skipping index 0 if it's headers)
		for i := 1; i < len(rows); i++ {
			row := rows[i]

			// safety check: ensure row has enough columns before indexing
			if len(row) < 5 { // at least need some basic info
				continue
			}

			// parsing cols with safe access
			getCol := func(idx int) string {
				if idx < len(row) {
					return row[idx]
				}
				return ""
			}

			programIDStr := getCol(0)
			programTitle := getCol(1)
			mappedCategory := getCol(2)
			cluster := getCol(3)
			pesNo := getCol(4)
			empName := getCol(5)
			status := getCol(6)
			mode := getCol(7)
			fromDateStr := getCol(8)
			toDateStr := getCol(9)
			monthKey := getCol(10)
			manDaysStr := getCol(11)
			manHoursStr := getCol(12)
			costStr := getCol(13)

			// Parsing ID/UUID
			programID, _ := uuid.Parse(programIDStr)

			// Parsing Dates
			fromDate := parseExcelDate(fromDateStr)
			toDate := parseExcelDate(toDateStr)

			// Parsing Floats/Ints
			manDays, _ := strconv.ParseFloat(manDaysStr, 64)
			manHours, _ := strconv.ParseFloat(manHoursStr, 64)
			cost, _ := strconv.ParseInt(costStr, 10, 64)

			// insert into SQLite via tx
			err := qtx.CreateHistoricalRecord(ctx, db.CreateHistoricalRecordParams{
				ID:               uuid.New(),
				ProgramID:        programID,
				ProgramTitle:     programTitle,
				MappedCategory:   sql.NullString{String: mappedCategory, Valid: mappedCategory != ""},
				Cluster:          sql.NullString{String: cluster, Valid: cluster != ""},
				EmployeePesNo:    sql.NullString{String: pesNo, Valid: pesNo != ""},
				EmployeeName:     sql.NullString{String: empName, Valid: empName != ""},
				CompletionStatus: sql.NullString{String: status, Valid: status != ""},
				ModeOfDelivery:   sql.NullString{String: mode, Valid: mode != ""},
				FromDate:         sql.NullTime{Time: fromDate, Valid: !fromDate.IsZero()},
				ToDate:           sql.NullTime{Time: toDate, Valid: !toDate.IsZero()},
				MonthKey:         monthKey,
				ManDays:          sql.NullFloat64{Float64: manDays, Valid: manDaysStr != ""},
				ManHours:         sql.NullFloat64{Float64: manHours, Valid: manHoursStr != ""},
				TotalCostInr:     sql.NullInt64{Int64: cost, Valid: costStr != ""},
				SourceFile:       sql.NullString{String: originalName, Valid: true},
			})

			if err != nil {
				return fmt.Errorf("failed to insert row %d: %w", i, err)
			}

			// D. Track unique stats
			programKey := fmt.Sprintf("%s-%s-%s-%s", programIDStr, programTitle, fromDateStr, toDateStr)
			uniquePrograms[programKey] = struct{}{}
			uniqueMonths[monthKey] = struct{}{}
			importedCount++
		}

		return nil // commit the massive batch of inserts
	})

	if err != nil {
		return models.ImportResponse{}, fmt.Errorf("import transaction failed: %w", err)
	}

	// 4. Return
	return models.ImportResponse{
		TotalRows:       int64(len(rows) - 1),
		Imported:        int64(importedCount),
		UniqueTrainings: int64(len(uniquePrograms)),
		MonthCoverage:   int64(len(uniqueMonths)),
	}, nil
}

// Helper to parse excel dates which might be in different formats
func parseExcelDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}
	// Try common formats
	formats := []string{
		"2006-01-02",
		"01-02-06",
		"01-02-2006",
		"02-Jan-2006",
		"1/2/06",
		"1/2/2006",
		time.RFC3339,
	}
	for _, f := range formats {
		if t, err := time.Parse(f, dateStr); err == nil {
			return t
		}
	}
	return time.Time{}
}
