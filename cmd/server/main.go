package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/messenger/backend/internal/api/handlers"
	"github.com/messenger/backend/internal/services"
	"github.com/messenger/backend/internal/storage/memory"
	"github.com/oklog/ulid/v2"
)

func main() {
	// 1. Initialize Router
	router := gin.Default()

	// 2. Setup Dependencies
	// In a real app, you would initialize a real database repository here.
	contactRepo := memory.NewInMemoryContactRepository()
	contactsService := services.NewContactsService(contactRepo)
	contactsHandler := handlers.NewContactsHandler(contactsService)

	// 3. Setup Middleware
	// This is a dummy middleware for testing. It injects a hardcoded user ID.
	// In a real app, this would be your JWT/PASETO authentication middleware.
	router.Use(func(c *gin.Context) {
		// Hardcoded user ID for user1 from the in-memory repo
		testUserID := ulid.MustParse("01H8XGJWBWBAQ1JBS9M6S3S2A1")
		c.Set("userID", testUserID)
		c.Next()
	})

	// 4. Register Routes
	v1 := router.Group("/api/v1")
	{
		contactsHandler.RegisterContactRoutes(v1)
		// Other handlers would be registered here
	}

	// 5. Start Server
	log.Println("Starting server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
