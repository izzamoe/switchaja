package controllers

import (
	"net/http"
	"strconv"
	"switchiot/internal/domain/usecases"

	"github.com/gofiber/fiber/v2"
)

// ConsoleController handles console-related HTTP requests
type ConsoleController struct {
	consoleService     usecases.ConsoleService
	transactionService usecases.TransactionService
}

// NewConsoleController creates a new console controller
func NewConsoleController(consoleService usecases.ConsoleService, transactionService usecases.TransactionService) *ConsoleController {
	return &ConsoleController{
		consoleService:     consoleService,
		transactionService: transactionService,
	}
}

// GetStatus returns the current status of all consoles
func (cc *ConsoleController) GetStatus(c *fiber.Ctx) error {
	consoles, err := cc.consoleService.GetAllConsoles()
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(consoles)
}

// StartRental starts a rental session for a console
func (cc *ConsoleController) StartRental(c *fiber.Ctx) error {
	var request struct {
		ConsoleID   int64 `json:"console_id"`
		DurationMin int   `json:"duration_minutes"`
	}

	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	if err := cc.consoleService.StartRental(request.ConsoleID, request.DurationMin); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

// ExtendRental extends the current rental session
func (cc *ConsoleController) ExtendRental(c *fiber.Ctx) error {
	var request struct {
		ConsoleID  int64 `json:"console_id"`
		AddMinutes int   `json:"add_minutes"`
	}

	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	if err := cc.consoleService.ExtendRental(request.ConsoleID, request.AddMinutes); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

// StopRental stops the current rental session
func (cc *ConsoleController) StopRental(c *fiber.Ctx) error {
	var request struct {
		ConsoleID int64 `json:"console_id"`
	}

	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	if err := cc.consoleService.StopRental(request.ConsoleID); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

// UpdatePrice updates the hourly price for a console
func (cc *ConsoleController) UpdatePrice(c *fiber.Ctx) error {
	var request struct {
		ConsoleID int64 `json:"console_id"`
		NewPrice  int   `json:"new_price"`
	}

	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	if err := cc.consoleService.UpdatePrice(request.ConsoleID, request.NewPrice); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

// GetTransactions returns transactions for a specific console
func (cc *ConsoleController) GetTransactions(c *fiber.Ctx) error {
	consoleIDStr := c.Params("console_id")
	consoleID, err := strconv.ParseInt(consoleIDStr, 10, 64)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid console ID")
	}

	transactions, err := cc.transactionService.GetTransactionsByConsole(consoleID)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(transactions)
}