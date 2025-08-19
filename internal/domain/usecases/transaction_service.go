package usecases

import (
	"switchiot/internal/domain/entities"
)

// TransactionService defines the interface for transaction business logic
type TransactionService interface {
	// GetTransactionsByConsole returns transactions for a specific console
	GetTransactionsByConsole(consoleID int64) ([]entities.Transaction, error)
	
	// GetLastTransaction returns the most recent transaction for a console
	GetLastTransaction(consoleID int64) (*entities.Transaction, error)
}