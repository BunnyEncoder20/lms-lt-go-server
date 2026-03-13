package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"go-server/internal/database"
	"go-server/internal/database/db"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func main() {
	log.Println("🌱 Seeding database...")

	// Connecting to the database
	ctx := context.Background()
	dbInstance := database.New()

	// Hash the common password for all users
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(os.Getenv("DEFAULT_PASSWORD")), bcrypt.DefaultCost)
	defaultPassword := string(hashedPassword)

	// Create an admin
	adminID := uuid.New()
	_, err := queries.CreateUser(ctx, dbInstance.CreateUserParams{
		ID:        adminID,
		PesNumber: "PES-001",
		Passowrd:  defaultPassword,
		FirstName: "System",
		LastName:  "Administrator",
		Email:     "admin@skillsync.com",
		Role:      "ADMIN",
		Cluster:   sql.NullString{String: "IT", Valid: true}, // Handle nullable fields
		ManagerID: uuid.NullUUID{Valid: false},               // No manager for top admin
	})
	if err != nil {
		log.Printf("Could not seed admin (may already exists): %v", err)
	} else {
		log.Println("✅ Admin created")
	}

	// Seed Manager
	_, err := queries.CreateUser(ctx, db.CreateUserParams{})

	log.Println("🌱 Database seeded successfully with default password")
}
