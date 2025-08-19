package usecases

import (
	"fmt"
	"switchiot/internal/domain/entities"
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
	return c.consoleRepo.GetAll()
}

// StartRental starts a rental session for a console
func (c *ConsoleUseCase) StartRental(consoleID int64, durationMinutes int) error {
	if durationMinutes <= 0 {
		return fmt.Errorf("duration must be positive")
	}

	console, err := c.consoleRepo.GetByID(consoleID)
	if err != nil {
		return fmt.Errorf("failed to get console: %w", err)
	}

	if console.IsRunning() {
		return fmt.Errorf("console %s is already running", console.Name)
	}

	// Start the rental in the entity
	console.StartRental(durationMinutes)

	// Update the console
	if err := c.consoleRepo.Update(console); err != nil {
		return fmt.Errorf("failed to update console: %w", err)
	}

	// Create transaction record
	transaction := entities.NewTransaction(consoleID, durationMinutes, console.PricePerHour)
	if err := c.transactionRepo.Create(transaction); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// ExtendRental extends the current rental session
func (c *ConsoleUseCase) ExtendRental(consoleID int64, additionalMinutes int) error {
	if additionalMinutes <= 0 {
		return fmt.Errorf("additional minutes must be positive")
	}

	console, err := c.consoleRepo.GetByID(consoleID)
	if err != nil {
		return fmt.Errorf("failed to get console: %w", err)
	}

	if !console.IsRunning() {
		return fmt.Errorf("console %s is not running", console.Name)
	}

	// Extend the rental in the entity
	console.ExtendRental(additionalMinutes)

	// Update the console
	if err := c.consoleRepo.Update(console); err != nil {
		return fmt.Errorf("failed to update console: %w", err)
	}

	// Update the latest transaction
	transaction, err := c.transactionRepo.GetLast(consoleID)
	if err != nil {
		return fmt.Errorf("failed to get last transaction: %w", err)
	}

	if transaction != nil {
		transaction.ExtendDuration(additionalMinutes)
		if err := c.transactionRepo.Update(transaction); err != nil {
			return fmt.Errorf("failed to update transaction: %w", err)
		}
	}

	return nil
}

// StopRental stops the current rental session
func (c *ConsoleUseCase) StopRental(consoleID int64) error {
	console, err := c.consoleRepo.GetByID(consoleID)
	if err != nil {
		return fmt.Errorf("failed to get console: %w", err)
	}

	if !console.IsRunning() {
		return fmt.Errorf("console %s is not running", console.Name)
	}

	// Stop the rental in the entity
	console.StopRental()

	// Update the console
	if err := c.consoleRepo.Update(console); err != nil {
		return fmt.Errorf("failed to update console: %w", err)
	}

	return nil
}

// UpdatePrice updates the hourly price for a console
func (c *ConsoleUseCase) UpdatePrice(consoleID int64, newPrice int) error {
	if newPrice <= 0 {
		return fmt.Errorf("price must be positive")
	}

	return c.consoleRepo.UpdatePrice(consoleID, newPrice)
}

// GetDueSoon returns consoles that will expire soon
func (c *ConsoleUseCase) GetDueSoon(threshold time.Duration) ([]entities.Console, error) {
	return c.consoleRepo.GetDueSoon(threshold)
}

// CheckExpiredRentals checks and stops expired rental sessions
func (c *ConsoleUseCase) CheckExpiredRentals() ([]entities.Console, error) {
	consoles, err := c.consoleRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get consoles: %w", err)
	}

	var expiredConsoles []entities.Console
	for _, console := range consoles {
		if console.IsExpired() {
			console.StopRental()
			if err := c.consoleRepo.Update(&console); err != nil {
				return nil, fmt.Errorf("failed to update console %s: %w", console.Name, err)
			}
			expiredConsoles = append(expiredConsoles, console)
		}
	}

	return expiredConsoles, nil
}