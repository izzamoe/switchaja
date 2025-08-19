package main

import (
	"log"
	"os"
	"switchiot/internal/config"
	"switchiot/internal/domain/entities"
	"time"
)

// Simple test to verify our domain entities work correctly
func main() {
	log.Println("Testing domain entities...")

	// Test Console entity
	console := entities.Console{
		ID:           1,
		Name:         "PS1",
		Status:       entities.StatusIdle,
		PricePerHour: 40000,
	}

	log.Printf("Initial console: %+v", console)

	// Test starting a rental
	console.StartRental(30) // 30 minutes
	log.Printf("After starting rental: %+v", console)
	log.Printf("Is running: %v", console.IsRunning())
	log.Printf("Time remaining: %v", console.TimeRemaining())

	// Test extending rental
	console.ExtendRental(15) // 15 more minutes
	log.Printf("After extending rental: %v", console.TimeRemaining())

	// Test Transaction entity
	transaction := entities.NewTransaction(console.ID, 30, console.PricePerHour)
	log.Printf("New transaction: %+v", transaction)

	// Test extending transaction
	transaction.ExtendDuration(15)
	log.Printf("After extending transaction: %+v", transaction)

	// Test User entity
	user := entities.User{
		ID:        1,
		Username:  "admin",
		Role:      entities.RoleAdmin,
		CreatedAt: time.Now(),
	}

	log.Printf("User: %+v", user)
	log.Printf("Is admin: %v", user.IsAdmin())
	log.Printf("Can manage users: %v", user.CanManageUsers())

	// Test Configuration
	config := config.LoadConfig()
	log.Printf("Config: %+v", config)

	log.Println("Domain entities test completed successfully!")
	os.Exit(0)
}