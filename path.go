package json

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	errUnexpectedEOF    = errors.New("unexpected EOF")
	errUnexpectedChar   = errors.New("unexpected character")
	errStringNotClosed  = errors.New("string not closed")
	errBracketNotClosed = errors.New("bracket not closed")
)

// JSONPath returns the nodes that match the given JSON path.
//
// It parses the JSON path, unmarshals the JSON data, and processes each command in the path
// to traverse the JSON structure and retrieve the matching nodes.
//
// JSONPath expressions:
//
//	$	the root object/element
//	@	the current object/element
//	. or []	child operator
//	..	recursive descent. JSONPath borrows this syntax from E4X.
//	*	wildcard. All objects/elements regardless their names.
//	[]	subscript operator. XPath uses it to iterate over element collections and for predicates. In Javascript and JSON it is the native array operator.
//	[,]	Union operator in XPath results in a combination of node sets. JSONPath allows alternate names or array indices as a set.
//	[start:end:step]	array slice operator borrowed from ES4.
//	?()	applies a filter (script) expression.
//	()	script expression, using the underlying script engine.
func JSONPath(data []byte, path string) ([]*Node, error) {
	commands, err := ParsePath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path: %v", err)
	}

	root, err := Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	result := make([]*Node, 0)
	for i, cmd := range commands {
		if i == 0 && cmd == "$" {
			result = append(result, root)
			continue
		}

		result, err = processCommand(cmd, result)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// processCommand processes a single command on the given nodes.
//
// It determines the type of command and calls the corresponding function to handle the command.
func processCommand(cmd string, nodes []*Node) ([]*Node, error) {
	switch cmd {
	case "..":
		return processRecursiveDescent(nodes), nil
	case "*":
		return processWildcard(nodes), nil
	default:
		return processKeyUnion(cmd, nodes), nil
	}
}

// processRecursiveDescent performs a recursive descent on the given nodes.
//
// It recursively retrieves all the child nodes of each node in the given slice.
func processRecursiveDescent(nodes []*Node) []*Node {
	var result []*Node
	for _, node := range nodes {
		result = append(result, recursiveChildren(node)...)
	}
	return result
}

// processWildcard processes a wildcard command on the given nodes.
//
// It retrieves all the child nodes of each node in the given slice.
func processWildcard(nodes []*Node) []*Node {
	var result []*Node
	for _, node := range nodes {
		result = append(result, node.getSortedChildren()...)
	}
	return result
}

// processKeyUnion processes a key union command on the given nodes.
//
// It retrieves the child nodes of each node in the given slice that match any of the specified keys
func processKeyUnion(cmd string, nodes []*Node) []*Node {
	keys := strings.Split(cmd, ",")

	var result []*Node
	for _, node := range nodes {
		if !node.isContainer() {
			continue
		}

		for _, key := range keys {
			if child, ok := node.next[key]; ok {
				result = append(result, child)
			}
		}
	}
	return result
}

// recursiveChildren returns all the recursive child nodes of the given node that are containers.
//
// It recursively traverses the child nodes of the given node and their child nodes,
// and returns a slice of pointers to all the child nodes that are containers.
func recursiveChildren(node *Node) (result []*Node) {
	if node.isContainer() {
		for _, element := range node.getSortedChildren() {
			if element.isContainer() {
				result = append(result, element)
			}
		}
	}

	temp := make([]*Node, 0, len(result))
	temp = append(temp, result...)

	for _, element := range result {
		temp = append(temp, recursiveChildren(element)...)
	}

	return temp
}

var pathSegmentDelimiters = map[byte]bool{dot: true, bracketOpen: true}

// ParsePath parses the given path string and returns a slice of commands to be run.
//
// The function uses a state machine approach to parse the path based on the encountered tokens.
// It supports the following tokens and their corresponding states:
//   - Dollar sign ('$'): Appends a literal "$" to the result slice.
//   - Dot ('.'): Calls the processDot function to handle the dot token.
//   - Single dot ('.') followed by a child end character: Appends the substring between the dots to the result slice.
//   - Double dot ('..'): Appends ".." to the result slice.
//   - Opening bracket ('['): Calls the processBracketOpen function to handle the opening bracket token.
//   - Single quote (”') after the opening bracket: Calls the processSingleQuote function to handle the string within single quotes.
//   - Any other character after the opening bracket: Calls the processWithoutSingleQuote function to handle the string without single quotes.
//
// The function returns the slice of parsed commands and any error encountered during the parsing process.
// If an unexpected character is encountered, an error (errUnexpectedChar) is returned.
func ParsePath(path string) ([]string, error) {
	buf := newBuffer([]byte(path))
	result := make([]string, 0)

	for {
		b, err := buf.current()
		if err != nil {
			break
		}
		switch b {
		case dollarSign:
			result = append(result, "$")
		case dot:
			result, err = processDot(buf, result)
			if err != nil {
				return nil, err
			}
		case bracketOpen:
			result, err = processBracketOpen(buf, result)
			if err != nil {
				return nil, err
			}
		default:
			return nil, errUnexpectedChar
		}

		err = buf.step()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
	}

	return result, nil
}

// processDot handles the processing when a dot character is found in the buffer.
//
// It checks the next character in the buffer.
// If the next character is also a dot, it appends ".." to the result slice.
// Otherwise, it reads the characters until the next child end character (childEnd) and appends the substring to the result slice.
// It returns the updated result slice and any error encountered.
func processDot(buf *buffer, result []string) ([]string, error) {
	start := buf.index

	b, err := buf.next()
	if err == io.EOF {
		err = nil
		return result, nil
	}

	if err != nil {
		return nil, err
	}

	if b == dot {
		result = append(result, "..")
		buf.index--
		return result, nil
	}

	err = buf.skipAny(pathSegmentDelimiters)
	stop := buf.index
	if err == io.EOF {
		err = nil
		stop = buf.length
	} else {
		buf.index--
	}

	if err != nil {
		return nil, err
	}

	if start+1 < stop {
		result = append(result, string(buf.data[start+1:stop]))
	}

	return result, nil
}

// processBracketOpen handles the processing when an opening bracket character ('[') is found in the buffer.
//
// It reads the next character in the buffer and determines the appropriate processing based on the character:
//   - If the next character is a single quote (”'), it calls the processSingleQuote function to handle the string within single quotes.
//   - Otherwise, it calls the processWithoutSingleQuote function to handle the string without single quotes.
//
// It returns the updated result slice and any error encountered.
func processBracketOpen(buf *buffer, result []string) ([]string, error) {
	b, err := buf.next()
	if err != nil {
		return nil, errUnexpectedEOF
	}

	start := buf.index
	if b == singleQuote {
		result, err = processSingleQuote(buf, result, start)
	} else {
		result, err = processWithoutSingleQuote(buf, result, start)
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

// processSingleQuote handles the processing when a single quote character (”') is encountered after an opening bracket ('[') in the buffer.
//
// It assumes that the current position of the buffer is just after the single quote character.
//
// The function performs the following steps:
//  1. It skips the single quote character and reads the string until the next single quote character is found.
//  2. It checks if the character after the closing single quote is a closing bracket (']').
//     - If it is, the string between the single quotes is appended to the result slice.
//     - If it is not, an error (errBracketNotClosed) is returned.
//
// It returns the updated result slice and any error encountered.
func processSingleQuote(buf *buffer, result []string, start int) ([]string, error) {
	start++

	err := buf.string(singleQuote, true)
	if err != nil {
		return nil, errStringNotClosed
	}

	stop := buf.index

	b, err := buf.next()
	if err != nil {
		return nil, errUnexpectedEOF
	}

	if b != bracketClose {
		return nil, errBracketNotClosed
	}

	result = append(result, string(buf.data[start:stop]))

	return result, nil
}

// processWithoutSingleQuote handles the processing when a character other than
// a single quote (”') is encountered after an opening bracket ('[') in the buffer.
//
// It assumes that the current position of the buffer is just after the opening bracket.
//
// The function reads the characters until the next closing bracket (']') is found
// and appends the substring between the brackets to the result slice.
//
// It returns the updated result slice and any error encountered.
// If the closing bracket is not found, an error (errUnexpectedEOF) is returned.
func processWithoutSingleQuote(buf *buffer, result []string, start int) ([]string, error) {
	err := buf.skip(bracketClose)
	if err != nil {
		return nil, errUnexpectedEOF
	}

	stop := buf.index
	result = append(result, string(buf.data[start:stop]))

	return result, nil
}

// Paths returns calculated paths of underlying nodes
func Paths(array []*Node) []string {
	result := make([]string, 0, len(array))
	for _, element := range array {
		result = append(result, element.Path())
	}
	return result
}
