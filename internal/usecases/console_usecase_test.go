package usecases

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"switchiot/internal/domain/entities"
	domainErrors "switchiot/internal/domain/errors"
	"switchiot/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConsoleUseCase_NewConsoleUseCase(t *testing.T) {
	consoleRepo := &mocks.MockConsoleRepository{}
	transactionRepo := &mocks.MockTransactionRepository{}

	useCase := NewConsoleUseCase(consoleRepo, transactionRepo)
	assert.NotNil(t, useCase)
}

func TestConsoleUseCase_GetAllConsoles_Success(t *testing.T) {
	consoleRepo := &mocks.MockConsoleRepository{}
	transactionRepo := &mocks.MockTransactionRepository{}
	useCase := NewConsoleUseCase(consoleRepo, transactionRepo)

	expectedConsoles := []entities.Console{
		{ID: 1, Name: "PS1", Status: entities.StatusIdle},
		{ID: 2, Name: "PS2", Status: entities.StatusRunning},
	}

	consoleRepo.On("GetAll").Return(expectedConsoles, nil)

	consoles, err := useCase.GetAllConsoles()

	assert.NoError(t, err)
	assert.Equal(t, expectedConsoles, consoles)
	consoleRepo.AssertExpectations(t)
}

func TestConsoleUseCase_GetAllConsoles_Error(t *testing.T) {
	consoleRepo := &mocks.MockConsoleRepository{}
	transactionRepo := &mocks.MockTransactionRepository{}
	useCase := NewConsoleUseCase(consoleRepo, transactionRepo)

	repoError := errors.New("database error")
	consoleRepo.On("GetAll").Return([]entities.Console(nil), repoError)

	consoles, err := useCase.GetAllConsoles()

	assert.Error(t, err)
	assert.Nil(t, consoles)

	domainErr := err.(*domainErrors.DomainError)
	assert.Equal(t, domainErrors.CodeInternalError, domainErr.Code)
	assert.Equal(t, repoError, domainErr.Cause)

	consoleRepo.AssertExpectations(t)
}

func TestConsoleUseCase_StartRental_Success(t *testing.T) {
	consoleRepo := &mocks.MockConsoleRepository{}
	transactionRepo := &mocks.MockTransactionRepository{}
	useCase := NewConsoleUseCase(consoleRepo, transactionRepo)

	console := &entities.Console{
		ID:           1,
		Name:         "PS1",
		Status:       entities.StatusIdle,
		PricePerHour: 40000,
	}

	consoleRepo.On("GetByID", int64(1)).Return(console, nil)
	consoleRepo.On("Update", mock.AnythingOfType("*entities.Console")).Return(nil)
	transactionRepo.On("Create", mock.AnythingOfType("*entities.Transaction")).Return(nil)

	err := useCase.StartRental(1, 30)

	assert.NoError(t, err)
	consoleRepo.AssertExpectations(t)
	transactionRepo.AssertExpectations(t)
}

func TestConsoleUseCase_StartRental_InvalidDuration(t *testing.T) {
	consoleRepo := &mocks.MockConsoleRepository{}
	transactionRepo := &mocks.MockTransactionRepository{}
	useCase := NewConsoleUseCase(consoleRepo, transactionRepo)

	err := useCase.StartRental(1, 0)

	assert.Error(t, err)
	domainErr := err.(*domainErrors.DomainError)
	assert.Equal(t, domainErrors.CodeInvalidDuration, domainErr.Code)
}

func TestConsoleUseCase_StartRental_ConsoleNotFound(t *testing.T) {
	consoleRepo := &mocks.MockConsoleRepository{}
	transactionRepo := &mocks.MockTransactionRepository{}
	useCase := NewConsoleUseCase(consoleRepo, transactionRepo)

	consoleRepo.On("GetByID", int64(999)).Return(nil, sql.ErrNoRows)

	err := useCase.StartRental(999, 30)

	assert.Error(t, err)
	domainErr := err.(*domainErrors.DomainError)
	assert.Equal(t, domainErrors.CodeConsoleNotFound, domainErr.Code)
	consoleRepo.AssertExpectations(t)
}

func TestConsoleUseCase_StartRental_ConsoleAlreadyRunning(t *testing.T) {
	consoleRepo := &mocks.MockConsoleRepository{}
	transactionRepo := &mocks.MockTransactionRepository{}
	useCase := NewConsoleUseCase(consoleRepo, transactionRepo)

	console := &entities.Console{
		ID:     1,
		Name:   "PS1",
		Status: entities.StatusRunning,
	}

	consoleRepo.On("GetByID", int64(1)).Return(console, nil)

	err := useCase.StartRental(1, 30)

	assert.Error(t, err)
	domainErr := err.(*domainErrors.DomainError)
	assert.Equal(t, domainErrors.CodeConsoleAlreadyRunning, domainErr.Code)
	consoleRepo.AssertExpectations(t)
}

func TestConsoleUseCase_ExtendRental_Success(t *testing.T) {
	consoleRepo := &mocks.MockConsoleRepository{}
	transactionRepo := &mocks.MockTransactionRepository{}
	useCase := NewConsoleUseCase(consoleRepo, transactionRepo)

	console := &entities.Console{
		ID:      1,
		Name:    "PS1",
		Status:  entities.StatusRunning,
		EndTime: time.Now().Add(time.Hour),
	}
	transaction := &entities.Transaction{
		ID:          1,
		ConsoleID:   1,
		DurationMin: 60,
	}

	consoleRepo.On("GetByID", int64(1)).Return(console, nil)
	consoleRepo.On("Update", mock.AnythingOfType("*entities.Console")).Return(nil)
	transactionRepo.On("GetLast", int64(1)).Return(transaction, nil)
	transactionRepo.On("Update", mock.AnythingOfType("*entities.Transaction")).Return(nil)

	err := useCase.ExtendRental(1, 30)

	assert.NoError(t, err)
	consoleRepo.AssertExpectations(t)
	transactionRepo.AssertExpectations(t)
}

func TestConsoleUseCase_ExtendRental_InvalidDuration(t *testing.T) {
	consoleRepo := &mocks.MockConsoleRepository{}
	transactionRepo := &mocks.MockTransactionRepository{}
	useCase := NewConsoleUseCase(consoleRepo, transactionRepo)

	err := useCase.ExtendRental(1, 0)

	assert.Error(t, err)
	domainErr := err.(*domainErrors.DomainError)
	assert.Equal(t, domainErrors.CodeInvalidDuration, domainErr.Code)
}

func TestConsoleUseCase_ExtendRental_ConsoleNotRunning(t *testing.T) {
	consoleRepo := &mocks.MockConsoleRepository{}
	transactionRepo := &mocks.MockTransactionRepository{}
	useCase := NewConsoleUseCase(consoleRepo, transactionRepo)

	console := &entities.Console{
		ID:     1,
		Name:   "PS1",
		Status: entities.StatusIdle,
	}

	consoleRepo.On("GetByID", int64(1)).Return(console, nil)

	err := useCase.ExtendRental(1, 30)

	assert.Error(t, err)
	domainErr := err.(*domainErrors.DomainError)
	assert.Equal(t, domainErrors.CodeConsoleNotRunning, domainErr.Code)
	consoleRepo.AssertExpectations(t)
}

func TestConsoleUseCase_UpdatePrice_Success(t *testing.T) {
	consoleRepo := &mocks.MockConsoleRepository{}
	transactionRepo := &mocks.MockTransactionRepository{}
	useCase := NewConsoleUseCase(consoleRepo, transactionRepo)

	consoleRepo.On("UpdatePrice", int64(1), 50000).Return(nil)

	err := useCase.UpdatePrice(1, 50000)

	assert.NoError(t, err)
	consoleRepo.AssertExpectations(t)
}

func TestConsoleUseCase_UpdatePrice_InvalidPrice(t *testing.T) {
	consoleRepo := &mocks.MockConsoleRepository{}
	transactionRepo := &mocks.MockTransactionRepository{}
	useCase := NewConsoleUseCase(consoleRepo, transactionRepo)

	err := useCase.UpdatePrice(1, 0)

	assert.Error(t, err)
	domainErr := err.(*domainErrors.DomainError)
	assert.Equal(t, domainErrors.CodeInvalidPrice, domainErr.Code)
}

func TestConsoleUseCase_GetDueSoon_Success(t *testing.T) {
	consoleRepo := &mocks.MockConsoleRepository{}
	transactionRepo := &mocks.MockTransactionRepository{}
	useCase := NewConsoleUseCase(consoleRepo, transactionRepo)

	threshold := 5 * time.Minute
	expectedConsoles := []entities.Console{
		{ID: 1, Name: "PS1", Status: entities.StatusRunning},
	}

	consoleRepo.On("GetDueSoon", threshold).Return(expectedConsoles, nil)

	consoles, err := useCase.GetDueSoon(threshold)

	assert.NoError(t, err)
	assert.Equal(t, expectedConsoles, consoles)
	consoleRepo.AssertExpectations(t)
}

func TestConsoleUseCase_CheckExpiredRentals_Success(t *testing.T) {
	consoleRepo := &mocks.MockConsoleRepository{}
	transactionRepo := &mocks.MockTransactionRepository{}
	useCase := NewConsoleUseCase(consoleRepo, transactionRepo)

	now := time.Now()
	consoles := []entities.Console{
		{
			ID:      1,
			Name:    "PS1",
			Status:  entities.StatusRunning,
			EndTime: now.Add(-time.Hour), // Expired
		},
		{
			ID:      2,
			Name:    "PS2",
			Status:  entities.StatusRunning,
			EndTime: now.Add(time.Hour), // Not expired
		},
	}

	consoleRepo.On("GetAll").Return(consoles, nil)
	consoleRepo.On("Update", mock.MatchedBy(func(c *entities.Console) bool {
		return c.ID == 1 && c.Status == entities.StatusIdle
	})).Return(nil)

	expiredConsoles, err := useCase.CheckExpiredRentals()

	assert.NoError(t, err)
	assert.Len(t, expiredConsoles, 1)
	assert.Equal(t, int64(1), expiredConsoles[0].ID)
	assert.Equal(t, entities.StatusIdle, expiredConsoles[0].Status)
	consoleRepo.AssertExpectations(t)
}
