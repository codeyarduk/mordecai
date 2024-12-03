package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTokenOperations(t *testing.T) {
	// Setup test environment
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	t.Run("Test Save and Load Token", func(t *testing.T) {
		testToken := "test-token-123"

		// Test saving token
		err := saveToken(testToken)
		if err != nil {
			t.Errorf("Failed to save token: %v", err)
		}

		// Test loading token
		loadedToken, err := loadToken()
		if err != nil {
			t.Errorf("Failed to load token: %v", err)
		}

		if loadedToken != testToken {
			t.Errorf("Loaded token doesn't match saved token. Got %s, want %s",
				loadedToken, testToken)
		}
	})

	t.Run("Test Token Validation", func(t *testing.T) {
		valid, err := checkIfTokenIsValid()
		if err != nil {
			t.Errorf("Error checking token validity: %v", err)
		}

		if !valid {
			t.Error("Token should be valid but was reported as invalid")
		}
	})

	t.Run("Test Delete Token", func(t *testing.T) {
		err := deleteToken()
		if err != nil {
			t.Errorf("Failed to delete token: %v", err)
		}

		// Verify token is deleted
		token, err := loadToken()
		if err != nil {
			t.Errorf("Error loading token after deletion: %v", err)
		}
		if token != "" {
			t.Error("Token should be empty after deletion")
		}
	})
}
