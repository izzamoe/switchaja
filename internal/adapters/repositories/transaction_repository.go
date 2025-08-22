package repositories

import (
	"database/sql"
	"switchiot/internal/db"
	"switchiot/internal/domain/entities"
	"switchiot/internal/domain/repositories"
)

// SQLTransactionRepository implements TransactionRepository using SQL database
type SQLTransactionRepository struct {
	db *sql.DB
}

// NewSQLTransactionRepository creates a new SQL transaction repository
func NewSQLTransactionRepository(database *sql.DB) repositories.TransactionRepository {
	return &SQLTransactionRepository{db: database}
}

// Create creates a new transaction
func (r *SQLTransactionRepository) Create(transaction *entities.Transaction) error {
	query := `INSERT INTO transactions (console_id, start_time, end_time, duration_minutes, total_price, price_per_hour_snapshot) 
			  VALUES (?, ?, ?, ?, ?, ?)`
	result, err := r.db.Exec(query,
		transaction.ConsoleID,
		transaction.StartTime,
		transaction.EndTime,
		transaction.DurationMin,
		transaction.TotalPrice,
		transaction.PricePerHourSnapshot)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	transaction.ID = id
	return nil
}

// GetByConsoleID returns transactions for a specific console
func (r *SQLTransactionRepository) GetByConsoleID(consoleID int64, limit int) ([]entities.Transaction, error) {
	rows, err := r.db.Query(`SELECT id, console_id, start_time, end_time, duration_minutes, total_price, COALESCE(price_per_hour_snapshot,0) 
							 FROM transactions WHERE console_id=? ORDER BY id DESC LIMIT ?`, consoleID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []entities.Transaction
	for rows.Next() {
		var t entities.Transaction
		if err := rows.Scan(&t.ID, &t.ConsoleID, &t.StartTime, &t.EndTime, &t.DurationMin, &t.TotalPrice, &t.PricePerHourSnapshot); err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}

	return transactions, nil
}

// GetLast returns the most recent transaction for a console
func (r *SQLTransactionRepository) GetLast(consoleID int64) (*entities.Transaction, error) {
	dbTransaction, found, err := db.LastTransaction(r.db, consoleID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}

	transaction := &entities.Transaction{
		ID:                   dbTransaction.ID,
		ConsoleID:            dbTransaction.ConsoleID,
		StartTime:            dbTransaction.StartTime,
		EndTime:              dbTransaction.EndTime,
		DurationMin:          dbTransaction.DurationMin,
		TotalPrice:           dbTransaction.TotalPrice,
		PricePerHourSnapshot: dbTransaction.PricePerHourSnapshot,
	}

	return transaction, nil
}

// Update updates a transaction
func (r *SQLTransactionRepository) Update(transaction *entities.Transaction) error {
	query := `UPDATE transactions SET end_time = ?, duration_minutes = ?, total_price = ? WHERE id = ?`
	_, err := r.db.Exec(query, transaction.EndTime, transaction.DurationMin, transaction.TotalPrice, transaction.ID)
	return err
}
