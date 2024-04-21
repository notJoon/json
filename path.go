package json

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// PathToken represents a token in a JSON path.
type PathToken struct {
	Type  string	// Type is the type of the token.
	Value string	// Value is the value of the token.
}

// Tokenize breaks a JSON path into tokens.
func tokenize(path string) ([]PathToken, error) {
	var tokens []PathToken
	i := 0
	n := len(path)

	for i < n {
		switch {
		case path[i] == '$':
			tokens = append(tokens, PathToken{"ROOT", "$"})
			i++

		case path[i] == '.':
			tokens = append(tokens, PathToken{"DOT", "."})
			i++

		case path[i] == '*':
			tokens = append(tokens, PathToken{"WILDCARD", "*"})
			i++

		case path[i] == '[':
			j := i
			if j < n && path[j] == '[' {
				for j < n && path[j] != ']' {
					j++
				}
				if j < n && path[j] == ']' {
					j++ // Include the closing bracket
					tokens = append(tokens, PathToken{"BRACKET", path[i:j]})
					i = j
				} else {
					return nil, fmt.Errorf("mismatched brackets")
				}
			}

		case path[i] == '"' || path[i] == '\'':
			delimiter := path[i]
			i++
			start := i
			for i < n && path[i] != delimiter {
				if path[i] == '\\' { // Handle escaped quotes
					i++
				}
				i++
			}
			if i < n && path[i] == delimiter {
				tokens = append(tokens, PathToken{"STRING", path[start:i]})
				i++
			} else {
				return nil, fmt.Errorf("unterminated string")
			}

		case path[i] == '?' || path[i] == '(' || path[i] == ')':
			tokens = append(tokens, PathToken{string(path[i]), string(path[i])})
			i++

		case unicode.IsDigit(rune(path[i])) || path[i] == '-':
			start := i
			i++
			for i < n && unicode.IsDigit(rune(path[i])) {
				i++
			}
			tokens = append(tokens, PathToken{"NUMBER", path[start:i]})

		default:
			if regexp.MustCompile(`[a-zA-Z_]+`).MatchString(string(path[i])) {
				start := i
				for i < n && regexp.MustCompile(`[a-zA-Z0-9_]+`).MatchString(string(path[i])) {
					i++
				}
				tokens = append(tokens, PathToken{"IDENTIFIER", path[start:i]})
			} else {
				i++
			}
		}
	}

	return tokens, nil
}

// ClassifiedToken represents a token in a JSON path that has been classified.
type ClassifiedToken struct {
	PathToken
	Class string	// Class is the classification of the token.
}

// Classify classifies tokens in a JSON path.
func classify(tokens []PathToken) ([]ClassifiedToken, error) {
	var classified []ClassifiedToken
	for _, token := range tokens {
		switch token.Type {
		case "ROOT", "DOT", "WILDCARD":
			classified = append(classified, ClassifiedToken{token, "OPERATOR"})
		case "BRACKET":
			if strings.HasPrefix(token.Value, "[*]") {
				classified = append(classified, ClassifiedToken{PathToken: token, Class: "ARRAY_WILDCARD"})
			} else if strings.Contains(token.Value, "?") {
				classified = append(classified, ClassifiedToken{PathToken: token, Class: "FILTER"})
			} else if strings.ContainsAny(token.Value, ":") {
				classified = append(classified, ClassifiedToken{PathToken: token, Class: "SLICE"})
			} else {
				classified = append(classified, ClassifiedToken{PathToken: token, Class: "ARRAY_INDEX_OR_KEY"})
			}
		case "STRING", "NUMBER":
			classified = append(classified, ClassifiedToken{PathToken: token, Class: "LITERAL"})
		case "IDENTIFIER":
			classified = append(classified, ClassifiedToken{PathToken: token, Class: "IDENTIFIER"})
		default:
			return nil, fmt.Errorf("unclassified token type: %s", token.Type)
		}
	}

	return classified, nil
}

type PathNode interface {
	Eval(ctx interface{}) interface{}
}

// IdentifierNode handles property access on an object.
type IdentifierNode struct {
	Key string
}

func (node *IdentifierNode) Eval(ctx interface{}) interface{} {
	if m, ok := ctx.(map[string]interface{}); ok {
		return m[node.Key]
	}
	return nil
}

// PathRootNode starts the JSON path
type PathRootNode struct {
	children []PathNode
}

func NewPathRootNode() *PathRootNode {
	return &PathRootNode{}
}

func (node PathRootNode) Eval(ctx interface{}) interface{} {
	for _, child := range node.children {
		ctx = child.Eval(ctx)
	}
	return ctx
}

func parseJSONPath(path string) (PathNode, error) {
	tt, err := tokenize(path)
	if err != nil {
		return nil, err
	}
	classified, err := classify(tt)
	if err != nil {
		return nil, err
	}

	rootNode := NewPathRootNode()
	currNode := rootNode

	for _, tok := range classified {
		switch tok.Class {
		case "IDENTIFIER":
			currNode.children = append(currNode.children, &IdentifierNode{tok.Value})
		}
	}

	return rootNode, nil
}