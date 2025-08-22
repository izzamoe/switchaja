package entities

import (
	"time"
)

// Transaction records a rental usage window for a console.
type Transaction struct {
	ID                   int64     `json:"id"`
	ConsoleID            int64     `json:"console_id"`
	StartTime            time.Time `json:"start_time"`
	EndTime              time.Time `json:"end_time"`
	DurationMin          int       `json:"duration_minutes"`
	TotalPrice           int       `json:"total_price"`
	PricePerHourSnapshot int       `json:"price_per_hour"`
}

// CalculatePrice calculates the total price based on duration and hourly rate
func CalculatePrice(pricePerHour, durationMinutes int) int {
	return (pricePerHour * durationMinutes) / 60
}

// NewTransaction creates a new transaction
func NewTransaction(consoleID int64, durationMinutes, pricePerHour int) *Transaction {
	now := time.Now()
	return &Transaction{
		ConsoleID:            consoleID,
		StartTime:            now,
		EndTime:              now.Add(time.Duration(durationMinutes) * time.Minute),
		DurationMin:          durationMinutes,
		TotalPrice:           CalculatePrice(pricePerHour, durationMinutes),
		PricePerHourSnapshot: pricePerHour,
	}
}

// UpdateDuration updates the transaction duration and recalculates the total price
func (t *Transaction) UpdateDuration(newDurationMinutes int) {
	t.DurationMin = newDurationMinutes
	t.EndTime = t.StartTime.Add(time.Duration(newDurationMinutes) * time.Minute)
	t.TotalPrice = CalculatePrice(t.PricePerHourSnapshot, newDurationMinutes)
}

// ExtendDuration extends the transaction by additional minutes
func (t *Transaction) ExtendDuration(additionalMinutes int) {
	newDuration := t.DurationMin + additionalMinutes
	t.UpdateDuration(newDuration)
}
