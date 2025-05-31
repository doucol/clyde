package util

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	// Test with environment variable set
	os.Setenv("TEST_ENV_VAR", "test_value")
	defer os.Unsetenv("TEST_ENV_VAR")

	tests := []struct {
		name     string
		key      string
		fallback string
		expected string
	}{
		{
			name:     "existing env var",
			key:      "TEST_ENV_VAR",
			fallback: "fallback_value",
			expected: "test_value",
		},
		{
			name:     "non-existing env var",
			key:      "NON_EXISTING_VAR",
			fallback: "fallback_value",
			expected: "fallback_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetEnv(tt.key, tt.fallback)
			if result != tt.expected {
				t.Errorf("GetEnv() = %v; want %v", result, tt.expected)
			}
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	// Test with environment variable set
	os.Setenv("TEST_INT_VAR", "42")
	defer os.Unsetenv("TEST_INT_VAR")

	tests := []struct {
		name     string
		key      string
		fallback int
		expected int
	}{
		{
			name:     "existing int env var",
			key:      "TEST_INT_VAR",
			fallback: 0,
			expected: 42,
		},
		{
			name:     "non-existing env var",
			key:      "NON_EXISTING_VAR",
			fallback: 10,
			expected: 10,
		},
		{
			name:     "invalid int env var",
			key:      "TEST_INT_VAR",
			fallback: 10,
			expected: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetEnvInt(tt.key, tt.fallback)
			if result != tt.expected {
				t.Errorf("GetEnvInt() = %v; want %v", result, tt.expected)
			}
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		fallback bool
		expected bool
	}{
		{
			name:     "true value",
			key:      "TEST_BOOL_VAR",
			value:    "true",
			fallback: false,
			expected: true,
		},
		{
			name:     "yes value",
			key:      "TEST_BOOL_VAR",
			value:    "yes",
			fallback: false,
			expected: true,
		},
		{
			name:     "1 value",
			key:      "TEST_BOOL_VAR",
			value:    "1",
			fallback: false,
			expected: true,
		},
		{
			name:     "false value",
			key:      "TEST_BOOL_VAR",
			value:    "false",
			fallback: true,
			expected: false,
		},
		{
			name:     "non-existing env var",
			key:      "NON_EXISTING_VAR",
			value:    "",
			fallback: true,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			result := GetEnvBool(tt.key, tt.fallback)
			if result != tt.expected {
				t.Errorf("GetEnvBool() = %v; want %v", result, tt.expected)
			}
		})
	}
}
