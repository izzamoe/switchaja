package repositories

import (
	"database/sql"
	"switchiot/internal/db"
	"switchiot/internal/domain/entities"
	"switchiot/internal/domain/repositories"
	"time"
)

// SQLConsoleRepository implements ConsoleRepository using SQL database
type SQLConsoleRepository struct {
	db *sql.DB
}

// NewSQLConsoleRepository creates a new SQL console repository
func NewSQLConsoleRepository(database *sql.DB) repositories.ConsoleRepository {
	return &SQLConsoleRepository{db: database}
}

// GetAll returns all consoles
func (r *SQLConsoleRepository) GetAll() ([]entities.Console, error) {
	dbConsoles, err := db.GetConsoles(r.db)
	if err != nil {
		return nil, err
	}

	consoles := make([]entities.Console, len(dbConsoles))
	for i, dbConsole := range dbConsoles {
		consoles[i] = entities.Console{
			ID:           dbConsole.ID,
			Name:         dbConsole.Name,
			Status:       dbConsole.Status,
			EndTime:      dbConsole.EndTime,
			PricePerHour: dbConsole.PricePerHour,
		}
	}

	return consoles, nil
}

// GetByID returns a console by its ID
func (r *SQLConsoleRepository) GetByID(id int64) (*entities.Console, error) {
	consoles, err := r.GetAll()
	if err != nil {
		return nil, err
	}

	for _, console := range consoles {
		if console.ID == id {
			return &console, nil
		}
	}

	return nil, sql.ErrNoRows
}

// Update updates a console
func (r *SQLConsoleRepository) Update(console *entities.Console) error {
	// Use existing database functions to update the console
	if console.IsRunning() {
		// If starting a rental, we need to update the database appropriately
		// This would typically involve updating the consoles table directly
		query := `UPDATE consoles SET status = ?, end_time = ? WHERE id = ?`
		_, err := r.db.Exec(query, console.Status, console.EndTime, console.ID)
		return err
	} else {
		// If stopping a rental, use the existing StopRental function
		return db.StopRental(r.db, console.ID)
	}
}

// UpdatePrice updates the price of a console
func (r *SQLConsoleRepository) UpdatePrice(consoleID int64, newPrice int) error {
	return db.UpdatePrice(r.db, consoleID, newPrice)
}

// GetDueSoon returns consoles whose rentals will expire within the threshold
func (r *SQLConsoleRepository) GetDueSoon(threshold time.Duration) ([]entities.Console, error) {
	dbConsoles, err := db.DueSoon(r.db, threshold)
	if err != nil {
		return nil, err
	}

	consoles := make([]entities.Console, len(dbConsoles))
	for i, dbConsole := range dbConsoles {
		consoles[i] = entities.Console{
			ID:           dbConsole.ID,
			Name:         dbConsole.Name,
			Status:       dbConsole.Status,
			EndTime:      dbConsole.EndTime,
			PricePerHour: dbConsole.PricePerHour,
		}
	}

	return consoles, nil
}