package usecases

import (
	"database/sql"
	"switchiot/internal/domain/entities"
	"switchiot/internal/domain/errors"
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
		return 0, errors.NewInvalidUserData("username cannot be empty")
	}
	if password == "" {
		return 0, errors.NewInvalidUserData("password cannot be empty")
	}
	if role != entities.RoleAdmin && role != entities.RoleUser {
		return 0, errors.NewInvalidUserData("invalid role: " + role)
	}

	// Check if user already exists
	if _, _, err := u.userRepo.GetByUsername(username); err == nil {
		return 0, errors.NewUserAlreadyExists(username)
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, errors.NewInternalError(err)
	}

	userID, err := u.userRepo.Create(username, string(hashedPassword), role)
	if err != nil {
		return 0, errors.NewInternalError(err)
	}

	return userID, nil
}

// AuthenticateUser authenticates a user and returns user info
func (u *UserUseCase) AuthenticateUser(username, password string) (*entities.User, error) {
	if username == "" || password == "" {
		return nil, errors.NewInvalidCredentials()
	}

	user, hashedPassword, err := u.userRepo.GetByUsername(username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewInvalidCredentials()
		}
		return nil, errors.NewInternalError(err)
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return nil, errors.NewInvalidCredentials()
	}

	return user, nil
}

// GetAllUsers returns all users
func (u *UserUseCase) GetAllUsers() ([]entities.User, error) {
	users, err := u.userRepo.GetAll()
	if err != nil {
		return nil, errors.NewInternalError(err)
	}
	return users, nil
}

// DeleteUser deletes a user by ID
func (u *UserUseCase) DeleteUser(id int64) error {
	if id <= 0 {
		return errors.NewInvalidUserData("invalid user ID")
	}

	err := u.userRepo.Delete(id)
	if err != nil {
		return errors.NewInternalError(err)
	}
	return nil
}

// GetUserCount returns the total number of users
func (u *UserUseCase) GetUserCount() (int, error) {
	count, err := u.userRepo.Count()
	if err != nil {
		return 0, errors.NewInternalError(err)
	}
	return count, nil
}
