package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // Standard library driver for goose
	"github.com/messenger/backend/internal/db"
	"github.com/pressly/goose/v3"
)

var testQueries *db.Queries
var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	// In a CI/CD environment, this would be configured via env vars.
	// For this sandbox, we'll use hardcoded defaults assuming a local Docker setup.
	dbUser := "postgres"
	dbPassword := "postgres"
	dbHost := "localhost" // or "postgres" if running in a docker-compose network
	dbPort := "5432"
	dbName := "messenger_test"

	testDbDsn := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	// Override with env var if provided
	if envDsn := os.Getenv("TEST_DATABASE_URL"); envDsn != "" {
		testDbDsn = envDsn
	}

	ctx := context.Background()
	var err error
	testPool, err = pgxpool.New(ctx, testDbDsn)
	if err != nil {
		log.Fatalf("Could not connect to test database: %s", err)
	}
	log.Println("Test database pool created successfully.")

	dbConn, err := goose.OpenDBWithDriver("pgx", testDbDsn)
	if err != nil {
		log.Fatalf("goose: failed to open DB: %v\n", err)
	}

	// It's good practice to bring the db down before up, to reset state.
	if err := goose.DownTo(dbConn, "../db/migrations", 0); err != nil {
		log.Fatalf("goose: down failed: %v", err)
	}
	if err := goose.Up(dbConn, "../db/migrations"); err != nil {
		log.Fatalf("goose: up failed: %v", err)
	}
	if err := dbConn.Close(); err != nil {
		log.Fatalf("goose: failed to close DB: %v\n", err)
	}
	log.Println("Test database migrations applied successfully.")

	testQueries = db.New(testPool)

	exitCode := m.Run()

	testPool.Close()
	os.Exit(exitCode)
}

func truncateTables(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return fmt.Errorf("test database pool is nil")
	}
	tables := []string{
		"blocks",
		"contact_requests",
		"contacts",
		"auth_sessions",
		"devices",
		"users",
	}
	for _, table := range tables {
		if _, err := pool.Exec(ctx, "TRUNCATE TABLE "+table+" RESTART IDENTITY CASCADE"); err != nil {
			return err
		}
	}
	return nil
}
