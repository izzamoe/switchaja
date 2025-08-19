package app

import (
	"database/sql"
	"log"
	"switchiot/internal/adapters/controllers"
	"switchiot/internal/adapters/repositories"
	"switchiot/internal/config"
	"switchiot/internal/db"
	"switchiot/internal/domain/usecases"
	"switchiot/internal/iot"
	usecaseimpl "switchiot/internal/usecases"

	"golang.org/x/crypto/bcrypt"
)

// Application represents the main application
type Application struct {
	Config             *config.Config
	Database           *sql.DB
	ConsoleController  *controllers.ConsoleController
	UserController     *controllers.UserController
	ConsoleService     usecases.ConsoleService
	UserService        usecases.UserService
	TransactionService usecases.TransactionService
	IoTSender          iot.CommandSender
	Hub                *iot.Hub
}

// NewApplication creates a new application instance with all dependencies wired up
func NewApplication(cfg *config.Config) (*Application, error) {
	// Initialize database
	database, err := sql.Open("sqlite", cfg.GetDatabaseConnectionString())
	if err != nil {
		return nil, err
	}

	// Apply SQLite tuning
	if err := tuneSQLite(database, cfg.Database.SQLMode); err != nil {
		log.Printf("sqlite tuning (%s) error: %v", cfg.Database.SQLMode, err)
	} else {
		log.Printf("sqlite tuning mode=%s applied", cfg.Database.SQLMode)
	}

	// Initialize database schema
	if err := db.Init(database, cfg.App.ConsoleCount, cfg.App.DefaultPrice); err != nil {
		return nil, err
	}

	// Seed default admin if no users exist
	if err := seedDefaultAdmin(database, cfg.App.DefaultAdmin); err != nil {
		return nil, err
	}

	// Initialize repositories
	consoleRepo := repositories.NewSQLConsoleRepository(database)
	transactionRepo := repositories.NewSQLTransactionRepository(database)
	userRepo := repositories.NewSQLUserRepository(database)

	// Initialize use cases
	consoleService := usecaseimpl.NewConsoleUseCase(consoleRepo, transactionRepo)
	userService := usecaseimpl.NewUserUseCase(userRepo)
	transactionService := usecaseimpl.NewTransactionUseCase(transactionRepo)

	// Initialize controllers
	consoleController := controllers.NewConsoleController(consoleService, transactionService)
	userController := controllers.NewUserController(userService)

	// Initialize IoT components
	iotSender := initializeIoTSender(database, cfg.MQTT)
	hub := iot.NewHub()

	return &Application{
		Config:             cfg,
		Database:           database,
		ConsoleController:  consoleController,
		UserController:     userController,
		ConsoleService:     consoleService,
		UserService:        userService,
		TransactionService: transactionService,
		IoTSender:          iotSender,
		Hub:                hub,
	}, nil
}

// Close closes the application and releases resources
func (app *Application) Close() error {
	if app.Database != nil {
		return app.Database.Close()
	}
	return nil
}

// seedDefaultAdmin creates a default admin user if no users exist
func seedDefaultAdmin(database *sql.DB, adminConfig config.AdminConfig) error {
	if cnt, _ := db.CountUsers(database); cnt == 0 {
		pwHash, err := bcrypt.GenerateFromPassword([]byte(adminConfig.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		_, err = db.CreateUser(database, adminConfig.Username, string(pwHash), "admin")
		if err != nil {
			return err
		}
		log.Printf("seeded default admin user '%s' password '%s'", adminConfig.Username, adminConfig.Password)
	}
	return nil
}

// initializeIoTSender creates and configures the IoT sender
func initializeIoTSender(database *sql.DB, mqttConfig config.MQTTConfig) iot.CommandSender {
	var sender iot.CommandSender

	// Priority: DB stored config > env (legacy) > mock
	if cfg, ok, _ := db.LoadMQTTConfig(database); ok && cfg.Broker != "" {
		for attempt := 1; attempt <= 3; attempt++ {
			ms, err := iot.NewMQTTSender(cfg.Broker, iot.MQTTSenderOptions{
				Prefix:         cfg.Prefix,
				Username:       cfg.Username,
				Password:       cfg.Password,
				QOS:            1,
				CleanSession:   true,
				StatusCallback: func(id int64, payload string) {},
			})
			if err == nil {
				sender = ms
				break
			}
			log.Printf("MQTT stored connect failed (%d/3): %v", attempt, err)
		}
	}

	if sender == nil && mqttConfig.Broker != "" {
		ms, err := iot.NewFromEnv(func(id int64, payload string) {
			log.Printf("status update from device %d: %s", id, payload)
		})
		if err == nil {
			sender = ms
		} else {
			log.Printf("MQTT env connect failed: %v", err)
			sender = iot.NewMockSender()
		}
	}

	if sender == nil {
		sender = iot.NewMockSender()
	}

	// wrap with idempotent filter to avoid duplicate ON/OFF publishes
	return iot.NewIdempotentSender(sender)
}