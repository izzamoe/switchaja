package main

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	_ "modernc.org/sqlite"

	switchiot "switchiot"
	"switchiot/internal/api"
	"switchiot/internal/db"
	"switchiot/internal/iot"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	websocket "github.com/gofiber/websocket/v2"
)

// main boots the Fiber web server serving the REST API and static admin UI.
// main starts the HTTP server and background workers:
//   - Auto-stop watcher for expired sessions
//   - Due-soon notifier (logs when < 1 minute remaining)
func main() {
	// Determine listen port from env (PORT). Defaults to 8080.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// normalize (strip leading colon if user provides :8080)
	port = strings.TrimPrefix(port, ":")
	// open sqlite (file based)
	database, err := sql.Open("sqlite", "file:heheswitch.db?_pragma=busy_timeout(5000)")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// Apply performance pragmas depending on SQLITE_MODE env: aggressive|balanced|safe (default balanced)
	mode := strings.ToLower(os.Getenv("SQLITE_MODE"))
	if mode == "" {
		mode = "balanced"
	}
	if err := tuneSQLite(database, mode); err != nil {
		log.Printf("sqlite tuning (%s) error: %v", mode, err)
	} else {
		log.Printf("sqlite tuning mode=%s applied", mode)
	}

	if err := db.Init(database, 5, 40000); err != nil {
		log.Fatal(err)
	}
	// seed default admin if no users
	if cnt, _ := db.CountUsers(database); cnt == 0 {
		// password: admin123 (bcrypt)
		pwHash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		_, _ = db.CreateUser(database, "admin", string(pwHash), "admin")
		log.Println("seeded default admin user 'admin' password 'admin123'")
	}

	app := fiber.New()
	app.Use(logger.New())

	// Choose IoT sender: MQTT if MQTT_BROKER set, else mock
	var sender iot.CommandSender
	var initialMqtt *iot.MQTTSender
	// Priority: DB stored config > env (legacy) > mock
	if cfg, ok, _ := db.LoadMQTTConfig(database); ok && cfg.Broker != "" {
		for attempt := 1; attempt <= 3; attempt++ {
			ms, err := iot.NewMQTTSender(cfg.Broker, iot.MQTTSenderOptions{Prefix: cfg.Prefix, Username: cfg.Username, Password: cfg.Password, QOS: 1, CleanSession: true, StatusCallback: func(id int64, payload string) {}})
			if err == nil {
				initialMqtt = ms
				sender = ms
				break
			}
			log.Printf("MQTT stored connect failed (%d/3): %v", attempt, err)
			time.Sleep(1 * time.Second)
		}
	}
	if sender == nil { // fallback env if still nil
		if os.Getenv("MQTT_BROKER") != "" {
			ms, err := iot.NewFromEnv(func(id int64, payload string) { log.Printf("status update from device %d: %s", id, payload) })
			if err == nil {
				initialMqtt = ms
				sender = ms
			} else {
				log.Printf("MQTT env connect failed: %v", err)
				sender = iot.NewMockSender()
			}
		} else {
			sender = iot.NewMockSender()
		}
	}
	// wrap with idempotent filter to avoid duplicate ON/OFF publishes
	sender = iot.NewIdempotentSender(sender)
	hub := iot.NewHub()
	apiLayer := api.New(database, sender, hub)
	apiLayer.Mqtt = initialMqtt
	apiLayer.Register(app)

	// websocket endpoint: send immediate snapshot, reply to ping
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		defer c.Close()
		hub.Add(c)
		// initial snapshot
		if err := c.WriteMessage(websocket.TextMessage, apiLayer.StatusPayload()); err != nil {
			hub.Remove(c)
			return
		}
		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				hub.Remove(c)
				break
			}
			if mt == websocket.TextMessage {
				if string(msg) == "{\"type\":\"ping\"}" || string(msg) == "ping" {
					_ = c.WriteMessage(websocket.TextMessage, apiLayer.StatusPayload())
				}
			}
		}
	}))

	// static admin page (embedded)
	sub, err := fs.Sub(switchiot.EmbeddedStatic, "web/static")
	if err != nil {
		log.Fatalf("embed sub: %v", err)
	}
	app.Get("/", func(c *fiber.Ctx) error {
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
	app.Get("/login", func(c *fiber.Ctx) error {
		b, err := fs.ReadFile(sub, "login.html")
		if err != nil {
			return err
		}
		c.Type("html")
		return c.Send(b)
	})
	app.Get("/style.css", func(c *fiber.Ctx) error {
		b, err := fs.ReadFile(sub, "style.css")
		if err != nil {
			return err
		}
		c.Type("css")
		return c.Send(b)
	})

	// Adaptive background loop (faster when clients connected, slower when idle)
	go func() {
		fast := 2 * time.Second
		slow := 10 * time.Second
		lastTick := time.Now()
		for {
			interval := slow
			if hub.Size() > 0 {
				interval = fast
			}
			time.Sleep(time.Until(lastTick.Add(interval)))
			lastTick = time.Now()
			// always handle expirations even if no clients
			consoles, err := db.GetConsoles(database)
			if err != nil {
				continue
			}
			now := time.Now()
			changed := false
			for _, c := range consoles {
				if c.Status == "RUNNING" && !c.EndTime.IsZero() {
					if now.After(c.EndTime) {
						log.Printf("auto-stop %s (expired)\n", c.Name)
						_ = db.StopRental(database, c.ID)
						_ = sender.Send(c.ID, "OFF")
						changed = true
					} else if c.EndTime.Sub(now) < time.Minute && hub.Size() > 0 {
						log.Printf("warning: %s akan habis dalam %d detik", c.Name, int(c.EndTime.Sub(now).Seconds()))
					}
				}
			}
			if hub.Size() == 0 { // skip broadcasting snapshot if no clients
				continue
			}
			apiLayer.BroadcastStatus()
			if changed {
				apiLayer.BroadcastStatus()
			}
		}
	}()

	log.Printf("listening on :%s\n", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}

// tuneSQLite applies PRAGMA settings for performance/safety trade-offs.
// Modes:
//
//	aggressive: maximum speed, risk of data loss on crash (synchronous=OFF, exclusive locking)
//	balanced: good speed with reasonable safety (synchronous=NORMAL)
//	safe: highest durability (synchronous=FULL)
func tuneSQLite(db *sql.DB, mode string) error {
	type pragmaSet struct {
		stmts            []string
		maxOpen, maxIdle int
	}
	makeBase := func(sync string, cacheMB int, extra []string) pragmaSet {
		// cache_size expects negative value = size in KB * -1
		cacheKB := cacheMB * 1024
		stmts := []string{
			"PRAGMA journal_mode=WAL",
			fmt.Sprintf("PRAGMA synchronous=%s", sync),
			"PRAGMA temp_store=MEMORY",
			fmt.Sprintf("PRAGMA cache_size=%d", -cacheKB),
			"PRAGMA wal_autocheckpoint=1000",
			"PRAGMA busy_timeout=5000",
		}
		stmts = append(stmts, extra...)
		return pragmaSet{stmts: stmts, maxOpen: 4, maxIdle: 4}
	}
	var cfg pragmaSet
	switch mode {
	case "aggressive":
		cfg = makeBase("OFF", 256, []string{"PRAGMA locking_mode=EXCLUSIVE", "PRAGMA mmap_size=30000000000", "PRAGMA journal_size_limit=67108864"})
		cfg.maxOpen, cfg.maxIdle = 1, 1 // exclusive mode works best single connection
	case "safe":
		cfg = makeBase("FULL", 16, []string{"PRAGMA mmap_size=100000000"})
	default: // balanced
		cfg = makeBase("NORMAL", 64, []string{"PRAGMA mmap_size=100000000", "PRAGMA journal_size_limit=67108864"})
	}
	for _, s := range cfg.stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("exec %s: %w", s, err)
		}
	}
	db.SetMaxOpenConns(cfg.maxOpen)
	db.SetMaxIdleConns(cfg.maxIdle)
	db.SetConnMaxLifetime(0)
	return nil
}
