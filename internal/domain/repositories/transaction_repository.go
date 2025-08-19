package repositories

import (
	"switchiot/internal/domain/entities"
)

// TransactionRepository defines the interface for transaction data access
type TransactionRepository interface {
	// Create creates a new transaction
	Create(transaction *entities.Transaction) error
	
	// GetByConsoleID returns transactions for a specific console
	GetByConsoleID(consoleID int64, limit int) ([]entities.Transaction, error)
	
	// GetLast returns the most recent transaction for a console
	GetLast(consoleID int64) (*entities.Transaction, error)
	
	// Update updates a transaction
	Update(transaction *entities.Transaction) error
}