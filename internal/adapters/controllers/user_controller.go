package controllers

import (
	"net/http"
	"strconv"
	"switchiot/internal/domain/errors"
	"switchiot/internal/domain/usecases"

	"github.com/gofiber/fiber/v2"
)

// UserController handles user-related HTTP requests
type UserController struct {
	userService usecases.UserService
}

// NewUserController creates a new user controller
func NewUserController(userService usecases.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// Login authenticates a user
func (uc *UserController) Login(c *fiber.Ctx) error {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid request body")
	}

	user, err := uc.userService.AuthenticateUser(request.Username, request.Password)
	if err != nil {
		return uc.handleError(c, err)
	}

	// Here you would typically create a session token
	// For now, return user info
	return c.JSON(fiber.Map{
		"status": "ok",
		"user":   user,
	})
}

// GetAllUsers returns all users (admin only)
func (uc *UserController) GetAllUsers(c *fiber.Ctx) error {
	users, err := uc.userService.GetAllUsers()
	if err != nil {
		return uc.handleError(c, err)
	}
	return c.JSON(users)
}

// CreateUser creates a new user (admin only)
func (uc *UserController) CreateUser(c *fiber.Ctx) error {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid request body")
	}

	userID, err := uc.userService.CreateUser(request.Username, request.Password, request.Role)
	if err != nil {
		return uc.handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"status":  "ok",
		"user_id": userID,
	})
}

// DeleteUser deletes a user (admin only)
func (uc *UserController) DeleteUser(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid user ID")
	}

	if err := uc.userService.DeleteUser(userID); err != nil {
		return uc.handleError(c, err)
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

// handleError converts domain errors to appropriate HTTP responses
func (uc *UserController) handleError(c *fiber.Ctx, err error) error {
	if domainErr, ok := err.(*errors.DomainError); ok {
		switch domainErr.Code {
		case errors.CodeUserNotFound:
			return fiber.NewError(http.StatusNotFound, domainErr.Message)
		case errors.CodeInvalidCredentials:
			return fiber.NewError(http.StatusUnauthorized, domainErr.Message)
		case errors.CodeUserAlreadyExists:
			return fiber.NewError(http.StatusConflict, domainErr.Message)
		case errors.CodeInvalidUserData:
			return fiber.NewError(http.StatusBadRequest, domainErr.Message)
		case errors.CodeUnauthorized:
			return fiber.NewError(http.StatusUnauthorized, domainErr.Message)
		case errors.CodeInternalError:
			return fiber.NewError(http.StatusInternalServerError, "internal server error")
		default:
			return fiber.NewError(http.StatusBadRequest, domainErr.Message)
		}
	}
	return fiber.NewError(http.StatusInternalServerError, "internal server error")
}