package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUser_IsAdmin(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{
			name:     "admin user",
			role:     RoleAdmin,
			expected: true,
		},
		{
			name:     "regular user",
			role:     RoleUser,
			expected: false,
		},
		{
			name:     "empty role",
			role:     "",
			expected: false,
		},
		{
			name:     "invalid role",
			role:     "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			assert.Equal(t, tt.expected, user.IsAdmin())
		})
	}
}

func TestUser_CanManageUsers(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{
			name:     "admin can manage users",
			role:     RoleAdmin,
			expected: true,
		},
		{
			name:     "regular user cannot manage users",
			role:     RoleUser,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			assert.Equal(t, tt.expected, user.CanManageUsers())
		})
	}
}

func TestUser_CanManagePricing(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{
			name:     "admin can manage pricing",
			role:     RoleAdmin,
			expected: true,
		},
		{
			name:     "regular user cannot manage pricing",
			role:     RoleUser,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			assert.Equal(t, tt.expected, user.CanManagePricing())
		})
	}
}

func TestUser_RoleConstants(t *testing.T) {
	assert.Equal(t, "admin", RoleAdmin)
	assert.Equal(t, "user", RoleUser)
}

func TestUser_Creation(t *testing.T) {
	now := time.Now()
	user := User{
		ID:        1,
		Username:  "testuser",
		Role:      RoleUser,
		CreatedAt: now,
	}

	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, RoleUser, user.Role)
	assert.Equal(t, now, user.CreatedAt)
	assert.False(t, user.IsAdmin())
	assert.False(t, user.CanManageUsers())
	assert.False(t, user.CanManagePricing())
}

func TestUser_AdminPrivileges(t *testing.T) {
	admin := User{
		ID:       1,
		Username: "admin",
		Role:     RoleAdmin,
	}

	assert.True(t, admin.IsAdmin())
	assert.True(t, admin.CanManageUsers())
	assert.True(t, admin.CanManagePricing())
}
