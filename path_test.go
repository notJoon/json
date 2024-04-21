package json

import (
	"testing"
)

func TestPath_Tokenize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		jsonPath string
		expected []PathToken
	}{
		{
			"$.store.book[*].author",
			[]PathToken{{"ROOT", "$"}, {"DOT", "."}, {"IDENTIFIER", "store"}, {"DOT", "."}, {"IDENTIFIER", "book"}, {"BRACKET", "[*]"}, {"DOT", "."}, {"IDENTIFIER", "author"}},
		},
		{
			"$['store']['book'][0]['title']",
			[]PathToken{{"ROOT", "$"}, {"BRACKET", "['store']"}, {"BRACKET", "['book']"}, {"BRACKET", "[0]"}, {"BRACKET", "['title']"}},
		},
		{
			"$[?(@.age >= 25)]",
			[]PathToken{{"ROOT", "$"}, {"BRACKET", "[?(@.age >= 25)]"}},
		},
		{
			"$.person.name",
			[]PathToken{{"ROOT", "$"}, {"DOT", "."}, {"IDENTIFIER", "person"}, {"DOT", "."}, {"IDENTIFIER", "name"}},
		},
		{
			"$['person']['name']",
			[]PathToken{{"ROOT", "$"}, {"BRACKET", "['person']"}, {"BRACKET", "['name']"}},
		},
	}

	for _, tt := range tests {
		tt := tt
		got, err := tokenize(tt.jsonPath)
		if err != nil {
			t.Errorf("tokenize(%q) resulted in an error: %v", tt.jsonPath, err)
		}
		if len(got) != len(tt.expected) {
			t.Errorf("tokenize(%q) = %v, expected %v", tt.jsonPath, got, tt.expected)
		}
		for i := range got {
			if got[i] != tt.expected[i] {
				t.Errorf("tokenize(%q)[%d] = {%q, %q}, expected {%q, %q}", tt.jsonPath, i, got[i].Type, got[i].Value, tt.expected[i].Type, tt.expected[i].Value)
			}
		}
	}
}

func TestClassify(t *testing.T) {
	t.Parallel()
    tests := []struct {
        tokens   []PathToken
        expected []ClassifiedToken
    }{
        {
            []PathToken{
                {Type: "ROOT", Value: "$"},
                {Type: "DOT", Value: "."},
                {Type: "IDENTIFIER", Value: "store"},
                {Type: "DOT", Value: "."},
                {Type: "IDENTIFIER", Value: "book"},
                {Type: "BRACKET", Value: "[*]"},
                {Type: "DOT", Value: "."},
                {Type: "IDENTIFIER", Value: "author"},
                {Type: "BRACKET", Value: "[?(@.price<20)]"},
                {Type: "BRACKET", Value: "[1:3]"},
                {Type: "BRACKET", Value: "[0]"},
                {Type: "STRING", Value: "'John Doe'"},
            },
            []ClassifiedToken{
                {PathToken{Type: "ROOT", Value: "$"}, "OPERATOR"},
                {PathToken{Type: "DOT", Value: "."}, "OPERATOR"},
                {PathToken{Type: "IDENTIFIER", Value: "store"}, "IDENTIFIER"},
                {PathToken{Type: "DOT", Value: "."}, "OPERATOR"},
                {PathToken{Type: "IDENTIFIER", Value: "book"}, "IDENTIFIER"},
                {PathToken{Type: "BRACKET", Value: "[*]"}, "ARRAY_WILDCARD"},
                {PathToken{Type: "DOT", Value: "."}, "OPERATOR"},
                {PathToken{Type: "IDENTIFIER", Value: "author"}, "IDENTIFIER"},
                {PathToken{Type: "BRACKET", Value: "[?(@.price<20)]"}, "FILTER"},
                {PathToken{Type: "BRACKET", Value: "[1:3]"}, "SLICE"},
                {PathToken{Type: "BRACKET", Value: "[0]"}, "ARRAY_INDEX_OR_KEY"},
                {PathToken{Type: "STRING", Value: "'John Doe'"}, "LITERAL"},
            },
        },
    }

    for _, tt := range tests {
		tt := tt
        got, err := classify(tt.tokens)
        if err != nil {
            t.Errorf("classify resulted in an error: %v", err)
        }
        if len(got) != len(tt.expected) {
            t.Errorf("classify = %v, expected %v", got, tt.expected)
        }
        for i := range got {
            if got[i] != tt.expected[i] {
                t.Errorf("classify[%d] = {%q, %q, %q}, expected {%q, %q, %q}", i, got[i].Type, got[i].Value, got[i].Class, tt.expected[i].Type, tt.expected[i].Value, tt.expected[i].Class)
            }
        }
    }
}