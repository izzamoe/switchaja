package main

import (
	"log"
	"switchiot/internal/app"
	"switchiot/internal/config"
	"switchiot/internal/server"

	_ "modernc.org/sqlite"
)

// main boots the application using clean architecture principles
func main() {
	// Load configuration
	cfg := config.LoadConfig()
	cfg.NormalizePort()

	// Initialize application with all dependencies
	application, err := app.NewApplication(cfg)
	if err != nil {
		log.Fatal("Failed to initialize application:", err)
	}
	defer application.Close()

	// Create and configure HTTP server
	httpServer := server.NewServer(application)

	// Start background tasks
	httpServer.StartBackgroundTasks()

	// Start the HTTP server
	if err := httpServer.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
