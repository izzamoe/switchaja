package entities

import (
	"time"
)

// Console represents a game console (e.g., PS1, PS2) with rental status.
type Console struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	EndTime      time.Time `json:"end_time"`
	PricePerHour int       `json:"price_per_hour"`
}

// ConsoleStatus constants
const (
	StatusIdle    = "IDLE"
	StatusRunning = "RUNNING"
)

// IsRunning returns true if the console is currently running
func (c *Console) IsRunning() bool {
	return c.Status == StatusRunning
}

// IsExpired returns true if the console rental has expired
func (c *Console) IsExpired() bool {
	return c.IsRunning() && !c.EndTime.IsZero() && time.Now().After(c.EndTime)
}

// TimeRemaining returns the time remaining for the rental
func (c *Console) TimeRemaining() time.Duration {
	if !c.IsRunning() || c.EndTime.IsZero() {
		return 0
	}
	remaining := c.EndTime.Sub(time.Now())
	if remaining < 0 {
		return 0
	}
	return remaining
}

// StartRental starts a rental session for the specified duration
func (c *Console) StartRental(durationMinutes int) {
	c.Status = StatusRunning
	c.EndTime = time.Now().Add(time.Duration(durationMinutes) * time.Minute)
}

// ExtendRental extends the current rental by the specified minutes
func (c *Console) ExtendRental(additionalMinutes int) {
	if c.IsRunning() {
		c.EndTime = c.EndTime.Add(time.Duration(additionalMinutes) * time.Minute)
	}
}

// StopRental stops the current rental session
func (c *Console) StopRental() {
	c.Status = StatusIdle
	c.EndTime = time.Time{}
}