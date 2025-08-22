package usecases

import (
	"switchiot/internal/domain/entities"
)

// UserService defines the interface for user management business logic
type UserService interface {
	// CreateUser creates a new user with hashed password
	CreateUser(username, password, role string) (int64, error)

	// AuthenticateUser authenticates a user and returns user info
	AuthenticateUser(username, password string) (*entities.User, error)

	// GetAllUsers returns all users
	GetAllUsers() ([]entities.User, error)

	// DeleteUser deletes a user by ID
	DeleteUser(id int64) error

	// GetUserCount returns the total number of users
	GetUserCount() (int, error)
}
