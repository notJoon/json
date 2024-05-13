package json

import (
	"testing"
)

func TestEqualOperation(t *testing.T) {
	tests := []struct {
		name     string
		node1    *Node
		node2    *Node
		expected bool
	}{
		{
			name:     "equal strings",
			node1:    StringNode("", "test"),
			node2:    StringNode("", "test"),
			expected: true,
		},
		{
			name:     "different strings",
			node1:    StringNode("", "test"),
			node2:    StringNode("", "different"),
			expected: false,
		},
		{
			name:     "equal numbers",
			node1:    NumberNode("", 123),
			node2:    NumberNode("", 123),
			expected: true,
		},
		{
			name:     "different numbers",
			node1:    NumberNode("", 123),
			node2:    NumberNode("", 456),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := equalOperation(test.node1, test.node2)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if result.value != test.expected {
				t.Errorf("wrong result: got %v, expected %v", result.value, test.expected)
			}
		})
	}
}

func TestNotEqualOperation(t *testing.T) {
	tests := []struct {
		name     string
		node1    *Node
		node2    *Node
		expected bool
	}{
		{
			name:     "equal strings",
			node1:    StringNode("", "test"),
			node2:    StringNode("", "test"),
			expected: false,
		},
		{
			name:     "different strings",
			node1:    StringNode("", "test"),
			node2:    StringNode("", "different"),
			expected: true,
		},
		{
			name:     "equal numbers",
			node1:    NumberNode("", 123),
			node2:    NumberNode("", 123),
			expected: false,
		},
		{
			name:     "different numbers",
			node1:    NumberNode("", 123),
			node2:    NumberNode("", 456),
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := notEqualOperation(test.node1, test.node2)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if result.value != test.expected {
				t.Errorf("wrong result: got %v, expected %v", result.value, test.expected)
			}
		})
	}
}

func TestGreaterThanOperation(t *testing.T) {
	tests := []struct {
		name     string
		node1    *Node
		node2    *Node
		expected bool
	}{
		{
			name:     "greater numbers",
			node1:    NumberNode("", 456),
			node2:    NumberNode("", 123),
			expected: true,
		},
		{
			name:     "less numbers",
			node1:    NumberNode("", 123),
			node2:    NumberNode("", 456),
			expected: false,
		},
		{
			name:     "equal numbers",
			node1:    NumberNode("", 123),
			node2:    NumberNode("", 123),
			expected: false,
		},
		{
			name:     "greater strings",
			node1:    StringNode("", "test2"),
			node2:    StringNode("", "test1"),
			expected: true,
		},
		{
			name:     "less strings",
			node1:    StringNode("", "test1"),
			node2:    StringNode("", "test2"),
			expected: false,
		},
		{
			name:     "equal strings",
			node1:    StringNode("", "test"),
			node2:    StringNode("", "test"),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := greaterThanOperation(test.node1, test.node2)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if result.value != test.expected {
				t.Errorf("wrong result: got %v, expected %v", result.value, test.expected)
			}
		})
	}
}

func TestGreaterThanEqualOperation(t *testing.T) {
	tests := []struct {
		name     string
		node1    *Node
		node2    *Node
		expected bool
	}{
		{
			name:     "greater numbers",
			node1:    NumberNode("", 456),
			node2:    NumberNode("", 123),
			expected: true,
		},
		{
			name:     "equal numbers",
			node1:    NumberNode("", 123),
			node2:    NumberNode("", 123),
			expected: true,
		},
		{
			name:     "less numbers",
			node1:    NumberNode("", 123),
			node2:    NumberNode("", 456),
			expected: false,
		},
		{
			name:     "greater strings",
			node1:    StringNode("", "test2"),
			node2:    StringNode("", "test1"),
			expected: true,
		},
		{
			name:     "equal strings",
			node1:    StringNode("", "test"),
			node2:    StringNode("", "test"),
			expected: true,
		},
		{
			name:     "less strings",
			node1:    StringNode("", "test1"),
			node2:    StringNode("", "test2"),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := greaterThanEqualOperation(test.node1, test.node2)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if result.value != test.expected {
				t.Errorf("wrong result: got %v, expected %v", result.value, test.expected)
			}
		})
	}
}

func TestLessThanOperation(t *testing.T) {
	tests := []struct {
		name     string
		node1    *Node
		node2    *Node
		expected bool
	}{
		{
			name:     "less numbers",
			node1:    NumberNode("", 123),
			node2:    NumberNode("", 456),
			expected: true,
		},
		{
			name:     "equal numbers",
			node1:    NumberNode("", 123),
			node2:    NumberNode("", 123),
			expected: false,
		},
		{
			name:     "greater numbers",
			node1:    NumberNode("", 456),
			node2:    NumberNode("", 123),
			expected: false,
		},
		{
			name:     "less strings",
			node1:    StringNode("", "test1"),
			node2:    StringNode("", "test2"),
			expected: true,
		},
		{
			name:     "equal strings",
			node1:    StringNode("", "test"),
			node2:    StringNode("", "test"),
			expected: false,
		},
		{
			name:     "greater strings",
			node1:    StringNode("", "test2"),
			node2:    StringNode("", "test1"),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := lessThanOperation(test.node1, test.node2)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if result.value != test.expected {
				t.Errorf("wrong result: got %v, expected %v", result.value, test.expected)
			}
		})
	}
}

func TestLessThanOrEqualOperation(t *testing.T) {
	tests := []struct {
		name     string
		node1    *Node
		node2    *Node
		expected bool
	}{
		{
			name:     "less numbers",
			node1:    NumberNode("", 123),
			node2:    NumberNode("", 456),
			expected: true,
		},
		{
			name:     "equal numbers",
			node1:    NumberNode("", 123),
			node2:    NumberNode("", 123),
			expected: true,
		},
		{
			name:     "greater numbers",
			node1:    NumberNode("", 456),
			node2:    NumberNode("", 123),
			expected: false,
		},
		{
			name:     "less strings",
			node1:    StringNode("", "test1"),
			node2:    StringNode("", "test2"),
			expected: true,
		},
		{
			name:     "equal strings",
			node1:    StringNode("", "test"),
			node2:    StringNode("", "test"),
			expected: true,
		},
		{
			name:     "greater strings",
			node1:    StringNode("", "test2"),
			node2:    StringNode("", "test1"),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := lessThanOrEqualOperation(test.node1, test.node2)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if result.value != test.expected {
				t.Errorf("wrong result: got %v, expected %v", result.value, test.expected)
			}
		})
	}
}

func TestOrOperation(t *testing.T) {
	tests := []struct {
		name     string
		lhs      *Node
		rhs      *Node
		expected bool
	}{
		{
			name:     "true OR true",
			lhs:      BoolNode("", true),
			rhs:      BoolNode("", true),
			expected: true,
		},
		{
			name:     "true OR false",
			lhs:      BoolNode("", true),
			rhs:      BoolNode("", false),
			expected: true,
		},
		{
			name:     "false OR true",
			lhs:      BoolNode("", false),
			rhs:      BoolNode("", true),
			expected: true,
		},
		{
			name:     "false OR false",
			lhs:      BoolNode("", false),
			rhs:      BoolNode("", false),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := orOperation(tt.lhs, tt.rhs)
			if err != nil {
				t.Errorf("orOperation() error = %v", err)
				return
			}
			boolResult, err := result.GetBool()
			if err != nil {
				t.Errorf("GetBool() error = %v", err)
				return
			}
			if boolResult != tt.expected {
				t.Errorf("orOperation() = %v, want %v", boolResult, tt.expected)
			}
		})
	}
}

func TestAndOperation(t *testing.T) {
	tests := []struct {
		name     string
		lhs      *Node
		rhs      *Node
		expected bool
	}{
		{
			name:     "true and true",
			lhs:      BoolNode("", true),
			rhs:      BoolNode("", true),
			expected: true,
		},
		{
			name:     "true and false",
			lhs:      BoolNode("", true),
			rhs:      BoolNode("", false),
			expected: false,
		},
		{
			name:     "false and true",
			lhs:      BoolNode("", false),
			rhs:      BoolNode("", true),
			expected: false,
		},
		{
			name:     "false and false",
			lhs:      BoolNode("", false),
			rhs:      BoolNode("", false),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := andOperation(tt.lhs, tt.rhs)
			if err != nil {
				t.Errorf("andOperation() error = %v", err)
				return
			}
			if result.value != tt.expected {
				t.Errorf("andOperation() = %v, want %v", result.value, tt.expected)
			}
		})
	}
}
