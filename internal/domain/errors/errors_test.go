package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDomainError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *DomainError
		expected string
	}{
		{
			name: "error without cause",
			err: &DomainError{
				Code:    "TEST_CODE",
				Message: "test message",
			},
			expected: "test message",
		},
		{
			name: "error with cause",
			err: &DomainError{
				Code:    "TEST_CODE",
				Message: "test message",
				Cause:   errors.New("underlying error"),
			},
			expected: "test message: underlying error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestDomainError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("underlying error")
	domainErr := &DomainError{
		Code:    "TEST_CODE",
		Message: "test message",
		Cause:   underlyingErr,
	}

	assert.Equal(t, underlyingErr, domainErr.Unwrap())
}

func TestDomainError_Unwrap_NoCause(t *testing.T) {
	domainErr := &DomainError{
		Code:    "TEST_CODE",
		Message: "test message",
	}

	assert.Nil(t, domainErr.Unwrap())
}

// Console error tests
func TestNewConsoleNotFound(t *testing.T) {
	consoleID := int64(123)
	err := NewConsoleNotFound(consoleID)

	assert.Equal(t, CodeConsoleNotFound, err.Code)
	assert.Equal(t, fmt.Sprintf("console with ID %d not found", consoleID), err.Message)
	assert.Nil(t, err.Cause)
}

func TestNewConsoleAlreadyRunning(t *testing.T) {
	consoleName := "PS1"
	err := NewConsoleAlreadyRunning(consoleName)

	assert.Equal(t, CodeConsoleAlreadyRunning, err.Code)
	assert.Equal(t, fmt.Sprintf("console %s is already running", consoleName), err.Message)
	assert.Nil(t, err.Cause)
}

func TestNewConsoleNotRunning(t *testing.T) {
	consoleName := "PS2"
	err := NewConsoleNotRunning(consoleName)

	assert.Equal(t, CodeConsoleNotRunning, err.Code)
	assert.Equal(t, fmt.Sprintf("console %s is not running", consoleName), err.Message)
	assert.Nil(t, err.Cause)
}

func TestNewInvalidDuration(t *testing.T) {
	err := NewInvalidDuration()

	assert.Equal(t, CodeInvalidDuration, err.Code)
	assert.Equal(t, "duration must be positive", err.Message)
	assert.Nil(t, err.Cause)
}

func TestNewInvalidPrice(t *testing.T) {
	err := NewInvalidPrice()

	assert.Equal(t, CodeInvalidPrice, err.Code)
	assert.Equal(t, "price must be positive", err.Message)
	assert.Nil(t, err.Cause)
}

// User error tests
func TestNewUserNotFound(t *testing.T) {
	err := NewUserNotFound()

	assert.Equal(t, CodeUserNotFound, err.Code)
	assert.Equal(t, "user not found", err.Message)
	assert.Nil(t, err.Cause)
}

func TestNewInvalidCredentials(t *testing.T) {
	err := NewInvalidCredentials()

	assert.Equal(t, CodeInvalidCredentials, err.Code)
	assert.Equal(t, "invalid credentials", err.Message)
	assert.Nil(t, err.Cause)
}

func TestNewUserAlreadyExists(t *testing.T) {
	username := "testuser"
	err := NewUserAlreadyExists(username)

	assert.Equal(t, CodeUserAlreadyExists, err.Code)
	assert.Equal(t, fmt.Sprintf("user %s already exists", username), err.Message)
	assert.Nil(t, err.Cause)
}

func TestNewInvalidUserData(t *testing.T) {
	message := "custom validation message"
	err := NewInvalidUserData(message)

	assert.Equal(t, CodeInvalidUserData, err.Code)
	assert.Equal(t, message, err.Message)
	assert.Nil(t, err.Cause)
}

func TestNewUnauthorized(t *testing.T) {
	err := NewUnauthorized()

	assert.Equal(t, CodeUnauthorized, err.Code)
	assert.Equal(t, "unauthorized access", err.Message)
	assert.Nil(t, err.Cause)
}

// Transaction error tests
func TestNewTransactionNotFound(t *testing.T) {
	consoleID := int64(456)
	err := NewTransactionNotFound(consoleID)

	assert.Equal(t, CodeTransactionNotFound, err.Code)
	assert.Equal(t, fmt.Sprintf("no transaction found for console %d", consoleID), err.Message)
	assert.Nil(t, err.Cause)
}

// General error tests
func TestNewInternalError(t *testing.T) {
	cause := errors.New("database connection failed")
	err := NewInternalError(cause)

	assert.Equal(t, CodeInternalError, err.Code)
	assert.Equal(t, "internal system error", err.Message)
	assert.Equal(t, cause, err.Cause)
}

func TestNewValidationError(t *testing.T) {
	message := "validation failed for field X"
	err := NewValidationError(message)

	assert.Equal(t, CodeValidationError, err.Code)
	assert.Equal(t, message, err.Message)
	assert.Nil(t, err.Cause)
}

// Test error code constants
func TestErrorCodes(t *testing.T) {
	// Console error codes
	assert.Equal(t, "CONSOLE_NOT_FOUND", CodeConsoleNotFound)
	assert.Equal(t, "CONSOLE_ALREADY_RUNNING", CodeConsoleAlreadyRunning)
	assert.Equal(t, "CONSOLE_NOT_RUNNING", CodeConsoleNotRunning)
	assert.Equal(t, "INVALID_DURATION", CodeInvalidDuration)
	assert.Equal(t, "INVALID_PRICE", CodeInvalidPrice)

	// User error codes
	assert.Equal(t, "USER_NOT_FOUND", CodeUserNotFound)
	assert.Equal(t, "INVALID_CREDENTIALS", CodeInvalidCredentials)
	assert.Equal(t, "USER_ALREADY_EXISTS", CodeUserAlreadyExists)
	assert.Equal(t, "INVALID_USER_DATA", CodeInvalidUserData)
	assert.Equal(t, "UNAUTHORIZED", CodeUnauthorized)

	// Transaction error codes
	assert.Equal(t, "TRANSACTION_NOT_FOUND", CodeTransactionNotFound)
	assert.Equal(t, "INVALID_TRANSACTION", CodeInvalidTransaction)

	// General error codes
	assert.Equal(t, "INTERNAL_ERROR", CodeInternalError)
	assert.Equal(t, "VALIDATION_ERROR", CodeValidationError)
}