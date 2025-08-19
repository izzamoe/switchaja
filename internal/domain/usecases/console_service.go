package usecases

import (
	"switchiot/internal/domain/entities"
	"time"
)

// ConsoleService defines the interface for console business logic
type ConsoleService interface {
	// GetAllConsoles returns all consoles with current status
	GetAllConsoles() ([]entities.Console, error)
	
	// StartRental starts a rental session for a console
	StartRental(consoleID int64, durationMinutes int) error
	
	// ExtendRental extends the current rental session
	ExtendRental(consoleID int64, additionalMinutes int) error
	
	// StopRental stops the current rental session
	StopRental(consoleID int64) error
	
	// UpdatePrice updates the hourly price for a console
	UpdatePrice(consoleID int64, newPrice int) error
	
	// GetDueSoon returns consoles that will expire soon
	GetDueSoon(threshold time.Duration) ([]entities.Console, error)
	
	// CheckExpiredRentals checks and stops expired rental sessions
	CheckExpiredRentals() ([]entities.Console, error)
}