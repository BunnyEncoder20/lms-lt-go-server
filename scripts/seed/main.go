package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"go-server/internal/database/db"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func main() {
	log.Println("🌱 Seeding database...")

	// Connecting to the database
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

	// Hash the common password for all users
	defaultPassword := os.Getenv("DEFAULT_PASSWORD")
	if defaultPassword == "" {
		log.Println(".env file not found or the DEFAULT_PASSWORD was not set. Using fallback")
		defaultPassword = "password123" // incase the env is not there
	}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	defaultPassword = string(hashedPassword)

	// Create an admin
	_, err = queries.CreateUser(ctx, db.CreateUserParams{
		ID:        seed[0],
		PesNumber: "PES-001",
		Password:  defaultPassword,
		FirstName: "Administrator",
		LastName:  "HR",
		Email:     "admin@hrlms.com",
		Role:      "ADMIN",
		Cluster:   sql.NullString{String: "Technology", Valid: true}, // Handle nullable fields
		ManagerID: uuid.NullUUID{Valid: false},                       // No manager for top admin
	})
	if err != nil {
		log.Printf("Could not seed admin (may already exists): %v", err)
	} else {
		log.Println("✅ Admin created")
	}

	// Seed Manager
	_, err = queries.CreateUser(ctx, db.CreateUserParams{
		ID:        seed[1],
		PesNumber: "PES-002",
		Password:  defaultPassword,
	})

	log.Println("🌱 Database seeded successfully with default password")
}
