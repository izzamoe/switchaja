package usecases

import (
	"fmt"
	"switchiot/internal/domain/entities"
	"switchiot/internal/domain/repositories"
	"switchiot/internal/domain/usecases"

	"golang.org/x/crypto/bcrypt"
)

// UserUseCase implements user business logic
type UserUseCase struct {
	userRepo repositories.UserRepository
}

// NewUserUseCase creates a new user use case
func NewUserUseCase(userRepo repositories.UserRepository) usecases.UserService {
	return &UserUseCase{
		userRepo: userRepo,
	}
}

// CreateUser creates a new user with hashed password
func (u *UserUseCase) CreateUser(username, password, role string) (int64, error) {
	if username == "" {
		return 0, fmt.Errorf("username cannot be empty")
	}
	if password == "" {
		return 0, fmt.Errorf("password cannot be empty")
	}
	if role != entities.RoleAdmin && role != entities.RoleUser {
		return 0, fmt.Errorf("invalid role: %s", role)
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	return u.userRepo.Create(username, string(hashedPassword), role)
}

// AuthenticateUser authenticates a user and returns user info
func (u *UserUseCase) AuthenticateUser(username, password string) (*entities.User, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password are required")
	}

	user, hashedPassword, err := u.userRepo.GetByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	return user, nil
}

// GetAllUsers returns all users
func (u *UserUseCase) GetAllUsers() ([]entities.User, error) {
	return u.userRepo.GetAll()
}

// DeleteUser deletes a user by ID
func (u *UserUseCase) DeleteUser(id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid user ID")
	}
	return u.userRepo.Delete(id)
}

// GetUserCount returns the total number of users
func (u *UserUseCase) GetUserCount() (int, error) {
	return u.userRepo.Count()
}