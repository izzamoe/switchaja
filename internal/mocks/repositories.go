package mocks

import (
	"switchiot/internal/domain/entities"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockConsoleRepository is a mock implementation of ConsoleRepository
type MockConsoleRepository struct {
	mock.Mock
}

func (m *MockConsoleRepository) GetAll() ([]entities.Console, error) {
	args := m.Called()
	return args.Get(0).([]entities.Console), args.Error(1)
}

func (m *MockConsoleRepository) GetByID(id int64) (*entities.Console, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Console), args.Error(1)
}

func (m *MockConsoleRepository) Update(console *entities.Console) error {
	args := m.Called(console)
	return args.Error(0)
}

func (m *MockConsoleRepository) UpdatePrice(consoleID int64, newPrice int) error {
	args := m.Called(consoleID, newPrice)
	return args.Error(0)
}

func (m *MockConsoleRepository) GetDueSoon(threshold time.Duration) ([]entities.Console, error) {
	args := m.Called(threshold)
	return args.Get(0).([]entities.Console), args.Error(1)
}

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(username, passwordHash, role string) (int64, error) {
	args := m.Called(username, passwordHash, role)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(username string) (*entities.User, string, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(*entities.User), args.String(1), args.Error(2)
}

func (m *MockUserRepository) GetAll() ([]entities.User, error) {
	args := m.Called()
	return args.Get(0).([]entities.User), args.Error(1)
}

func (m *MockUserRepository) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) Count() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

// MockTransactionRepository is a mock implementation of TransactionRepository
type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Create(transaction *entities.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) GetByConsoleID(consoleID int64, limit int) ([]entities.Transaction, error) {
	args := m.Called(consoleID, limit)
	return args.Get(0).([]entities.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetLast(consoleID int64) (*entities.Transaction, error) {
	args := m.Called(consoleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) Update(transaction *entities.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}
