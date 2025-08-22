package usecases

import (
	"errors"
	"testing"
	"time"

	"switchiot/internal/domain/entities"
	domainErrors "switchiot/internal/domain/errors"
	"switchiot/internal/mocks"

	"github.com/stretchr/testify/assert"
)

func TestTransactionUseCase_NewTransactionUseCase(t *testing.T) {
	transactionRepo := &mocks.MockTransactionRepository{}

	useCase := NewTransactionUseCase(transactionRepo)
	assert.NotNil(t, useCase)
}

func TestTransactionUseCase_GetTransactionsByConsole(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		transactionRepo := &mocks.MockTransactionRepository{}
		useCase := NewTransactionUseCase(transactionRepo)

		consoleID := int64(1)
		expectedTransactions := []entities.Transaction{
			{
				ID:          1,
				ConsoleID:   consoleID,
				StartTime:   time.Now().Add(-time.Hour),
				EndTime:     time.Now().Add(-30 * time.Minute),
				DurationMin: 30,
				TotalPrice:  20000,
			},
		}

		transactionRepo.On("GetByConsoleID", consoleID, 50).Return(expectedTransactions, nil)

		transactions, err := useCase.GetTransactionsByConsole(consoleID)

		assert.NoError(t, err)
		assert.Equal(t, expectedTransactions, transactions)
		transactionRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		transactionRepo := &mocks.MockTransactionRepository{}
		useCase := NewTransactionUseCase(transactionRepo)

		consoleID := int64(1)
		repoError := errors.New("database error")

		transactionRepo.On("GetByConsoleID", consoleID, 50).Return([]entities.Transaction(nil), repoError)

		transactions, err := useCase.GetTransactionsByConsole(consoleID)

		assert.Error(t, err)
		assert.Nil(t, transactions)
		domainErr := err.(*domainErrors.DomainError)
		assert.Equal(t, domainErrors.CodeInternalError, domainErr.Code)
		assert.Equal(t, repoError, domainErr.Cause)
		transactionRepo.AssertExpectations(t)
	})
}

func TestTransactionUseCase_GetLastTransaction(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		transactionRepo := &mocks.MockTransactionRepository{}
		useCase := NewTransactionUseCase(transactionRepo)

		consoleID := int64(1)
		expectedTransaction := &entities.Transaction{
			ID:          1,
			ConsoleID:   consoleID,
			StartTime:   time.Now().Add(-time.Hour),
			EndTime:     time.Now().Add(-30 * time.Minute),
			DurationMin: 30,
			TotalPrice:  20000,
		}

		transactionRepo.On("GetLast", consoleID).Return(expectedTransaction, nil)

		transaction, err := useCase.GetLastTransaction(consoleID)

		assert.NoError(t, err)
		assert.Equal(t, expectedTransaction, transaction)
		transactionRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		transactionRepo := &mocks.MockTransactionRepository{}
		useCase := NewTransactionUseCase(transactionRepo)

		consoleID := int64(1)
		repoError := errors.New("database error")

		transactionRepo.On("GetLast", consoleID).Return(nil, repoError)

		transaction, err := useCase.GetLastTransaction(consoleID)

		assert.Error(t, err)
		assert.Nil(t, transaction)
		domainErr := err.(*domainErrors.DomainError)
		assert.Equal(t, domainErrors.CodeInternalError, domainErr.Code)
		assert.Equal(t, repoError, domainErr.Cause)
		transactionRepo.AssertExpectations(t)
	})
}
