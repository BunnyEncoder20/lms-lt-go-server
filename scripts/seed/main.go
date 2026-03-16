package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"go-server/internal/database/db"
	"go-server/internal/models"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

// Helper to hash passwords quickly during seed
func hashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes)
}

func main() {
	log.Println("🌱 Seeding database...")

	// Connecting to the database (we use ./.env cause the makefile commands runs from the root of the project)
	if err := godotenv.Load("./.env"); err != nil {
		log.Println("❗️ WARN: could not load .env file, DB_URL is likely empty", err)
	}
	conn, err := sql.Open("sqlite", os.Getenv("DB_URL"))
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer conn.Close()

	// Initial sqlc quereis object
	queries := db.New(conn)
	ctx := context.Background()

	// Fixed UUIDs so upserts are idempotent across multiple seed runs

	seed := [11]uuid.UUID{
		uuid.MustParse("e83e0a1b-a580-41d6-96ef-79bc3b476781"),
		uuid.MustParse("017ee8c7-9424-4a38-bca0-34add327c28a"),
		uuid.MustParse("e0959662-e6d1-4c0d-9676-59c119e47fda"),
		uuid.MustParse("a775afc2-8c9e-4eda-9c5e-cb3e1220d7f1"),
		uuid.MustParse("d47773c4-12d1-4570-99a4-50f1bd11fb7a"),
		uuid.MustParse("062d05b9-475b-4ba2-8a33-452c663747b1"),
		uuid.MustParse("3ee57473-98de-4d8e-a052-3c375d67053a"),
		uuid.MustParse("1a23fed8-cb9e-4e98-a0a9-19a7acb1cc74"),
		uuid.MustParse("74ae6a28-dafd-484f-a3d2-5ddaeaba6c06"),
		uuid.MustParse("70b2c025-5b06-48aa-842f-2f8a373c3bc9"),
		uuid.MustParse("4711849d-b0ab-4630-824e-379db3855cbc"),
	}

	// Create an admin
	admin, err := queries.CreateUser(ctx, db.CreateUserParams{
		ID:        seed[0],
		PesNumber: "PES-001",
		Password:  hashPassword(os.Getenv("DEFAULT_PASSWORD")),
		FirstName: "Administrator",
		LastName:  "HR",
		Email:     "admin@hrlms.com",
		Role:      models.RoleAdmin,
		Cluster:   sql.NullString{String: "Technology", Valid: true}, // Handle nullable fields
		ManagerID: uuid.NullUUID{Valid: false},                       // No manager for top admin
	})
	if err != nil {
		log.Printf("❌ Failed to seed admin (may already exists): %v", err)
	} else {
		log.Println("✅ Admin created")
	}

	// Seed Managers
	manager1, err := queries.CreateUser(ctx, db.CreateUserParams{
		ID:        seed[1],
		PesNumber: "PES-002",
		Password:  hashPassword("manager123"),
		FirstName: "James",
		LastName:  "Wilson",
		Email:     "manager@skillsync.com",
		Role:      models.RoleManager,
		Cluster:   sql.NullString{String: "Operations", Valid: true},
		ManagerID: uuid.NullUUID{UUID: admin.ID},
	})
	if err != nil {
		log.Printf("❌ Failed to seed manager1: %v", err)
	} else {
		log.Printf("✅ Seeded Manager1")
	}

	manager2, err := queries.CreateUser(ctx, db.CreateUserParams{
		ID:        seed[2],
		PesNumber: "PES-005",
		Password:  hashPassword("manager123"),
		FirstName: "Gergory",
		LastName:  "House",
		Email:     "gergory@skillsync.com",
		Role:      models.RoleManager,
		Cluster:   sql.NullString{String: "HR", Valid: true},
		ManagerID: uuid.NullUUID{UUID: admin.ID, Valid: true},
	})
	if err != nil {
		log.Printf("❌ Failed to seed manager2: %v", err)
	} else {
		log.Printf("✅ Seeded Manager2")
	}

	// 3. Employees (Referencing Managers)
	employees := []db.CreateUserParams{
		{
			ID:        seed[3],
			PesNumber: "PES-003",
			Password:  hashPassword("employee123"),
			FirstName: "Sarah",
			LastName:  "Johnson",
			Email:     "sarah@skillsync.com",
			Role:      models.RoleEmployee,
			Cluster:   sql.NullString{String: "Operations", Valid: true},
			ManagerID: uuid.NullUUID{UUID: manager1.ID, Valid: true},
		},
		{
			ID:        seed[4],
			PesNumber: "PES-004",
			Password:  hashPassword("employee123"),
			FirstName: "Michael",
			LastName:  "Chen",
			Email:     "michael@skillsync.com",
			Role:      models.RoleEmployee,
			Cluster:   sql.NullString{String: "Operations", Valid: true},
			ManagerID: uuid.NullUUID{UUID: manager1.ID, Valid: true},
		},
		{
			ID:        seed[5],
			PesNumber: "PES-006",
			Password:  hashPassword("employee123"),
			FirstName: "Priya",
			LastName:  "Sharma",
			Email:     "priya@skillsync.com",
			Role:      models.RoleEmployee,
			Cluster:   sql.NullString{String: "HR", Valid: true},
			ManagerID: uuid.NullUUID{UUID: manager2.ID, Valid: true},
		},
	}

	for _, emp := range employees {
		_, err := queries.CreateUser(ctx, emp)
		if err != nil {
			log.Printf("❌ Failed to seed %s: %v", emp.Email, err)
		} else {
			log.Printf("✅ Seeded Employee: %s", emp.Email)
		}
	}

	log.Println("🌱 Seeding Trainings...")

	now := time.Now()

	// Helper to mimic your future/past JS functions
	offsetDate := func(days int) time.Time {
		return now.AddDate(0, 0, days)
	}

	trainings := []db.CreateTrainingParams{
		{
			ID:          seed[6],
			Title:       "Leadership Excellence Program",
			Description: sql.NullString{String: "Develop leadership and team management skills for managerial roles.", Valid: true},
			Category:    models.TrainingBehavioral,
			StartDate:   offsetDate(10),
			EndDate:     offsetDate(11),
			Location:    sql.NullString{String: "Conference Room A", Valid: true},
			PreReadUri:  sql.NullString{String: "https://example.com/leadership-prework.pdf", Valid: true},
			CreatedByID: admin.ID,
		},
		{
			ID:          uuid.New(),
			Title:       "Advanced Excel & Data Analysis",
			Description: sql.NullString{String: "Master pivot tables, VLOOKUP, and data visualization.", Valid: true},
			Category:    models.TrainingTechnical,
			StartDate:   offsetDate(15),
			EndDate:     offsetDate(15),
			Location:    sql.NullString{String: "Training Lab 2", Valid: true},
			PreReadUri:  sql.NullString{String: "https://example.com/excel-guide.pdf", Valid: true},
			CreatedByID: admin.ID,
		},
		{
			ID:          uuid.New(),
			Title:       "Effective Communication",
			Description: sql.NullString{String: "Build confident public speaking skills.", Valid: true},
			Category:    models.TrainingBehavioral,
			StartDate:   offsetDate(20),
			EndDate:     offsetDate(21),
			VirtualLink: sql.NullString{String: "https://teams.microsoft.com/skillsync-comm", Valid: true},
			CreatedByID: admin.ID,
		},
		{
			ID:          uuid.New(),
			Title:       "Cybersecurity Awareness",
			Description: sql.NullString{String: "Mandatory security awareness training.", Valid: true},
			Category:    models.TrainingTechnical,
			StartDate:   offsetDate(5),
			EndDate:     offsetDate(5),
			Location:    sql.NullString{String: "Auditorium", Valid: true},
			PreReadUri:  sql.NullString{String: "https://example.com/cybersecurity-intro.pdf", Valid: true},
			CreatedByID: admin.ID,
		},
		{
			ID:          uuid.New(),
			Title:       "Project Management Fundamentals",
			Description: sql.NullString{String: "Introduction to Agile, Scrum, and PMBOK.", Valid: true},
			Category:    models.TrainingTechnical,
			StartDate:   offsetDate(-30),
			EndDate:     offsetDate(-29),
			Location:    sql.NullString{String: "Conference Room B", Valid: true},
			PreReadUri:  sql.NullString{String: "https://example.com/pm-basics.pdf", Valid: true},
			CreatedByID: admin.ID,
		},
	}

	for _, t := range trainings {
		_, err := queries.CreateTraining(ctx, t)
		if err != nil {
			log.Printf("❌ Failed to seed training %s: %v", t.Title, err)
		} else {
			log.Printf("✅ Seeded Training: %s", t.Title)
		}
	}

	log.Println("🌱 Database seed script completed")
}
