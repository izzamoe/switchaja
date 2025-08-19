package server

import (
	"io/fs"
	"log"
	"switchiot/internal/api"
	"switchiot/internal/app"
	"time"

	switchiot "switchiot"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	websocket "github.com/gofiber/websocket/v2"
)

// Server represents the HTTP server
type Server struct {
	app     *app.Application
	fiberApp *fiber.App
}

// NewServer creates a new HTTP server
func NewServer(application *app.Application) *Server {
	fiberApp := fiber.New()
	fiberApp.Use(logger.New())

	server := &Server{
		app:      application,
		fiberApp: fiberApp,
	}

	server.setupRoutes()
	server.setupStaticFiles()
	server.setupWebSocket()

	return server
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() {
	// For now, we'll use the existing API layer to maintain compatibility
	// In a future iteration, we can fully replace it with our new controllers
	apiLayer := api.New(s.app.Database, s.app.IoTSender, s.app.Hub)
	apiLayer.Register(s.fiberApp)
}

// setupStaticFiles configures static file serving
func (s *Server) setupStaticFiles() {
	// static admin page (embedded)
	sub, err := fs.Sub(switchiot.EmbeddedStatic, "web/static")
	if err != nil {
		log.Fatalf("embed sub: %v", err)
	}

	s.fiberApp.Get("/", func(c *fiber.Ctx) error {
		// if not logged in or stale session redirect to /login
		if !api.IsValidToken(c.Cookies("hehetoken")) {
			return c.Redirect("/login")
		}
		b, err := fs.ReadFile(sub, "index.html")
		if err != nil {
			return err
		}
		c.Type("html")
		return c.Send(b)
	})

	s.fiberApp.Get("/login", func(c *fiber.Ctx) error {
		b, err := fs.ReadFile(sub, "login.html")
		if err != nil {
			return err
		}
		c.Type("html")
		return c.Send(b)
	})

	s.fiberApp.Get("/style.css", func(c *fiber.Ctx) error {
		b, err := fs.ReadFile(sub, "style.css")
		if err != nil {
			return err
		}
		c.Type("css")
		return c.Send(b)
	})
}

// setupWebSocket configures the WebSocket endpoint
func (s *Server) setupWebSocket() {
	apiLayer := api.New(s.app.Database, s.app.IoTSender, s.app.Hub)

	s.fiberApp.Get("/ws", websocket.New(func(c *websocket.Conn) {
		defer c.Close()
		s.app.Hub.Add(c)
		// initial snapshot
		if err := c.WriteMessage(websocket.TextMessage, apiLayer.StatusPayload()); err != nil {
			s.app.Hub.Remove(c)
			return
		}
		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				s.app.Hub.Remove(c)
				break
			}
			if mt == websocket.TextMessage {
				if string(msg) == "{\"type\":\"ping\"}" || string(msg) == "ping" {
					_ = c.WriteMessage(websocket.TextMessage, apiLayer.StatusPayload())
				}
			}
		}
	}))
}

// Start starts the HTTP server
func (s *Server) Start() error {
	port := s.app.Config.Server.Port
	log.Printf("listening on :%s\n", port)
	return s.fiberApp.Listen(":" + port)
}

// StartBackgroundTasks starts background tasks like auto-stop watcher
func (s *Server) StartBackgroundTasks() {
	go s.runBackgroundLoop()
}

// runBackgroundLoop runs the adaptive background loop
func (s *Server) runBackgroundLoop() {
	fast := 2 * time.Second
	slow := 10 * time.Second
	lastTick := time.Now()
	
	apiLayer := api.New(s.app.Database, s.app.IoTSender, s.app.Hub)

	for {
		interval := slow
		if s.app.Hub.Size() > 0 {
			interval = fast
		}
		time.Sleep(time.Until(lastTick.Add(interval)))
		lastTick = time.Now()

		// Check for expired rentals
		expiredConsoles, err := s.app.ConsoleService.CheckExpiredRentals()
		if err != nil {
			log.Printf("Error checking expired rentals: %v", err)
			continue
		}

		// Stop expired consoles and send IoT commands
		for _, console := range expiredConsoles {
			log.Printf("auto-stop %s (expired)\n", console.Name)
			_ = s.app.IoTSender.Send(console.ID, "OFF")
		}

		// Check for consoles due soon
		dueConsoles, err := s.app.ConsoleService.GetDueSoon(time.Minute)
		if err != nil {
			log.Printf("Error checking due soon consoles: %v", err)
			continue
		}

		// Log warnings for consoles due soon
		for _, console := range dueConsoles {
			if s.app.Hub.Size() > 0 {
				remaining := console.TimeRemaining()
				log.Printf("warning: %s akan habis dalam %d detik", console.Name, int(remaining.Seconds()))
			}
		}

		// Broadcast status updates if there are clients connected
		if s.app.Hub.Size() > 0 {
			apiLayer.BroadcastStatus()
		}
	}
}