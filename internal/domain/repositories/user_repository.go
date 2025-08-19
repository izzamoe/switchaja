package repositories

import (
	"switchiot/internal/domain/entities"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create creates a new user
	Create(username, passwordHash, role string) (int64, error)
	
	// GetByUsername returns a user by username along with password hash
	GetByUsername(username string) (*entities.User, string, error)
	
	// GetAll returns all users
	GetAll() ([]entities.User, error)
	
	// Delete deletes a user by ID
	Delete(id int64) error
	
	// Count returns the total number of users
	Count() (int, error)
}