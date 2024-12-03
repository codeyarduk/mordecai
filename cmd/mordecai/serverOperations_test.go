package main

import "testing"

func TestExtractRepoNameFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "GitHub HTTPS URL",
			url:      "https://github.com/codeyarduk/mordecai.git",
			expected: "mordecai",
		},
		{
			name:     "GitHub URL without .git",
			url:      "https://github.com/codeyarduk/mordecai",
			expected: "mordecai",
		},
		{
			name:     "URL with trailing spaces",
			url:      "https://github.com/codeyarduk/mordecai.git  ",
			expected: "mordecai",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRepoNameFromURL(tt.url)
			if result != tt.expected {
				t.Errorf("extractRepoNameFromURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}
