package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsValidToken(t *testing.T) {
	// Clear sessions before test
	sessions.m = make(map[string]sessionData)

	t.Run("empty token", func(t *testing.T) {
		result := IsValidToken("")
		assert.False(t, result)
	})

	t.Run("non-existent token", func(t *testing.T) {
		result := IsValidToken("nonexistent-token")
		assert.False(t, result)
	})

	t.Run("valid token", func(t *testing.T) {
		// Add a session
		testToken := "test-token-123"
		sessions.m[testToken] = sessionData{
			UserID:   1,
			Username: "testuser",
			Role:     "user",
			IssuedAt: time.Now().Unix(),
		}

		result := IsValidToken(testToken)
		assert.True(t, result)
	})
}

func TestNewToken(t *testing.T) {
	username := "testuser"
	
	// Generate two tokens
	token1 := newToken(username)
	time.Sleep(time.Nanosecond) // Ensure different timestamps
	token2 := newToken(username)

	// Tokens should be different
	assert.NotEqual(t, token1, token2)
	
	// Both tokens should contain the username
	assert.Contains(t, token1, username)
	assert.Contains(t, token2, username)
	
	// Tokens should be non-empty
	assert.NotEmpty(t, token1)
	assert.NotEmpty(t, token2)
}

func TestNewToken_Format(t *testing.T) {
	username := "admin"
	token := newToken(username)
	
	// Token should have the format: hexNumber-username
	assert.Contains(t, token, "-")
	assert.Contains(t, token, username)
	
	// Split and verify format
	parts := len(token)
	assert.Greater(t, parts, len(username)+1) // Should be longer than just username + dash
}