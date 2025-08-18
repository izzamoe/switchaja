package iot

import (
	"sync"
)

// Package iot contains abstractions for sending commands to physical
// relay / switch hardware (e.g. ESP8266). The current implementation is
// a lightweight in-memory mock; swap with WebSocket, HTTP, or MQTT
// sender to integrate real devices.

// CommandSender dispatches ON/OFF (or future) commands to a device.
// consoleID maps a logical console to the hardware endpoint.
type CommandSender interface {
	Send(consoleID int64, cmd string) error
}

// MockSender simple stdout or stub (can be replaced).
// Provided so rest of system can depend on abstraction.

// MockSender records last command per console; used for tests / demo.
type MockSender struct {
	mu   sync.Mutex
	last map[int64]string
}

// NewMockSender constructs a new mock implementation.
func NewMockSender() *MockSender { return &MockSender{last: make(map[int64]string)} }

// Send stores the command for later inspection.
func (m *MockSender) Send(consoleID int64, cmd string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.last[consoleID] = cmd
	return nil
}

// Last returns the last recorded command for a console.
func (m *MockSender) Last(consoleID int64) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.last[consoleID]
}
