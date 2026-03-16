// Package database provides a service for interacting with the database and checking its health.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "modernc.org/sqlite" // pure Go sqlite3 driver, no need for CGO, works across platforms and is easy to set up

	"go-server/internal/database/db"
)

// Service represents a service that interacts with a database
type Service interface {
	// Health returns a map of health status information
	Health() map[string]string

	// Close terminated the database connection
	Close() error

	// Funcs to give modules access to queries
	Read() *db.Queries
	Write() *db.Queries
}

type service struct {
	writer *sql.DB
	reader *sql.DB
	dburl  string
}

var dbInstance *service

func New() Service {
	// Reuse Connection: singalton pattern
	if dbInstance != nil {
		return dbInstance
	}

	dburl := os.Getenv("DB_URL")
	if dburl == "" {
		log.Fatal("DB_URL environment variable is not set")
	}

	// Making dedicated writerPool connection to the database for writing data to the db
	writerPool, err := sql.Open("sqlite", dburl)
	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initilaization error
		log.Fatalf("failed to initialize database writer connection: %v", err)
	}

	// Performance Tuning: SQLite handles concurrent writes better with 1 connection
	// Writes should not be done concurrently, so we set max open connections to 1 to avoid "database is locked" errors
	writerPool.SetMaxOpenConns(1)

	// Making dedicated readerPool connection to the database for reading data from the db
	readerPool, err := sql.Open("sqlite", dburl)
	if err != nil {
		log.Fatalf("failed to initialize database reader connection: %v", err)
	}
	// Allow up to 10 concurrent read connections, as SQLite can handle multiple readers
	readerPool.SetMaxOpenConns(10)

	dbInstance = &service{
		writer: writerPool,
		reader: readerPool,
		dburl:  dburl,
	}

	return dbInstance
}

// Implementing the functions needed by the DB_Service interface
// So that service struct implements that interface
func (s *service) Health() map[string]string {
	log.Println("Checking the health of the database...")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.reader.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		// We don't want to terminate the whole program here, just return the health status
		return stats
	}

	// Database is up, add more stats
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// DB status like open connections, in use, idle, etc
	dbStats := s.reader.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["is_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluating database stats to provide a health message
	if dbStats.OpenConnections > 40 {
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	log.Println(stats["message"])

	return stats
}

// Close closes the connection to the database
// It logs a message indicating the disconnection from the specific database
// If success, returns nil, else error
func (s *service) Close() error {
	log.Printf("Shutting down database pools for: %s", s.dburl)

	// Close the write pool
	errWriter := s.writer.Close()
	if errWriter != nil {
		return errWriter
	}

	// Close the read pool
	errReader := s.reader.Close()
	if errReader != nil {
		return errReader
	}

	log.Printf("Disconnected from database: %s", s.dburl)
	return nil
}

// We do not want to return *sql.DB to the moudules,
// We want to return proper typed db.Queries
func (s *service) Read() *db.Queries {
	return db.New(s.reader) // instead of seeding the sql.DB conn to the modules, we instead all the queries right here and send back the queries
}

func (s *service) Write() *db.Queries {
	return db.New(s.writer)
}
