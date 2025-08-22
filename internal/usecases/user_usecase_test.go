package usecases

import (
	"database/sql"
	"errors"
	"testing"

	"switchiot/internal/domain/entities"
	domainErrors "switchiot/internal/domain/errors"
	"switchiot/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserUseCase_NewUserUseCase(t *testing.T) {
	userRepo := &mocks.MockUserRepository{}

	useCase := NewUserUseCase(userRepo)
	assert.NotNil(t, useCase)
}

func TestUserUseCase_CreateUser(t *testing.T) {
	t.Run("success - admin user", func(t *testing.T) {
		userRepo := &mocks.MockUserRepository{}
		useCase := NewUserUseCase(userRepo)

		username := "admin"
		password := "password123"
		role := entities.RoleAdmin

		// Mock that user doesn't exist
		userRepo.On("GetByUsername", username).Return(nil, "", sql.ErrNoRows)
		userRepo.On("Create", username, mock.AnythingOfType("string"), role).Return(int64(1), nil)

		userID, err := useCase.CreateUser(username, password, role)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), userID)
		userRepo.AssertExpectations(t)
	})

	t.Run("empty username", func(t *testing.T) {
		userRepo := &mocks.MockUserRepository{}
		useCase := NewUserUseCase(userRepo)

		_, err := useCase.CreateUser("", "password", entities.RoleUser)

		assert.Error(t, err)
		domainErr := err.(*domainErrors.DomainError)
		assert.Equal(t, domainErrors.CodeInvalidUserData, domainErr.Code)
		assert.Contains(t, domainErr.Message, "username")
	})

	t.Run("invalid role", func(t *testing.T) {
		userRepo := &mocks.MockUserRepository{}
		useCase := NewUserUseCase(userRepo)

		_, err := useCase.CreateUser("user1", "password", "invalid_role")

		assert.Error(t, err)
		domainErr := err.(*domainErrors.DomainError)
		assert.Equal(t, domainErrors.CodeInvalidUserData, domainErr.Code)
		assert.Contains(t, domainErr.Message, "invalid role")
	})

	t.Run("user already exists", func(t *testing.T) {
		userRepo := &mocks.MockUserRepository{}
		useCase := NewUserUseCase(userRepo)

		username := "existing_user"
		existingUser := &entities.User{ID: 1, Username: username}

		userRepo.On("GetByUsername", username).Return(existingUser, "hashedpass", nil)

		_, err := useCase.CreateUser(username, "password", entities.RoleUser)

		assert.Error(t, err)
		domainErr := err.(*domainErrors.DomainError)
		assert.Equal(t, domainErrors.CodeUserAlreadyExists, domainErr.Code)
		userRepo.AssertExpectations(t)
	})
}

func TestUserUseCase_AuthenticateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userRepo := &mocks.MockUserRepository{}
		useCase := NewUserUseCase(userRepo)

		username := "testuser"
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		user := &entities.User{
			ID:       1,
			Username: username,
			Role:     entities.RoleUser,
		}

		userRepo.On("GetByUsername", username).Return(user, string(hashedPassword), nil)

		result, err := useCase.AuthenticateUser(username, password)

		assert.NoError(t, err)
		assert.Equal(t, user, result)
		userRepo.AssertExpectations(t)
	})

	t.Run("empty username", func(t *testing.T) {
		userRepo := &mocks.MockUserRepository{}
		useCase := NewUserUseCase(userRepo)

		_, err := useCase.AuthenticateUser("", "password")

		assert.Error(t, err)
		domainErr := err.(*domainErrors.DomainError)
		assert.Equal(t, domainErrors.CodeInvalidCredentials, domainErr.Code)
	})

	t.Run("user not found", func(t *testing.T) {
		userRepo := &mocks.MockUserRepository{}
		useCase := NewUserUseCase(userRepo)

		username := "nonexistent"

		userRepo.On("GetByUsername", username).Return(nil, "", sql.ErrNoRows)

		_, err := useCase.AuthenticateUser(username, "password")

		assert.Error(t, err)
		domainErr := err.(*domainErrors.DomainError)
		assert.Equal(t, domainErrors.CodeInvalidCredentials, domainErr.Code)
		userRepo.AssertExpectations(t)
	})

	t.Run("wrong password", func(t *testing.T) {
		userRepo := &mocks.MockUserRepository{}
		useCase := NewUserUseCase(userRepo)

		username := "testuser"
		correctPassword := "correct_password"
		wrongPassword := "wrong_password"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
		user := &entities.User{
			ID:       1,
			Username: username,
			Role:     entities.RoleUser,
		}

		userRepo.On("GetByUsername", username).Return(user, string(hashedPassword), nil)

		_, err := useCase.AuthenticateUser(username, wrongPassword)

		assert.Error(t, err)
		domainErr := err.(*domainErrors.DomainError)
		assert.Equal(t, domainErrors.CodeInvalidCredentials, domainErr.Code)
		userRepo.AssertExpectations(t)
	})
}

func TestUserUseCase_GetAllUsers(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userRepo := &mocks.MockUserRepository{}
		useCase := NewUserUseCase(userRepo)

		expectedUsers := []entities.User{
			{ID: 1, Username: "admin", Role: entities.RoleAdmin},
			{ID: 2, Username: "user1", Role: entities.RoleUser},
		}

		userRepo.On("GetAll").Return(expectedUsers, nil)

		users, err := useCase.GetAllUsers()

		assert.NoError(t, err)
		assert.Equal(t, expectedUsers, users)
		userRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		userRepo := &mocks.MockUserRepository{}
		useCase := NewUserUseCase(userRepo)

		repoError := errors.New("database error")
		userRepo.On("GetAll").Return([]entities.User(nil), repoError)

		users, err := useCase.GetAllUsers()

		assert.Error(t, err)
		assert.Nil(t, users)
		domainErr := err.(*domainErrors.DomainError)
		assert.Equal(t, domainErrors.CodeInternalError, domainErr.Code)
		assert.Equal(t, repoError, domainErr.Cause)
		userRepo.AssertExpectations(t)
	})
}

func TestUserUseCase_DeleteUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userRepo := &mocks.MockUserRepository{}
		useCase := NewUserUseCase(userRepo)

		userID := int64(1)

		userRepo.On("Delete", userID).Return(nil)

		err := useCase.DeleteUser(userID)

		assert.NoError(t, err)
		userRepo.AssertExpectations(t)
	})

	t.Run("invalid user ID", func(t *testing.T) {
		userRepo := &mocks.MockUserRepository{}
		useCase := NewUserUseCase(userRepo)

		err := useCase.DeleteUser(0)

		assert.Error(t, err)
		domainErr := err.(*domainErrors.DomainError)
		assert.Equal(t, domainErrors.CodeInvalidUserData, domainErr.Code)
		assert.Contains(t, domainErr.Message, "invalid user ID")
	})
}

func TestUserUseCase_GetUserCount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userRepo := &mocks.MockUserRepository{}
		useCase := NewUserUseCase(userRepo)

		expectedCount := 5

		userRepo.On("Count").Return(expectedCount, nil)

		count, err := useCase.GetUserCount()

		assert.NoError(t, err)
		assert.Equal(t, expectedCount, count)
		userRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		userRepo := &mocks.MockUserRepository{}
		useCase := NewUserUseCase(userRepo)

		repoError := errors.New("database error")
		userRepo.On("Count").Return(0, repoError)

		count, err := useCase.GetUserCount()

		assert.Error(t, err)
		assert.Equal(t, 0, count)
		domainErr := err.(*domainErrors.DomainError)
		assert.Equal(t, domainErrors.CodeInternalError, domainErr.Code)
		assert.Equal(t, repoError, domainErr.Cause)
		userRepo.AssertExpectations(t)
	})
}
