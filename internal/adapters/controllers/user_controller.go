package controllers

import (
	"net/http"
	"strconv"
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
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	user, err := uc.userService.AuthenticateUser(request.Username, request.Password)
	if err != nil {
		return fiber.NewError(http.StatusUnauthorized, "invalid credentials")
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
		return fiber.NewError(http.StatusInternalServerError, err.Error())
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
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	userID, err := uc.userService.CreateUser(request.Username, request.Password, request.Role)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
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
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(fiber.Map{"status": "ok"})
}