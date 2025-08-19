package usecases

import (
	"switchiot/internal/domain/entities"
	"switchiot/internal/domain/repositories"
	"switchiot/internal/domain/usecases"
)

// TransactionUseCase implements transaction business logic
type TransactionUseCase struct {
	transactionRepo repositories.TransactionRepository
}

// NewTransactionUseCase creates a new transaction use case
func NewTransactionUseCase(transactionRepo repositories.TransactionRepository) usecases.TransactionService {
	return &TransactionUseCase{
		transactionRepo: transactionRepo,
	}
}

// GetTransactionsByConsole returns transactions for a specific console
func (t *TransactionUseCase) GetTransactionsByConsole(consoleID int64) ([]entities.Transaction, error) {
	return t.transactionRepo.GetByConsoleID(consoleID, 50) // Limit to 50 recent transactions
}

// GetLastTransaction returns the most recent transaction for a console
func (t *TransactionUseCase) GetLastTransaction(consoleID int64) (*entities.Transaction, error) {
	return t.transactionRepo.GetLast(consoleID)
}