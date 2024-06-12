package json

// Stack represents a stack data structure.
type Stack struct {
	items []string
}

// NewStack creates a new stack.
func NewStack() *Stack {
	return &Stack{
		items: []string{},
	}
}

// Push adds an item to the top of the stack.
func (s *Stack) Push(item string) {
	s.items = append(s.items, item)
}

// Pop removes and returns the top item from the stack.
func (s *Stack) Pop() string {
	if s.IsEmpty() {
		return ""
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item
}

// Peek returns the top item from the stack without removing it.
func (s *Stack) Peek() string {
	if s.IsEmpty() {
		return ""
	}
	return s.items[len(s.items)-1]
}

// IsEmpty checks if the stack is empty.
func (s *Stack) IsEmpty() bool {
	return len(s.items) == 0
}
