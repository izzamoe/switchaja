package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"switchiot"
	"switchiot/internal/api"
	"switchiot/internal/db"
	"switchiot/internal/iot"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/websocket/v2"
	_ "modernc.org/sqlite"
)

var (
	buildDate = "unknown"
	commit    = "unknown"
)

func main() {
	log.Printf("HeheSwitch starting (build: %s, commit: %s)", buildDate, commit)

	// Database setup
	database, err := sql.Open("sqlite", "heheswitch.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer database.Close()

	// Initialize database schema with 8 consoles, 5000 price per hour
	if err := db.Init(database, 8, 5000); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// IoT components setup
	sender := iot.NewMockSender()
	hub := iot.NewHub()

	// API setup
	apiHandler := api.New(database, sender, hub)

	// Fiber app setup
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			log.Printf("Error: %v", err)
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// Static file serving using embedded files
	app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(switchiot.EmbeddedStatic),
		PathPrefix: "web/static",
		Browse:     false,
	}))

	// API routes
	apiHandler.Register(app)

	// WebSocket endpoint
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		defer c.Close()
		hub.Add(c)
		defer hub.Remove(c)

		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				break
			}
		}
	}))

	// Background task for auto-stopping expired rentals
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := db.AutoStopExpired(database, sender); err != nil {
				log.Printf("Auto-stop error: %v", err)
			}
		}
	}()

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Validate port
	if _, err := strconv.Atoi(port); err != nil {
		log.Fatal("Invalid PORT value:", port)
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server listening on %s", addr)
	log.Fatal(app.Listen(addr))
}