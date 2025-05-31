package util

import (
	"errors"
	"testing"
)

func TestIsErr(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")

	tests := []struct {
		name     string
		this     error
		those    []error
		expected bool
	}{
		{
			name:     "matching error",
			this:     err1,
			those:    []error{err1, err2},
			expected: true,
		},
		{
			name:     "non-matching error",
			this:     err3,
			those:    []error{err1, err2},
			expected: false,
		},
		{
			name:     "empty error list",
			this:     err1,
			those:    []error{},
			expected: false,
		},
		{
			name:     "nil error",
			this:     nil,
			those:    []error{err1, err2},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsErr(tt.this, tt.those...)
			if result != tt.expected {
				t.Errorf("IsErr() = %v; want %v", result, tt.expected)
			}
		})
	}
}
