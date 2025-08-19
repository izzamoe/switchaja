package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConsole_IsRunning(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "running console",
			status:   StatusRunning,
			expected: true,
		},
		{
			name:     "idle console",
			status:   StatusIdle,
			expected: false,
		},
		{
			name:     "empty status",
			status:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			console := &Console{Status: tt.status}
			assert.Equal(t, tt.expected, console.IsRunning())
		})
	}
}

func TestConsole_IsExpired(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name     string
		status   string
		endTime  time.Time
		expected bool
	}{
		{
			name:     "running console with future end time",
			status:   StatusRunning,
			endTime:  now.Add(time.Hour),
			expected: false,
		},
		{
			name:     "running console with past end time",
			status:   StatusRunning,
			endTime:  now.Add(-time.Hour),
			expected: true,
		},
		{
			name:     "idle console",
			status:   StatusIdle,
			endTime:  now.Add(-time.Hour),
			expected: false,
		},
		{
			name:     "running console with zero end time",
			status:   StatusRunning,
			endTime:  time.Time{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			console := &Console{
				Status:  tt.status,
				EndTime: tt.endTime,
			}
			assert.Equal(t, tt.expected, console.IsExpired())
		})
	}
}

func TestConsole_TimeRemaining(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name     string
		status   string
		endTime  time.Time
		expected time.Duration
	}{
		{
			name:     "running console with future end time",
			status:   StatusRunning,
			endTime:  now.Add(time.Hour),
			expected: time.Hour,
		},
		{
			name:     "running console with past end time",
			status:   StatusRunning,
			endTime:  now.Add(-time.Hour),
			expected: 0,
		},
		{
			name:     "idle console",
			status:   StatusIdle,
			endTime:  now.Add(time.Hour),
			expected: 0,
		},
		{
			name:     "running console with zero end time",
			status:   StatusRunning,
			endTime:  time.Time{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			console := &Console{
				Status:  tt.status,
				EndTime: tt.endTime,
			}
			remaining := console.TimeRemaining()
			
			// Allow for small time differences due to test execution time
			if tt.expected > 0 {
				assert.InDelta(t, tt.expected.Seconds(), remaining.Seconds(), 1.0)
			} else {
				assert.Equal(t, tt.expected, remaining)
			}
		})
	}
}

func TestConsole_StartRental(t *testing.T) {
	console := &Console{
		ID:           1,
		Name:         "PS1",
		Status:       StatusIdle,
		PricePerHour: 40000,
	}

	startTime := time.Now()
	console.StartRental(30)

	assert.Equal(t, StatusRunning, console.Status)
	assert.True(t, console.EndTime.After(startTime))
	// Should be approximately 30 minutes from now
	expectedEnd := startTime.Add(30 * time.Minute)
	assert.InDelta(t, expectedEnd.Unix(), console.EndTime.Unix(), 2) // 2 second tolerance
}

func TestConsole_ExtendRental(t *testing.T) {
	now := time.Now()
	console := &Console{
		Status:  StatusRunning,
		EndTime: now.Add(time.Hour),
	}

	originalEndTime := console.EndTime
	console.ExtendRental(30)

	expectedEndTime := originalEndTime.Add(30 * time.Minute)
	assert.Equal(t, expectedEndTime, console.EndTime)
}

func TestConsole_ExtendRental_NotRunning(t *testing.T) {
	now := time.Now()
	console := &Console{
		Status:  StatusIdle,
		EndTime: now.Add(time.Hour),
	}

	originalEndTime := console.EndTime
	console.ExtendRental(30)

	// End time should not change if console is not running
	assert.Equal(t, originalEndTime, console.EndTime)
}

func TestConsole_StopRental(t *testing.T) {
	console := &Console{
		Status:  StatusRunning,
		EndTime: time.Now().Add(time.Hour),
	}

	console.StopRental()

	assert.Equal(t, StatusIdle, console.Status)
	assert.True(t, console.EndTime.IsZero())
}

func TestConsole_StatusConstants(t *testing.T) {
	assert.Equal(t, "IDLE", StatusIdle)
	assert.Equal(t, "RUNNING", StatusRunning)
}