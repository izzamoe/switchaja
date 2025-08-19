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

func TestUserController_NewUserController(t *testing.T) {
	userService := &mocks.MockUserService{}

	controller := NewUserController(userService)
	assert.NotNil(t, controller)
}

func TestUserController_Login(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := fiber.New()
		userService := &mocks.MockUserService{}
		controller := NewUserController(userService)

		user := &entities.User{
			ID:       1,
			Username: "testuser",
			Role:     entities.RoleUser,
		}

		userService.On("AuthenticateUser", "testuser", "password123").Return(user, nil)

		app.Post("/login", controller.Login)

		requestBody := map[string]interface{}{
			"username": "testuser",
			"password": "password123",
		}
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "ok", result["status"])
		assert.NotNil(t, result["user"])

		userService.AssertExpectations(t)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		app := fiber.New()
		userService := &mocks.MockUserService{}
		controller := NewUserController(userService)

		serviceError := domainErrors.NewInvalidCredentials()
		userService.On("AuthenticateUser", "testuser", "wrongpass").Return(nil, serviceError)

		app.Post("/login", controller.Login)

		requestBody := map[string]interface{}{
			"username": "testuser",
			"password": "wrongpass",
		}
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		userService.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		app := fiber.New()
		userService := &mocks.MockUserService{}
		controller := NewUserController(userService)

		app.Post("/login", controller.Login)

		req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestUserController_GetAllUsers(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := fiber.New()
		userService := &mocks.MockUserService{}
		controller := NewUserController(userService)

		users := []entities.User{
			{ID: 1, Username: "admin", Role: entities.RoleAdmin},
			{ID: 2, Username: "user1", Role: entities.RoleUser},
		}

		userService.On("GetAllUsers").Return(users, nil)

		app.Get("/users", controller.GetAllUsers)

		req := httptest.NewRequest("GET", "/users", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result []entities.User
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, users, result)

		userService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		app := fiber.New()
		userService := &mocks.MockUserService{}
		controller := NewUserController(userService)

		serviceError := domainErrors.NewInternalError(fmt.Errorf("database error"))
		userService.On("GetAllUsers").Return([]entities.User(nil), serviceError)

		app.Get("/users", controller.GetAllUsers)

		req := httptest.NewRequest("GET", "/users", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		userService.AssertExpectations(t)
	})
}

func TestUserController_CreateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := fiber.New()
		userService := &mocks.MockUserService{}
		controller := NewUserController(userService)

		userService.On("CreateUser", "newuser", "password123", entities.RoleUser).Return(int64(5), nil)

		app.Post("/users", controller.CreateUser)

		requestBody := map[string]interface{}{
			"username": "newuser",
			"password": "password123",
			"role":     entities.RoleUser,
		}
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "ok", result["status"])
		assert.Equal(t, float64(5), result["user_id"]) // JSON numbers are floats

		userService.AssertExpectations(t)
	})

	t.Run("user already exists", func(t *testing.T) {
		app := fiber.New()
		userService := &mocks.MockUserService{}
		controller := NewUserController(userService)

		serviceError := domainErrors.NewUserAlreadyExists("existinguser")
		userService.On("CreateUser", "existinguser", "password123", entities.RoleUser).Return(int64(0), serviceError)

		app.Post("/users", controller.CreateUser)

		requestBody := map[string]interface{}{
			"username": "existinguser",
			"password": "password123",
			"role":     entities.RoleUser,
		}
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusConflict, resp.StatusCode)

		userService.AssertExpectations(t)
	})

	t.Run("invalid user data", func(t *testing.T) {
		app := fiber.New()
		userService := &mocks.MockUserService{}
		controller := NewUserController(userService)

		serviceError := domainErrors.NewInvalidUserData("username cannot be empty")
		userService.On("CreateUser", "", "password123", entities.RoleUser).Return(int64(0), serviceError)

		app.Post("/users", controller.CreateUser)

		requestBody := map[string]interface{}{
			"username": "",
			"password": "password123",
			"role":     entities.RoleUser,
		}
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		userService.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		app := fiber.New()
		userService := &mocks.MockUserService{}
		controller := NewUserController(userService)

		app.Post("/users", controller.CreateUser)

		req := httptest.NewRequest("POST", "/users", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestUserController_DeleteUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := fiber.New()
		userService := &mocks.MockUserService{}
		controller := NewUserController(userService)

		userService.On("DeleteUser", int64(5)).Return(nil)

		app.Delete("/users/:id", controller.DeleteUser)

		req := httptest.NewRequest("DELETE", "/users/5", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]string
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "ok", result["status"])

		userService.AssertExpectations(t)
	})

	t.Run("invalid user ID", func(t *testing.T) {
		app := fiber.New()
		userService := &mocks.MockUserService{}
		controller := NewUserController(userService)

		app.Delete("/users/:id", controller.DeleteUser)

		req := httptest.NewRequest("DELETE", "/users/invalid", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("user not found", func(t *testing.T) {
		app := fiber.New()
		userService := &mocks.MockUserService{}
		controller := NewUserController(userService)

		serviceError := domainErrors.NewUserNotFound()
		userService.On("DeleteUser", int64(999)).Return(serviceError)

		app.Delete("/users/:id", controller.DeleteUser)

		req := httptest.NewRequest("DELETE", "/users/999", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		userService.AssertExpectations(t)
	})
}

func TestUserController_HandleError(t *testing.T) {
	t.Run("unauthorized error", func(t *testing.T) {
		app := fiber.New()
		userService := &mocks.MockUserService{}
		controller := NewUserController(userService)

		serviceError := domainErrors.NewUnauthorized()
		userService.On("GetAllUsers").Return([]entities.User(nil), serviceError)

		app.Get("/users", controller.GetAllUsers)

		req := httptest.NewRequest("GET", "/users", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		userService.AssertExpectations(t)
	})

	t.Run("unknown domain error", func(t *testing.T) {
		app := fiber.New()
		userService := &mocks.MockUserService{}
		controller := NewUserController(userService)

		// Create a custom domain error with unknown code
		serviceError := &domainErrors.DomainError{
			Code:    "UNKNOWN_ERROR",
			Message: "unknown error",
		}
		userService.On("GetAllUsers").Return([]entities.User(nil), serviceError)

		app.Get("/users", controller.GetAllUsers)

		req := httptest.NewRequest("GET", "/users", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		userService.AssertExpectations(t)
	})

	t.Run("non-domain error", func(t *testing.T) {
		app := fiber.New()
		userService := &mocks.MockUserService{}
		controller := NewUserController(userService)

		// Regular error, not a domain error
		serviceError := fmt.Errorf("some other error")
		userService.On("GetAllUsers").Return([]entities.User(nil), serviceError)

		app.Get("/users", controller.GetAllUsers)

		req := httptest.NewRequest("GET", "/users", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		userService.AssertExpectations(t)
	})
}