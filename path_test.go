package json

import (
	"strings"
	"testing"
)

func TestJSONPath(t *testing.T) {
	t.Parallel()
	data := `{
		"name": "John Doe",
		"age": 30,
		"city": "New York",
		"hobbies": [
		  "reading",
		  "traveling",
		  "photography"
		],
		"education": {
		  "degree": "Bachelor's",
		  "major": "Computer Science",
		  "university": "ABC University"
		},
		"work": [
		  {
			"company": "XYZ Corp",
			"position": "Software Engineer",
			"years": 5
		  },
		  {
			"company": "123 Inc",
			"position": "Senior Developer",
			"years": 3
		  }
		],
		"married": false,
		"friends": [
		  {
			"name": "Alice",
			"age": 28
		  },
		  {
			"name": "Bob",
			"age": 32
		  }
		]
	  }`

	tests := []struct {
		name     string
		path     string
		expected string
		isErr    bool
	}{
		{name: "root", path: "$", expected: "[$]"},
		{name: "roots", path: "$.", expected: "[$]"},
		{name: "by key", path: "$.education.degree", expected: "[$['education']['degree']]"},
		{name: "all key", path: "$..degree", expected: "[$['education']['degree']]"},
		{name: "all key with bracket", path: "$..['degree']", expected: "[$['education']['degree']]"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			res, err := JSONPath([]byte(data), tt.path)
			if (err != nil) != tt.isErr {
				t.Errorf("expected error %v, got %v", tt.isErr, err)
			}
			if tt.isErr {
				return
			}
			if fullPath(res) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, fullPath(res))
			}
		})
	}
}

func TestParseJSONPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path     string
		expected []string
	}{
		{path: "$", expected: []string{"$"}},
		{path: "$.", expected: []string{"$"}},
		{path: "$..", expected: []string{"$", ".."}},
		{path: "$.*", expected: []string{"$", "*"}},
		{path: "$..*", expected: []string{"$", "..", "*"}},
		{path: "$.root.element", expected: []string{"$", "root", "element"}},
		{path: "$.root.*.element", expected: []string{"$", "root", "*", "element"}},
		{path: "$['root']['element']", expected: []string{"$", "root", "element"}},
		{path: "$['root'][*]['element']", expected: []string{"$", "root", "*", "element"}},
		{path: "$['store']['book'][0]['title']", expected: []string{"$", "store", "book", "0", "title"}},
		{path: "$['root'].*['element']", expected: []string{"$", "root", "*", "element"}},
		{path: "$.['root'].*.['element']", expected: []string{"$", "root", "*", "element"}},
		{path: "$['root'].*.['element']", expected: []string{"$", "root", "*", "element"}},
		{path: "$.phoneNumbers[*].type", expected: []string{"$", "phoneNumbers", "*", "type"}},
		{path: "$.store.book[?(@.price < 10)].title", expected: []string{"$", "store", "book", "?(@.price < 10)", "title"}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.path, func(t *testing.T) {
			result, err := ParsePath(tt.path)
			if err != nil {
				t.Errorf("error on path %s: %s", tt.path, err.Error())
			} else if !sliceEqual(result, tt.expected) {
				t.Errorf("expected %s, got %s", sliceString(tt.expected), sliceString(result))
			}
		})
	}
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func fullPath(array []*Node) string {
	return sliceString(Paths(array))
}

func sliceString(array []string) string {
	return "[" + strings.Join(array, ", ") + "]"
}
