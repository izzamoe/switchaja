package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculatePrice(t *testing.T) {
	tests := []struct {
		name            string
		pricePerHour    int
		durationMinutes int
		expected        int
	}{
		{
			name:            "one hour",
			pricePerHour:    60000,
			durationMinutes: 60,
			expected:        60000,
		},
		{
			name:            "30 minutes",
			pricePerHour:    60000,
			durationMinutes: 30,
			expected:        30000,
		},
		{
			name:            "90 minutes",
			pricePerHour:    60000,
			durationMinutes: 90,
			expected:        90000,
		},
		{
			name:            "15 minutes",
			pricePerHour:    40000,
			durationMinutes: 15,
			expected:        10000,
		},
		{
			name:            "zero duration",
			pricePerHour:    60000,
			durationMinutes: 0,
			expected:        0,
		},
		{
			name:            "zero price",
			pricePerHour:    0,
			durationMinutes: 60,
			expected:        0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculatePrice(tt.pricePerHour, tt.durationMinutes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewTransaction(t *testing.T) {
	consoleID := int64(1)
	durationMinutes := 30
	pricePerHour := 40000

	startTime := time.Now()
	transaction := NewTransaction(consoleID, durationMinutes, pricePerHour)

	require.NotNil(t, transaction)
	assert.Equal(t, consoleID, transaction.ConsoleID)
	assert.Equal(t, durationMinutes, transaction.DurationMin)
	assert.Equal(t, pricePerHour, transaction.PricePerHourSnapshot)

	// Check start time is approximately now
	assert.InDelta(t, startTime.Unix(), transaction.StartTime.Unix(), 2)

	// Check end time is start time + duration
	expectedEndTime := transaction.StartTime.Add(time.Duration(durationMinutes) * time.Minute)
	assert.Equal(t, expectedEndTime, transaction.EndTime)

	// Check calculated price
	expectedPrice := CalculatePrice(pricePerHour, durationMinutes)
	assert.Equal(t, expectedPrice, transaction.TotalPrice)
}

func TestTransaction_UpdateDuration(t *testing.T) {
	startTime := time.Now()
	originalDuration := 30
	newDuration := 45
	pricePerHour := 40000

	transaction := &Transaction{
		ConsoleID:            1,
		StartTime:            startTime,
		EndTime:              startTime.Add(time.Duration(originalDuration) * time.Minute),
		DurationMin:          originalDuration,
		PricePerHourSnapshot: pricePerHour,
		TotalPrice:           CalculatePrice(pricePerHour, originalDuration),
	}

	transaction.UpdateDuration(newDuration)

	assert.Equal(t, newDuration, transaction.DurationMin)

	expectedEndTime := startTime.Add(time.Duration(newDuration) * time.Minute)
	assert.Equal(t, expectedEndTime, transaction.EndTime)

	expectedPrice := CalculatePrice(pricePerHour, newDuration)
	assert.Equal(t, expectedPrice, transaction.TotalPrice)

	// Start time should remain unchanged
	assert.Equal(t, startTime, transaction.StartTime)
}

func TestTransaction_ExtendDuration(t *testing.T) {
	startTime := time.Now()
	originalDuration := 30
	additionalMinutes := 15
	pricePerHour := 40000

	transaction := &Transaction{
		ConsoleID:            1,
		StartTime:            startTime,
		EndTime:              startTime.Add(time.Duration(originalDuration) * time.Minute),
		DurationMin:          originalDuration,
		PricePerHourSnapshot: pricePerHour,
		TotalPrice:           CalculatePrice(pricePerHour, originalDuration),
	}

	transaction.ExtendDuration(additionalMinutes)

	expectedDuration := originalDuration + additionalMinutes
	assert.Equal(t, expectedDuration, transaction.DurationMin)

	expectedEndTime := startTime.Add(time.Duration(expectedDuration) * time.Minute)
	assert.Equal(t, expectedEndTime, transaction.EndTime)

	expectedPrice := CalculatePrice(pricePerHour, expectedDuration)
	assert.Equal(t, expectedPrice, transaction.TotalPrice)
}

func TestTransaction_ExtendDuration_ZeroExtension(t *testing.T) {
	startTime := time.Now()
	originalDuration := 30
	pricePerHour := 40000

	transaction := &Transaction{
		ConsoleID:            1,
		StartTime:            startTime,
		EndTime:              startTime.Add(time.Duration(originalDuration) * time.Minute),
		DurationMin:          originalDuration,
		PricePerHourSnapshot: pricePerHour,
		TotalPrice:           CalculatePrice(pricePerHour, originalDuration),
	}

	originalEndTime := transaction.EndTime
	originalPrice := transaction.TotalPrice

	transaction.ExtendDuration(0)

	// Should remain unchanged
	assert.Equal(t, originalDuration, transaction.DurationMin)
	assert.Equal(t, originalEndTime, transaction.EndTime)
	assert.Equal(t, originalPrice, transaction.TotalPrice)
}

func TestTransaction_ExtendDuration_NegativeExtension(t *testing.T) {
	startTime := time.Now()
	originalDuration := 30
	negativeExtension := -10
	pricePerHour := 40000

	transaction := &Transaction{
		ConsoleID:            1,
		StartTime:            startTime,
		EndTime:              startTime.Add(time.Duration(originalDuration) * time.Minute),
		DurationMin:          originalDuration,
		PricePerHourSnapshot: pricePerHour,
		TotalPrice:           CalculatePrice(pricePerHour, originalDuration),
	}

	transaction.ExtendDuration(negativeExtension)

	expectedDuration := originalDuration + negativeExtension // 20 minutes
	assert.Equal(t, expectedDuration, transaction.DurationMin)

	expectedEndTime := startTime.Add(time.Duration(expectedDuration) * time.Minute)
	assert.Equal(t, expectedEndTime, transaction.EndTime)

	expectedPrice := CalculatePrice(pricePerHour, expectedDuration)
	assert.Equal(t, expectedPrice, transaction.TotalPrice)
}

func TestTransaction_CompleteWorkflow(t *testing.T) {
	// Create new transaction
	consoleID := int64(2)
	initialDuration := 60
	pricePerHour := 50000

	transaction := NewTransaction(consoleID, initialDuration, pricePerHour)

	// Verify initial state
	assert.Equal(t, consoleID, transaction.ConsoleID)
	assert.Equal(t, initialDuration, transaction.DurationMin)
	assert.Equal(t, CalculatePrice(pricePerHour, initialDuration), transaction.TotalPrice)

	// Extend by 30 minutes
	transaction.ExtendDuration(30)
	expectedDuration := initialDuration + 30
	assert.Equal(t, expectedDuration, transaction.DurationMin)
	assert.Equal(t, CalculatePrice(pricePerHour, expectedDuration), transaction.TotalPrice)

	// Update to a specific duration
	newDuration := 120
	transaction.UpdateDuration(newDuration)
	assert.Equal(t, newDuration, transaction.DurationMin)
	assert.Equal(t, CalculatePrice(pricePerHour, newDuration), transaction.TotalPrice)
}
