package json

import (
	"bytes"
	"testing"
)

func TestIndentJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:  "Empty object",
			input: []byte(`{}`),
			expected: []byte(`{}`),
		},
		{
			name:  "Simple object",
			input: []byte(`{"name":"John"}`),
			expected: []byte(`{
  "name": "John"
}`),
		},
		{
			name:  "Nested object",
			input: []byte(`{"person":{"name":"John","age":30}}`),
			expected: []byte(`{
  "person": {
    "name": "John",
    "age": 30
  }
}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := Indent(tt.input)
			if err != nil {
				t.Errorf("IndentJSON() error = %v", err)
				return
			}
			if !bytes.Equal(actual, tt.expected) {
				t.Errorf("IndentJSON() = %q, want %q", actual, tt.expected)
			}
		})
	}
}
