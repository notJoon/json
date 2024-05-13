package json

import (
	"testing"
)

func TestParseFilterExpression(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected []string
		isErr    bool
	}{
		{
			name:     "Simple expression",
			input:    "?(@.name == 'John')",
			expected: []string{"@.name", "'John'", "=="},
		},
		{
			name:     "Expression with multiple conditions",
			input:    "?(@.age > 30 && @.city == 'New York')",
			expected: []string{"@.age", "30", ">", "@.city", "'New York'", "==", "&&"},
		},
		{
			name:     "Expression with parentheses",
			input:    "?(@.age > 30 && (@.city == 'New York' || @.city == 'London'))",
			expected: []string{"@.age", "30", ">", "@.city", "'New York'", "==", "@.city", "'London'", "==", "||", "&&"},
		},
		{
			name:     "length",
			input:    "@.length",
			expected: []string{"@.length"},
		},
		{
			name:     "length 2",
			input:    "@.length - 1",
			expected: []string{"@.length", "1", "-"},
		},
		{
			name:     "Invalid expression - mismatched parentheses",
			input:    "?(@.age > 30 && (@.city == 'New York')",
			expected: nil,
			isErr:    true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseFilterExpression(tt.input)
			if tt.isErr {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(result) != len(tt.expected) {
					t.Errorf("Length mismatch: expected %d, got %d", len(tt.expected), len(result))
				}
				if !equalSlices(result, tt.expected) {
					t.Errorf("Expected %v, but got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestTokenizeExpression(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Simple expression",
			input:    "@.name == 'John'",
			expected: []string{"@.name", "==", "'John'"},
		},
		{
			name:     "simple greater expression",
			input:    "@.age > 30",
			expected: []string{"@.age", ">", "30"},
		},
		{
			name:     "simple and expression",
			input:    "@.age > 30 && @.city == 'New York'",
			expected: []string{"@.age", ">", "30", "&&", "@.city", "==", "'New York'"},
		},
		{
			name:     "simple less than expression",
			input:    "@.age < 18",
			expected: []string{"@.age", "<", "18"},
		},
		{
			name:     "simple or expression",
			input:    "@.city == 'Los Angeles' || @.city == 'San Francisco'",
			expected: []string{"@.city", "==", "'Los Angeles'", "||", "@.city", "==", "'San Francisco'"},
		},
		{
			name:     "complex expression",
			input:    "(@.age > 30 && @.city == 'New York') || (@.age < 18 && @.city == 'Los Angeles')",
			expected: []string{"(", "@.age", ">", "30", "&&", "@.city", "==", "'New York'", ")", "||", "(", "@.age", "<", "18", "&&", "@.city", "==", "'Los Angeles'", ")"},
		},
		{
			name:     "expression with not equal",
			input:    "@.name != 'John'",
			expected: []string{"@.name", "!=", "'John'"},
		},
		{
			name:     "expression with greater than or equal",
			input:    "@.age >= 21",
			expected: []string{"@.age", ">=", "21"},
		},
		{
			name:     "expression with less than or equal",
			input:    "@.age <= 21",
			expected: []string{"@.age", "<=", "21"},
		},
		{
			name:     "length",
			input:    "@.length",
			expected: []string{"@.length"},
		},
		{
			name:     "length 2",
			input:    "@.length - 1",
			expected: []string{"@.length", "-", "1"},
		},
		{
			name:     "length 3 without space",
			input:    "@.length-1",
			expected: []string{"@.length", "-", "1"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			result, err := tokenizeExpression(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if len(result) != len(tt.expected) {
				t.Errorf("Length mismatch: expected %d, got %d", len(tt.expected), len(result))
			}
			for i := 0; i < len(tt.expected) && i < len(result); i++ {
				if result[i] != tt.expected[i] {
					t.Errorf("Mismatch at position %d: expected %v, got %v", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

func TestConvertToRPN(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    []string
		expected []string
		hasError bool
	}{
		{
			name:     "Simple expression",
			input:    []string{"@.name", "==", "'John'"},
			expected: []string{"@.name", "'John'", "=="},
			hasError: false,
		},
		{
			name:     "Expression with multiple conditions",
			input:    []string{"@.age", ">", "30", "&&", "@.city", "==", "'New York'"},
			expected: []string{"@.age", "30", ">", "@.city", "'New York'", "==", "&&"},
			hasError: false,
		},
		{
			name:     "Expression with parentheses",
			input:    []string{"(", "@.age", ">", "30", ")", "&&", "(", "@.city", "==", "'New York'", "||", "@.city", "==", "'London'", ")"},
			expected: []string{"@.age", "30", ">", "@.city", "'New York'", "==", "@.city", "'London'", "==", "||", "&&"},
			hasError: false,
		},
		{
			name:     "Invalid expression - mismatched parentheses",
			input:    []string{"(", "@.age", ">", "30", "&&", "@.city", "==", "'New York'"},
			expected: nil,
			hasError: true,
		},
		{
			name:     "length",
			input:    []string{"@.length", "-", "1"},
			expected: []string{"@.length", "1", "-"},
		},
		{
			name:     "simple math expression",
			input:    []string{"(", "3", "+", "5", ")", "*", "2"},
			expected: []string{"3", "5", "+", "2", "*"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertToRPN(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !equalSlices(result, tt.expected) {
					t.Errorf("Expected %v, but got %v", tt.expected, result)
				}
			}
		})
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
