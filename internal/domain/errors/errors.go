package errors

import "fmt"

// DomainError represents a domain-specific error
type DomainError struct {
	Code    string
	Message string
	Cause   error
}

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *DomainError) Unwrap() error {
	return e.Cause
}

// Error codes
const (
	// Console errors
	CodeConsoleNotFound       = "CONSOLE_NOT_FOUND"
	CodeConsoleAlreadyRunning = "CONSOLE_ALREADY_RUNNING"
	CodeConsoleNotRunning     = "CONSOLE_NOT_RUNNING"
	CodeInvalidDuration       = "INVALID_DURATION"
	CodeInvalidPrice          = "INVALID_PRICE"

	// User errors
	CodeUserNotFound       = "USER_NOT_FOUND"
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
	CodeUserAlreadyExists  = "USER_ALREADY_EXISTS"
	CodeInvalidUserData    = "INVALID_USER_DATA"
	CodeUnauthorized       = "UNAUTHORIZED"

	// Transaction errors
	CodeTransactionNotFound = "TRANSACTION_NOT_FOUND"
	CodeInvalidTransaction  = "INVALID_TRANSACTION"

	// General errors
	CodeInternalError   = "INTERNAL_ERROR"
	CodeValidationError = "VALIDATION_ERROR"
)

// Console errors
func NewConsoleNotFound(consoleID int64) *DomainError {
	return &DomainError{
		Code:    CodeConsoleNotFound,
		Message: fmt.Sprintf("console with ID %d not found", consoleID),
	}
}

func NewConsoleAlreadyRunning(consoleName string) *DomainError {
	return &DomainError{
		Code:    CodeConsoleAlreadyRunning,
		Message: fmt.Sprintf("console %s is already running", consoleName),
	}
}

func NewConsoleNotRunning(consoleName string) *DomainError {
	return &DomainError{
		Code:    CodeConsoleNotRunning,
		Message: fmt.Sprintf("console %s is not running", consoleName),
	}
}

func NewInvalidDuration() *DomainError {
	return &DomainError{
		Code:    CodeInvalidDuration,
		Message: "duration must be positive",
	}
}

func NewInvalidPrice() *DomainError {
	return &DomainError{
		Code:    CodeInvalidPrice,
		Message: "price must be positive",
	}
}

// User errors
func NewUserNotFound() *DomainError {
	return &DomainError{
		Code:    CodeUserNotFound,
		Message: "user not found",
	}
}

func NewInvalidCredentials() *DomainError {
	return &DomainError{
		Code:    CodeInvalidCredentials,
		Message: "invalid credentials",
	}
}

func NewUserAlreadyExists(username string) *DomainError {
	return &DomainError{
		Code:    CodeUserAlreadyExists,
		Message: fmt.Sprintf("user %s already exists", username),
	}
}

func NewInvalidUserData(message string) *DomainError {
	return &DomainError{
		Code:    CodeInvalidUserData,
		Message: message,
	}
}

func NewUnauthorized() *DomainError {
	return &DomainError{
		Code:    CodeUnauthorized,
		Message: "unauthorized access",
	}
}

// Transaction errors
func NewTransactionNotFound(consoleID int64) *DomainError {
	return &DomainError{
		Code:    CodeTransactionNotFound,
		Message: fmt.Sprintf("no transaction found for console %d", consoleID),
	}
}

// General errors
func NewInternalError(cause error) *DomainError {
	return &DomainError{
		Code:    CodeInternalError,
		Message: "internal system error",
		Cause:   cause,
	}
}

func NewValidationError(message string) *DomainError {
	return &DomainError{
		Code:    CodeValidationError,
		Message: message,
	}
}
