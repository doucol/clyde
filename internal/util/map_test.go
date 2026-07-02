package util

import (
	"testing"
)

func TestGetMapKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]int
		expected []string
	}{
		{
			name:     "empty map",
			input:    map[string]int{},
			expected: []string{},
		},
		{
			name: "map with values",
			input: map[string]int{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMapKeys(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("GetMapKeys() length = %v; want %v", len(result), len(tt.expected))
			}
			// Create a map to check if all expected keys are present
			expectedMap := make(map[string]bool)
			for _, k := range tt.expected {
				expectedMap[k] = true
			}
			for _, k := range result {
				if !expectedMap[k] {
					t.Errorf("GetMapKeys() returned unexpected key: %v", k)
				}
			}
		})
	}
}
