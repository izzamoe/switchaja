package entities

import (
	"time"
)

// User represents a system user
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// User roles
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// IsAdmin returns true if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// CanManageUsers returns true if the user can manage other users
func (u *User) CanManageUsers() bool {
	return u.IsAdmin()
}

// CanManagePricing returns true if the user can manage console pricing
func (u *User) CanManagePricing() bool {
	return u.IsAdmin()
}
