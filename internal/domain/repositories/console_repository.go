package repositories

import (
	"switchiot/internal/domain/entities"
	"time"
)

// ConsoleRepository defines the interface for console data access
type ConsoleRepository interface {
	// GetAll returns all consoles
	GetAll() ([]entities.Console, error)

	// GetByID returns a console by its ID
	GetByID(id int64) (*entities.Console, error)

	// Update updates a console
	Update(console *entities.Console) error

	// UpdatePrice updates the price of a console
	UpdatePrice(consoleID int64, newPrice int) error

	// GetDueSoon returns consoles whose rentals will expire within the threshold
	GetDueSoon(threshold time.Duration) ([]entities.Console, error)
}
