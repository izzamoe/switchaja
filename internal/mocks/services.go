package mocks

import (
	"switchiot/internal/domain/entities"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockConsoleService is a mock implementation of ConsoleService
type MockConsoleService struct {
	mock.Mock
}

func (m *MockConsoleService) GetAllConsoles() ([]entities.Console, error) {
	args := m.Called()
	return args.Get(0).([]entities.Console), args.Error(1)
}

func (m *MockConsoleService) StartRental(consoleID int64, durationMinutes int) error {
	args := m.Called(consoleID, durationMinutes)
	return args.Error(0)
}

func (m *MockConsoleService) ExtendRental(consoleID int64, additionalMinutes int) error {
	args := m.Called(consoleID, additionalMinutes)
	return args.Error(0)
}

func (m *MockConsoleService) StopRental(consoleID int64) error {
	args := m.Called(consoleID)
	return args.Error(0)
}

func (m *MockConsoleService) UpdatePrice(consoleID int64, newPrice int) error {
	args := m.Called(consoleID, newPrice)
	return args.Error(0)
}

func (m *MockConsoleService) GetDueSoon(threshold time.Duration) ([]entities.Console, error) {
	args := m.Called(threshold)
	return args.Get(0).([]entities.Console), args.Error(1)
}

func (m *MockConsoleService) CheckExpiredRentals() ([]entities.Console, error) {
	args := m.Called()
	return args.Get(0).([]entities.Console), args.Error(1)
}

// MockUserService is a mock implementation of UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(username, password, role string) (int64, error) {
	args := m.Called(username, password, role)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserService) AuthenticateUser(username, password string) (*entities.User, error) {
	args := m.Called(username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserService) GetAllUsers() ([]entities.User, error) {
	args := m.Called()
	return args.Get(0).([]entities.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserService) GetUserCount() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

// MockTransactionService is a mock implementation of TransactionService
type MockTransactionService struct {
	mock.Mock
}

func (m *MockTransactionService) GetTransactionsByConsole(consoleID int64) ([]entities.Transaction, error) {
	args := m.Called(consoleID)
	return args.Get(0).([]entities.Transaction), args.Error(1)
}

func (m *MockTransactionService) GetLastTransaction(consoleID int64) (*entities.Transaction, error) {
	args := m.Called(consoleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Transaction), args.Error(1)
}
