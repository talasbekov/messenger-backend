package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/messenger/backend/internal/api/handlers"
	"github.com/messenger/backend/internal/api/middleware"
	"github.com/messenger/backend/internal/config"
	"github.com/messenger/backend/internal/db"
	"github.com/messenger/backend/internal/services"
	"github.com/messenger/backend/internal/storage/postgres"
	"github.com/pressly/goose/v3"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Setup Database Connection
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

	// 3. Run Migrations
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to set goose dialect: %v", err)
	}
	dbConn, err := goose.OpenDBWithDriver("pgx", cfg.Database.DSN())
	if err != nil {
		log.Fatalf("goose: failed to open DB: %v\n", err)
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			log.Fatalf("goose: failed to close DB: %v\n", err)
		}
	}()

	if err := goose.Up(dbConn, "internal/db/migrations"); err != nil {
		log.Fatalf("goose: up failed: %v", err)
	}

	// 4. Setup Dependencies
	queries := db.New(pool)

	// Repositories
	contactRepo := postgres.NewPostgresContactRepository(queries)

	// Services
	authService := services.NewAuthService(queries, cfg.Auth, cfg.Security)
	contactsService := services.NewContactsService(contactRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	contactsHandler := handlers.NewContactsHandler(contactsService)

	// 5. Initialize Router
	router := gin.Default()

	// 6. Register Routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		authHandler.RegisterAuthRoutes(v1)

		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg.Auth))
		{
			contactsHandler.RegisterContactRoutes(protected)
			// Other protected handlers would be registered here
		}
	}

	// 7. Start Server
	log.Printf("Starting server on %s", cfg.Server.Port)
	if err := router.Run(cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
