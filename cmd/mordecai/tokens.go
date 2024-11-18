package main

import (
	"fmt"
	"os"
	"path/filepath"
)

//  _        _
// | |_ ___ | | _____ _ __  ___
// | __/ _ \| |/ / _ \ '_ \/ __|
// | || (_) |   <  __/ | | \__ \
//  \__\___/|_|\_\___|_| |_|___/
//

func checkIfTokenIsValid() (bool, error) {
	var token, err = loadToken()
	if err != nil {
		return false, err
	}

	if len(token) > 0 {
		return true, nil
	}

	return false, nil
}

func getTokenFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	mordecaiPath := filepath.Join(homeDir, ".mordecai")
	if err := os.MkdirAll(mordecaiPath, 0700); err != nil {
		return "", fmt.Errorf("failed to create .mordecai directory: %w", err)
	}
	return filepath.Join(mordecaiPath, ".mordecai_token"), nil
}

func saveToken(token string) error {
	filePath, err := getTokenFilePath()
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, []byte(token), 0600)
	if err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}
	return nil
}

func loadToken() (string, error) {
	filePath, err := getTokenFilePath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // No token file found, but not an error
		}
		return "", fmt.Errorf("failed to load token: %w", err)
	}
	return string(data), nil
}

func deleteToken() error {
	filePath, err := getTokenFilePath()
	if err != nil {
		return err
	}

	err = os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}
