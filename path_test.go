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
		// basic path
		{name: "root", path: "$", expected: "[$]"},
		{name: "roots", path: "$.", expected: "[$]"},
		{name: "by key", path: "$.education.degree", expected: "[$['education']['degree']]"},
		{name: "all key", path: "$..degree", expected: "[$['education']['degree']]"},
		{name: "all key with bracket", path: "$..['degree']", expected: "[$['education']['degree']]"},

		// array path
		{name: "slice array", path: "$.hobbies[0:2]", expected: "[$['hobbies'][0], $['hobbies'][1]]"},
		{name: "slice array with step", path: "$.hobbies[0:3:2]", expected: "[$['hobbies'][0], $['hobbies'][2]]"},
		{name: "slice array with negative index and step", path: "$.hobbies[-3:-1:1]", expected: "[$['hobbies'][0], $['hobbies'][1]]"},
		{name: "slice object array", path: "$.work[0:2].company", expected: "[$['work'][0]['company'], $['work'][1]['company']]"},
		{name: "slice object array with step", path: "$.work[0:2:2].position", expected: "[$['work'][0]['position']]"},
		{name: "slice object array with negative index and step", path: "$.friends[-2:-1:1].age", expected: "[$['friends'][0]['age']]"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			res, err := Path([]byte(data), tt.path)
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

func TestArrayPath(t *testing.T) {
	t.Parallel()
	data := `{
		"numbers": [1, 2, 3, 4, 5],
		"colors": ["red", "green", "blue"],
		"nested": [
			{
				"name": "Alice",
				"age": 30
			},
			{
				"name": "Bob",
				"age": 25
			},
			{
				"name": "Charlie",
				"age": 35
			}
		]
	}`

	tests := []struct {
		name     string
		path     string
		expected string
		isErr    bool
	}{
		{name: "array index", path: "$.numbers[2]", expected: "[$['numbers'][2]]"},
		{name: "array slice", path: "$.numbers[1:4]", expected: "[$['numbers'][1], $['numbers'][2], $['numbers'][3]]"},
		{name: "array slice with step", path: "$.numbers[0:5:2]", expected: "[$['numbers'][0], $['numbers'][2], $['numbers'][4]]"},
		{name: "array slice with negative index and step", path: "$.colors[-3:-1:1]", expected: "[$['colors'][0], $['colors'][1]]"},
		{name: "nested array slice", path: "$.nested[0:2].name", expected: "[$['nested'][0]['name'], $['nested'][1]['name']]"},
		{name: "nested array slice with step", path: "$.nested[0:3:2].age", expected: "[$['nested'][0]['age'], $['nested'][2]['age']]"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			res, err := Path([]byte(data), tt.path)
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
			result, err := parsePath(tt.path)
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
