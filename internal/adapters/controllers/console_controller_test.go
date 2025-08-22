package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"switchiot/internal/domain/entities"
	domainErrors "switchiot/internal/domain/errors"
	"switchiot/internal/mocks"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestConsoleController_NewConsoleController(t *testing.T) {
	consoleService := &mocks.MockConsoleService{}
	transactionService := &mocks.MockTransactionService{}

	controller := NewConsoleController(consoleService, transactionService)
	assert.NotNil(t, controller)
}

func TestConsoleController_GetStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := fiber.New()
		consoleService := &mocks.MockConsoleService{}
		transactionService := &mocks.MockTransactionService{}
		controller := NewConsoleController(consoleService, transactionService)

		consoles := []entities.Console{
			{ID: 1, Name: "PS1", Status: entities.StatusIdle},
			{ID: 2, Name: "PS2", Status: entities.StatusRunning},
		}

		consoleService.On("GetAllConsoles").Return(consoles, nil)

		app.Get("/status", controller.GetStatus)

		req := httptest.NewRequest("GET", "/status", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result []entities.Console
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, consoles, result)

		consoleService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		app := fiber.New()
		consoleService := &mocks.MockConsoleService{}
		transactionService := &mocks.MockTransactionService{}
		controller := NewConsoleController(consoleService, transactionService)

		serviceError := domainErrors.NewInternalError(fmt.Errorf("database error"))
		consoleService.On("GetAllConsoles").Return([]entities.Console(nil), serviceError)

		app.Get("/status", controller.GetStatus)

		req := httptest.NewRequest("GET", "/status", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		consoleService.AssertExpectations(t)
	})
}

func TestConsoleController_StartRental(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := fiber.New()
		consoleService := &mocks.MockConsoleService{}
		transactionService := &mocks.MockTransactionService{}
		controller := NewConsoleController(consoleService, transactionService)

		requestBody := map[string]interface{}{
			"console_id":       1,
			"duration_minutes": 30,
		}

		consoleService.On("StartRental", int64(1), 30).Return(nil)

		app.Post("/start", controller.StartRental)

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/start", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]string
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "ok", result["status"])

		consoleService.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		app := fiber.New()
		consoleService := &mocks.MockConsoleService{}
		transactionService := &mocks.MockTransactionService{}
		controller := NewConsoleController(consoleService, transactionService)

		app.Post("/start", controller.StartRental)

		req := httptest.NewRequest("POST", "/start", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("console not found", func(t *testing.T) {
		app := fiber.New()
		consoleService := &mocks.MockConsoleService{}
		transactionService := &mocks.MockTransactionService{}
		controller := NewConsoleController(consoleService, transactionService)

		requestBody := map[string]interface{}{
			"console_id":       999,
			"duration_minutes": 30,
		}

		serviceError := domainErrors.NewConsoleNotFound(999)
		consoleService.On("StartRental", int64(999), 30).Return(serviceError)

		app.Post("/start", controller.StartRental)

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/start", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		consoleService.AssertExpectations(t)
	})
}

func TestConsoleController_ExtendRental(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := fiber.New()
		consoleService := &mocks.MockConsoleService{}
		transactionService := &mocks.MockTransactionService{}
		controller := NewConsoleController(consoleService, transactionService)

		requestBody := map[string]interface{}{
			"console_id":  1,
			"add_minutes": 15,
		}

		consoleService.On("ExtendRental", int64(1), 15).Return(nil)

		app.Post("/extend", controller.ExtendRental)

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/extend", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		consoleService.AssertExpectations(t)
	})

	t.Run("console not running", func(t *testing.T) {
		app := fiber.New()
		consoleService := &mocks.MockConsoleService{}
		transactionService := &mocks.MockTransactionService{}
		controller := NewConsoleController(consoleService, transactionService)

		requestBody := map[string]interface{}{
			"console_id":  1,
			"add_minutes": 15,
		}

		serviceError := domainErrors.NewConsoleNotRunning("PS1")
		consoleService.On("ExtendRental", int64(1), 15).Return(serviceError)

		app.Post("/extend", controller.ExtendRental)

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/extend", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusConflict, resp.StatusCode)

		consoleService.AssertExpectations(t)
	})
}

func TestConsoleController_StopRental(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := fiber.New()
		consoleService := &mocks.MockConsoleService{}
		transactionService := &mocks.MockTransactionService{}
		controller := NewConsoleController(consoleService, transactionService)

		requestBody := map[string]interface{}{
			"console_id": 1,
		}

		consoleService.On("StopRental", int64(1)).Return(nil)

		app.Post("/stop", controller.StopRental)

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/stop", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		consoleService.AssertExpectations(t)
	})
}

func TestConsoleController_UpdatePrice(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := fiber.New()
		consoleService := &mocks.MockConsoleService{}
		transactionService := &mocks.MockTransactionService{}
		controller := NewConsoleController(consoleService, transactionService)

		requestBody := map[string]interface{}{
			"console_id": 1,
			"new_price":  50000,
		}

		consoleService.On("UpdatePrice", int64(1), 50000).Return(nil)

		app.Post("/update-price", controller.UpdatePrice)

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/update-price", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		consoleService.AssertExpectations(t)
	})

	t.Run("invalid price", func(t *testing.T) {
		app := fiber.New()
		consoleService := &mocks.MockConsoleService{}
		transactionService := &mocks.MockTransactionService{}
		controller := NewConsoleController(consoleService, transactionService)

		requestBody := map[string]interface{}{
			"console_id": 1,
			"new_price":  0,
		}

		serviceError := domainErrors.NewInvalidPrice()
		consoleService.On("UpdatePrice", int64(1), 0).Return(serviceError)

		app.Post("/update-price", controller.UpdatePrice)

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/update-price", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		consoleService.AssertExpectations(t)
	})
}

func TestConsoleController_GetTransactions(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := fiber.New()
		consoleService := &mocks.MockConsoleService{}
		transactionService := &mocks.MockTransactionService{}
		controller := NewConsoleController(consoleService, transactionService)

		transactions := []entities.Transaction{
			{ID: 1, ConsoleID: 1, DurationMin: 30, TotalPrice: 20000},
			{ID: 2, ConsoleID: 1, DurationMin: 60, TotalPrice: 40000},
		}

		transactionService.On("GetTransactionsByConsole", int64(1)).Return(transactions, nil)

		app.Get("/transactions/:console_id", controller.GetTransactions)

		req := httptest.NewRequest("GET", "/transactions/1", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result []entities.Transaction
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, transactions, result)

		transactionService.AssertExpectations(t)
	})

	t.Run("invalid console ID", func(t *testing.T) {
		app := fiber.New()
		consoleService := &mocks.MockConsoleService{}
		transactionService := &mocks.MockTransactionService{}
		controller := NewConsoleController(consoleService, transactionService)

		app.Get("/transactions/:console_id", controller.GetTransactions)

		req := httptest.NewRequest("GET", "/transactions/invalid", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestConsoleController_HandleError(t *testing.T) {
	app := fiber.New()
	consoleService := &mocks.MockConsoleService{}
	transactionService := &mocks.MockTransactionService{}
	controller := NewConsoleController(consoleService, transactionService)

	t.Run("console already running error", func(t *testing.T) {
		serviceError := domainErrors.NewConsoleAlreadyRunning("PS1")
		consoleService.On("StartRental", int64(1), 30).Return(serviceError)

		app.Post("/start", controller.StartRental)

		requestBody := map[string]interface{}{
			"console_id":       1,
			"duration_minutes": 30,
		}
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/start", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
		consoleService.AssertExpectations(t)
	})

	t.Run("invalid duration error", func(t *testing.T) {
		serviceError := domainErrors.NewInvalidDuration()
		consoleService.On("StartRental", int64(1), 0).Return(serviceError)

		app.Post("/start", controller.StartRental)

		requestBody := map[string]interface{}{
			"console_id":       1,
			"duration_minutes": 0,
		}
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/start", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		consoleService.AssertExpectations(t)
	})
}
