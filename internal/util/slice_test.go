package util

import (
	"testing"
	"time"
)

type testStruct struct {
	Name      string
	Age       int
	CreatedAt time.Time
}

func TestFilterSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		filter   func(int) bool
		expected []int
	}{
		{
			name:     "empty slice",
			input:    []int{},
			filter:   func(n int) bool { return n > 0 },
			expected: []int{},
		},
		{
			name:     "filter even numbers",
			input:    []int{1, 2, 3, 4, 5, 6},
			filter:   func(n int) bool { return n%2 == 0 },
			expected: []int{2, 4, 6},
		},
		{
			name:     "filter positive numbers",
			input:    []int{-2, -1, 0, 1, 2},
			filter:   func(n int) bool { return n > 0 },
			expected: []int{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterSlice(tt.input, tt.filter)
			if len(result) != len(tt.expected) {
				t.Errorf("FilterSlice() length = %v; want %v", len(result), len(tt.expected))
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("FilterSlice()[%d] = %v; want %v", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestSortSlice(t *testing.T) {
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour)

	tests := []struct {
		name      string
		input     []testStruct
		sortBy    string
		ascending bool
		expected  []testStruct
	}{
		{
			name: "sort by name ascending",
			input: []testStruct{
				{Name: "Charlie", Age: 30, CreatedAt: now},
				{Name: "Alice", Age: 25, CreatedAt: oneHourAgo},
				{Name: "Bob", Age: 35, CreatedAt: twoHoursAgo},
			},
			sortBy:    "Name",
			ascending: true,
			expected: []testStruct{
				{Name: "Alice", Age: 25, CreatedAt: oneHourAgo},
				{Name: "Bob", Age: 35, CreatedAt: twoHoursAgo},
				{Name: "Charlie", Age: 30, CreatedAt: now},
			},
		},
		{
			name: "sort by age descending",
			input: []testStruct{
				{Name: "Charlie", Age: 30, CreatedAt: now},
				{Name: "Alice", Age: 25, CreatedAt: oneHourAgo},
				{Name: "Bob", Age: 35, CreatedAt: twoHoursAgo},
			},
			sortBy:    "Age",
			ascending: false,
			expected: []testStruct{
				{Name: "Bob", Age: 35, CreatedAt: twoHoursAgo},
				{Name: "Charlie", Age: 30, CreatedAt: now},
				{Name: "Alice", Age: 25, CreatedAt: oneHourAgo},
			},
		},
		{
			name: "sort by created_at ascending",
			input: []testStruct{
				{Name: "Charlie", Age: 30, CreatedAt: now},
				{Name: "Alice", Age: 25, CreatedAt: oneHourAgo},
				{Name: "Bob", Age: 35, CreatedAt: twoHoursAgo},
			},
			sortBy:    "CreatedAt",
			ascending: true,
			expected: []testStruct{
				{Name: "Bob", Age: 35, CreatedAt: twoHoursAgo},
				{Name: "Alice", Age: 25, CreatedAt: oneHourAgo},
				{Name: "Charlie", Age: 30, CreatedAt: now},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the input slice to avoid modifying the original
			input := make([]testStruct, len(tt.input))
			copy(input, tt.input)

			SortSlice(input, tt.sortBy, tt.ascending)

			for i := range input {
				if input[i].Name != tt.expected[i].Name {
					t.Errorf("SortSlice()[%d].Name = %v; want %v", i, input[i].Name, tt.expected[i].Name)
				}
				if input[i].Age != tt.expected[i].Age {
					t.Errorf("SortSlice()[%d].Age = %v; want %v", i, input[i].Age, tt.expected[i].Age)
				}
				if !input[i].CreatedAt.Equal(tt.expected[i].CreatedAt) {
					t.Errorf("SortSlice()[%d].CreatedAt = %v; want %v", i, input[i].CreatedAt, tt.expected[i].CreatedAt)
				}
			}
		})
	}
}
