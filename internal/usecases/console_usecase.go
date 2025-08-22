package usecases

import (
	"database/sql"
	"switchiot/internal/domain/entities"
	"switchiot/internal/domain/errors"
	"switchiot/internal/domain/repositories"
	"switchiot/internal/domain/usecases"
	"time"
)

// ConsoleUseCase implements console business logic
type ConsoleUseCase struct {
	consoleRepo     repositories.ConsoleRepository
	transactionRepo repositories.TransactionRepository
}

// NewConsoleUseCase creates a new console use case
func NewConsoleUseCase(consoleRepo repositories.ConsoleRepository, transactionRepo repositories.TransactionRepository) usecases.ConsoleService {
	return &ConsoleUseCase{
		consoleRepo:     consoleRepo,
		transactionRepo: transactionRepo,
	}
}

// GetAllConsoles returns all consoles with current status
func (c *ConsoleUseCase) GetAllConsoles() ([]entities.Console, error) {
	consoles, err := c.consoleRepo.GetAll()
	if err != nil {
		return nil, errors.NewInternalError(err)
	}
	return consoles, nil
}

// StartRental starts a rental session for a console
func (c *ConsoleUseCase) StartRental(consoleID int64, durationMinutes int) error {
	if durationMinutes <= 0 {
		return errors.NewInvalidDuration()
	}

	console, err := c.consoleRepo.GetByID(consoleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NewConsoleNotFound(consoleID)
		}
		return errors.NewInternalError(err)
	}

	if console.IsRunning() {
		return errors.NewConsoleAlreadyRunning(console.Name)
	}

	// Start the rental in the entity
	console.StartRental(durationMinutes)

	// Update the console
	if err := c.consoleRepo.Update(console); err != nil {
		return errors.NewInternalError(err)
	}

	// Create transaction record
	transaction := entities.NewTransaction(consoleID, durationMinutes, console.PricePerHour)
	if err := c.transactionRepo.Create(transaction); err != nil {
		return errors.NewInternalError(err)
	}

	return nil
}

// ExtendRental extends the current rental session
func (c *ConsoleUseCase) ExtendRental(consoleID int64, additionalMinutes int) error {
	if additionalMinutes <= 0 {
		return errors.NewInvalidDuration()
	}

	console, err := c.consoleRepo.GetByID(consoleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NewConsoleNotFound(consoleID)
		}
		return errors.NewInternalError(err)
	}

	if !console.IsRunning() {
		return errors.NewConsoleNotRunning(console.Name)
	}

	// Extend the rental in the entity
	console.ExtendRental(additionalMinutes)

	// Update the console
	if err := c.consoleRepo.Update(console); err != nil {
		return errors.NewInternalError(err)
	}

	// Update the latest transaction
	transaction, err := c.transactionRepo.GetLast(consoleID)
	if err != nil {
		return errors.NewInternalError(err)
	}

	if transaction != nil {
		transaction.ExtendDuration(additionalMinutes)
		if err := c.transactionRepo.Update(transaction); err != nil {
			return errors.NewInternalError(err)
		}
	}

	return nil
}

// StopRental stops the current rental session
func (c *ConsoleUseCase) StopRental(consoleID int64) error {
	console, err := c.consoleRepo.GetByID(consoleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NewConsoleNotFound(consoleID)
		}
		return errors.NewInternalError(err)
	}

	if !console.IsRunning() {
		return errors.NewConsoleNotRunning(console.Name)
	}

	// Stop the rental in the entity
	console.StopRental()

	// Update the console
	if err := c.consoleRepo.Update(console); err != nil {
		return errors.NewInternalError(err)
	}

	return nil
}

// UpdatePrice updates the hourly price for a console
func (c *ConsoleUseCase) UpdatePrice(consoleID int64, newPrice int) error {
	if newPrice <= 0 {
		return errors.NewInvalidPrice()
	}

	err := c.consoleRepo.UpdatePrice(consoleID, newPrice)
	if err != nil {
		return errors.NewInternalError(err)
	}
	return nil
}

// GetDueSoon returns consoles that will expire soon
func (c *ConsoleUseCase) GetDueSoon(threshold time.Duration) ([]entities.Console, error) {
	consoles, err := c.consoleRepo.GetDueSoon(threshold)
	if err != nil {
		return nil, errors.NewInternalError(err)
	}
	return consoles, nil
}

// CheckExpiredRentals checks and stops expired rental sessions
func (c *ConsoleUseCase) CheckExpiredRentals() ([]entities.Console, error) {
	consoles, err := c.consoleRepo.GetAll()
	if err != nil {
		return nil, errors.NewInternalError(err)
	}

	var expiredConsoles []entities.Console
	for _, console := range consoles {
		if console.IsExpired() {
			console.StopRental()
			if err := c.consoleRepo.Update(&console); err != nil {
				return nil, errors.NewInternalError(err)
			}
			expiredConsoles = append(expiredConsoles, console)
		}
	}

	return expiredConsoles, nil
}
