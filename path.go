package json

import (
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

var (
	errUnexpectedEOF             = errors.New("unexpected EOF")
	errUnexpectedChar            = errors.New("unexpected character")
	errStringNotClosed           = errors.New("string not closed")
	errBracketNotClosed          = errors.New("bracket not closed")
	errInvalidSlicePathSyntax    = errors.New("invalid slice path syntax")
	errInvalidSliceFromValue     = errors.New("invalid slice from value")
	errInvalidSliceToValue       = errors.New("invalid slice to value")
	errInvalidSliceStepValue     = errors.New("invalid slice step value")
	errParenthesisMismatchInPath = errors.New("parenthesis mismatch in path")
)

// Path returns the nodes that match the given JSON path.
//
// It parses the JSON path, unmarshals the JSON data, and processes each command in the path
// to traverse the JSON structure and retrieve the matching nodes.
//
// Path expressions:
//
//	$	the root object/element
//	@	the current object/element
//	. or []	child operator
//	..	recursive descent. Path borrows this syntax from E4X.
//	*	wildcard. All objects/elements regardless their names.
//	[]	subscript operator. XPath uses it to iterate over element collections and for predicates. In Javascript and JSON it is the native array operator.
//	[,]	Union operator in XPath results in a combination of node sets. Path allows alternate names or array indices as a set.
//	[start:end:step]	array slice operator borrowed from ES4.
//	?()	applies a filter (script) expression.
//	()	script expression, using the underlying script engine.
func Path(data []byte, path string) ([]*Node, error) {
	commands, err := parsePath(path)
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
	switch {
	case cmd == "..":
		return processRecursiveDescent(nodes), nil
	case cmd == "*":
		return processWildcard(nodes), nil
	case strings.Contains(cmd, ":"):
		return processSlice(cmd, nodes)
	case strings.HasPrefix(cmd, "?(") && strings.HasSuffix(cmd, ")"):
		panic("not implemented eval")
	default:
		// return processKeyUnion(cmd, nodes), nil
		res, err := processKeyUnion(cmd, nodes)
		if err != nil {
			return nil, err
		}
		return res, nil
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
func processKeyUnion(cmd string, nodes []*Node) ([]*Node, error) {
	buf := newBuffer([]byte(cmd))
	keys := make([]string, 0)

	for {
		c, err := buf.first()
		if err != nil {
			return nil, err
		}

		if c == comma {
			return nil, errUnexpectedChar
		}

		from := buf.index
		err = buf.pathToken()
		if err != nil && err != io.EOF {
			return nil, err
		}

		key := string(buf.data[from:buf.index])
		if len(key) > 2 && key[0] == singleQuote && key[len(key)-1] == singleQuote {
			key = key[1 : len(key)-1]
		}

		keys = append(keys, key)
		c, err = buf.first()
		if err != nil {
			err = nil
			break
		}

		if c != comma {
			return nil, errUnexpectedChar
		}

		err = buf.step()
		if err != nil {
			return nil, err
		}
	}

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
	return result, nil
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

// parsePath parses the given path string and returns a slice of commands to be run.
//
// The function uses a state machine approach to parse the path based on the encountered tokens.
// It supports the following tokens and their corresponding states:
//   - Dollar sign ('$'): Appends a literal "$" to the result slice.
//   - Dot ('.'): Calls the processDot function to handle the dot token.
//   - Single dot ('.') followed by a child end character: Appends the substring between the dots to the result slice.
//   - Double dot ('..'): Appends ".." to the result slice.
//   - Opening bracket ('['): Calls the processBracketOpen function to handle the opening bracket token.
//   - Single quote ('”') after the opening bracket: Calls the processSingleQuote function to handle the string within single quotes.
//   - Any other character after the opening bracket: Calls the processWithoutSingleQuote function to handle the string without single quotes.
//
// The function returns the slice of parsed commands and any error encountered during the parsing process.
// If an unexpected character is encountered, an error (errUnexpectedChar) is returned.
func parsePath(path string) ([]string, error) {
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
//   - If the next character is a single quote (`'`), it calls the processSingleQuote function to handle the string within single quotes.
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

// processSingleQuote handles the processing when a single quote character (`'`) is encountered after an opening bracket ('[') in the buffer.
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
// a single quote (`'`) is encountered after an opening bracket ('[') in the buffer.
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

// processSlice processes a slice path on the given nodes.
//
// The slice path has the following syntax:
//
//		[start:end:step]
//
//	  - start: The starting index of the slice (inclusive). if omitted, it defaults to 0.
//	    If negative, it counts from the end of the array.
//	  - end: The ending index of the slice (exclusive). if omitted, it defaults to the length of the array.
//	    If negative, it counts from the end of the array.
//	  - step: The step value for the slice. if omitted, it defaults to 1.
//
// The function performs the following steps:
//
// 1. Split the slice path into start, end, and step values.
//
// 2. Parses the each syntax components as integers.
//
// 3. For each node in the given nodes:
//   - If the node is an array:
//   - Calculate the length of the array.
//   - Adjust the start and end values if they are negative.
//   - Check if the slice range is within the bounds of the array.
//   - Iterate over the array elements based on the start, end and step values.
//   - Append the selected elements to the result slice.
//
// It returns the slice of selected nodes and any error encountered during the parsing process.
func processSlice(cmd string, nodes []*Node) ([]*Node, error) {
	from, to, step, err := parseSliceParams(cmd)
	if err != nil {
		return nil, err
	}

	var result []*Node
	for _, node := range nodes {
		if node.IsArray() {
			result = append(result, selectArrayElement(node, from, to, step)...)
		}
	}

	return result, nil
}

// parseSliceParams parses the slice parameters from the given path command.
func parseSliceParams(cmd string) (int64, int64, int64, error) {
	keys := strings.Split(cmd, ":")
	ks := len(keys)
	if ks > 3 {
		return 0, 0, 0, errInvalidSlicePathSyntax
	}

	from, err := strconv.ParseInt(keys[0], 10, 64)
	if err != nil {
		return 0, 0, 0, errInvalidSliceFromValue
	}

	to := int64(0)
	if ks > 1 {
		to, err = strconv.ParseInt(keys[1], 10, 64)
		if err != nil {
			return 0, 0, 0, errInvalidSliceToValue
		}
	}

	step := int64(1)
	if ks == 3 {
		step, err = strconv.ParseInt(keys[2], 10, 64)
		if err != nil {
			return 0, 0, 0, errInvalidSliceStepValue
		}
	}

	return from, to, step, nil
}

// selectArrayElement selects the array elements based on the given from, to, and step values.
func selectArrayElement(node *Node, from, to, step int64) []*Node {
	length := int64(len(node.next))

	if to == 0 {
		to = length
	}

	if from < 0 {
		from += length
	}

	if to < 0 {
		to += length
	}

	from = int64(math.Max(0, math.Min(float64(from), float64(length))))
	to = int64(math.Max(0, math.Min(float64(to), float64(length))))

	if step <= 0 || from >= to {
		return nil
	}

	// This formula calculates the number of elements that will be selected based on the given
	// from, to, and step values. It ensures that the correct number of elements are allocated
	// in the result slice.
	size := (to - from + step - 1) / step
	result := make([]*Node, 0, size)

	for i := from; i < to; i += step {
		if child, ok := node.next[fmt.Sprintf("%d", i)]; ok {
			result = append(result, child)
		}
	}

	return result
}

// parseFilterExpr removes the filter prefix (`?(`) and the filter suffix (`)`) from the given command.
// After removing the prefix and suffix, it splits the command into tokens based on the comma (`,`) character.
func parseFilterExpression(cmd string) ([]string, error) {
	expr := strings.TrimPrefix(cmd, "?(")
	expr = strings.TrimSuffix(expr, ")")

	tokens, err := tokenizeExpression(expr)
	if err != nil {
		return nil, err
	}

	rpnTokens, err := convertToRPN(tokens)
	if err != nil {
		return nil, err
	}

	return rpnTokens, nil
}

// tokenizeExpression tokenize the given expression string into a slice of tokens.
// it handles quoted strings, parenthesis, commas, and variable references starting with '@' symbol.
//
// The function iterates through each character of the expression string and builds the tokens accordingly.
// It uses a strings.Builder to efficiently concatenate characters into tokens.
//
// Rules for tokenization:
//   - Quoted strings (enclosed in single quotes ') are treated as a single token.
//   - Parentheses '(' and ')' and commas ',' are treated as separate tokens.
//   - Variable references starting with '@' followed by '.' and alphanumeric characters or underscore are treated as a single token.
//   - Whitespace characters are used as token separators.
//
// Returns a slice of strings representing the tokens of the expression.
// Returns an error if the expression is invalid (e.g., unmatched quotes).
func tokenizeExpression(expr string) ([]string, error) {
	var (
		tokens   []string
		token    strings.Builder
		inQuotes bool
	)

	for i := 0; i < len(expr); i++ {
		char := expr[i]

		if char == '\'' {
			if inQuotes {
				token.WriteByte(char)
				tokens = append(tokens, token.String())
				token.Reset()
				inQuotes = false
			} else {
				if token.Len() > 0 {
					tokens = append(tokens, token.String())
					token.Reset()
				}
				token.WriteByte(char)
				inQuotes = true
			}
		} else if inQuotes {
			token.WriteByte(char)
		} else if char == '(' || char == ')' || char == ',' {
			if token.Len() > 0 {
				tokens = append(tokens, token.String())
				token.Reset()
			}
			tokens = append(tokens, string(char))
		} else if char == ' ' {
			if token.Len() > 0 {
				tokens = append(tokens, token.String())
				token.Reset()
			}
		} else if char == '@' {
			if token.Len() > 0 {
				tokens = append(tokens, token.String())
				token.Reset()
			}
			token.WriteByte(char)
			if i+1 < len(expr) && expr[i+1] == '.' {
				i++
				for i < len(expr) && (isAlphaNumeric(expr[i]) || expr[i] == '_') {
					token.WriteByte(expr[i])
					i++
				}
				i--
			}
		} else {
			token.WriteByte(char)
		}
	}

	if token.Len() > 0 {
		tokens = append(tokens, token.String())
	}

	return tokens, nil
}

func isAlphaNumeric(char byte) bool {
	ch := rune(char)
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_'
}

// convertToRPN converts an infix expression to Reverse Polish Notation (RPN).
// It takes a slice of strings representing the path tokens -- especially the filter expression --
// of the infix expression and returns the RPN format of the expression.
//
// This RPN format will make it easier to evaluate the expression.
func convertToRPN(tokens []string) ([]string, error) {
	var output, stack []string

	for _, token := range tokens {
		if isOperand(token) {
			output = append(output, token)
		} else if isOperator(token) {
			for len(stack) > 0 && precedence(token) <= precedence(stack[len(stack)-1]) {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token)
		} else if token == "(" {
			stack = append(stack, token)
		} else if token == ")" {
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				return nil, errParenthesisMismatchInPath
			}
			stack = stack[:len(stack)-1] // Pop "("
		} else {
			return nil, fmt.Errorf("unknown token: %s", token)
		}
	}

	for len(stack) > 0 {
		if stack[len(stack)-1] == "(" {
			return nil, errParenthesisMismatchInPath
		}
		output = append(output, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return output, nil
}

func isOperand(token string) bool {
	return !isOperator(token) && token != "(" && token != ")"
}

func isOperator(token string) bool {
	return token == "==" || token == "!=" || token == ">" || token == ">=" || token == "<" || token == "<=" || token == "&&" || token == "||"
}

func precedence(operator string) int {
	switch operator {
	case "||":
		return 1
	case "&&":
		return 2
	case "==", "!=":
		return 3
	case ">", ">=", "<", "<=":
		return 4
	default:
		return 0
	}
}

// prototype of eval function
// TODO: make it work on Node type.
func eval(rpn []string, ctx interface{}) (interface{}, error) {
	stack := make([]interface{}, 0)

	for _, token := range rpn {
		if isOperand(token) {
			op, err := resolveOperand(token, ctx)
			if err != nil {
				return nil, err
			}
			stack = append(stack, op)
		} else if isOperator(token) {
			if len(stack) < 2 {
				return nil, fmt.Errorf("insufficient operands for operator: %s", token)
			}

			rsh := stack[len(stack)-1]
			lsh := stack[len(stack)-2]
			stack := stack[:len(stack)-2]

			result, err := applyOperator(token, lsh, rsh)
			if err != nil {
				return nil, err
			}

			stack = append(stack, result)
		} else {
			return nil, fmt.Errorf("unknown token: %s", token)
		}
	}

	if len(stack) != 1 {
		return nil, errors.New("invalid expression")
	}

	return stack[0], nil
}

// resolveOperand resolves the value of an operand in the RPN expression.
// It takes a string representing the operand and an interface{} value representing the context.
//
// If the operand starts with '@', it is treated as a variable reference and the corresponding value is retrieved from the context.
// If the operand is enclosed in single quotes, it is treated as a string literal and the quotes are stripped.
// Otherwise, the operand is parsed as a literal value (e.g., integer, float, boolean).
//
// Returns the resolved value of the operand.
// Returns an error if the variable reference is invalid or the operand cannot be parsed.
//! Demo version
func resolveOperand(operand string, ctx interface{}) (interface{}, error) {
	if strings.HasPrefix(operand, "@") {
		path := strings.Split(operand[1:], ".")
		val, err := getValueFromContext(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("invalid variable reference: %s", operand)
		}

		return val, nil
	}

	if strings.HasPrefix(operand, "'") && strings.HasSuffix(operand, "'") {
		return operand[1 : len(operand)-1], nil
	}

	if operand == "true" {
		return true, nil
	} else if operand == "false" {
		return false, nil
	}

	if i, err := strconv.Atoi(operand); err == nil {
		return i, nil
	}

	// if f, err := ParseFloatLiteral([]byte(operand)); err == nil {
	// 	return f, nil
	// }

	if f, err := strconv.ParseFloat(operand, 64); err == nil {
		return f, nil
	}

	return nil, fmt.Errorf("invalid operand: %s", operand)
}

// getValueFromContext retrieves the value from the context based on the given path.
func getValueFromContext(context interface{}, path []string) (interface{}, error) {
    var value interface{} = context
    for _, key := range path {
        if m, ok := value.(map[string]interface{}); ok {
            value, ok = m[key]
            if !ok {
                return nil, fmt.Errorf("key not found: %s", key)
            }
        } else {
            return nil, fmt.Errorf("invalid context type")
        }
    }
    return value, nil
}

// applyOperator applies the given operator to the operands and returns the result.
// It takes a string representing the operator and two interface{} values representing the operands.
//
// Supported operators:
// - "==": Equality comparison
// - "!=": Inequality comparison
// - ">": Greater than comparison
// - ">=": Greater than or equal to comparison
// - "<": Less than comparison
// - "<=": Less than or equal to comparison
// - "&&": Logical AND
// - "||": Logical OR
//
// Returns the result of the operator application.
// Returns an error if the operator is unsupported or the operands are of incompatible types.
func applyOperator(operator string, left, right interface{}) (interface{}, error) {
    switch operator {
    case "==":
        return applyEqualityOperator(left, right)
    case "!=":
        result, err := applyEqualityOperator(left, right)
        if err != nil {
            return nil, err
        }
        return !result.(bool), nil
    case ">":
        return applyComparisonOperator(left, right, '>')
    case ">=":
        return applyComparisonOperator(left, right, '>')
    case "<":
        return applyComparisonOperator(left, right, '<')
    case "<=":
        return applyComparisonOperator(left, right, '<')
    case "&&":
        return applyLogicalOperator(left, right, '&')
    case "||":
        return applyLogicalOperator(left, right, '|')
    default:
        return nil, fmt.Errorf("unsupported operator: %s", operator)
    }
}

// applyEqualityOperator applies the equality operator to the operands.
// TODO: remove reflect. use type assertion instead.
func applyEqualityOperator(left, right interface{}) (interface{}, error) {
    leftValue := reflect.ValueOf(left)
    rightValue := reflect.ValueOf(right)

    if leftValue.Type() != rightValue.Type() {
        return false, nil
    }

    switch leftValue.Kind() {
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
        return leftValue.Int() == rightValue.Int(), nil
    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
        return leftValue.Uint() == rightValue.Uint(), nil
    case reflect.Float32, reflect.Float64:
        return leftValue.Float() == rightValue.Float(), nil
    case reflect.String:
        return leftValue.String() == rightValue.String(), nil
    case reflect.Bool:
        return leftValue.Bool() == rightValue.Bool(), nil
    default:
        return nil, fmt.Errorf("unsupported operand type for equality: %v", leftValue.Type())
    }
}

// applyComparisonOperator applies the comparison operator to the operands.
func applyComparisonOperator(left, right interface{}, operator rune) (interface{}, error) {
    leftValue := reflect.ValueOf(left)
    rightValue := reflect.ValueOf(right)

    switch leftValue.Kind() {
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
        switch operator {
        case '>':
            return leftValue.Int() > rightValue.Int(), nil
        case '<':
            return leftValue.Int() < rightValue.Int(), nil
        }
    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
        switch operator {
        case '>':
            return leftValue.Uint() > rightValue.Uint(), nil
        case '<':
            return leftValue.Uint() < rightValue.Uint(), nil
        }
    case reflect.Float32, reflect.Float64:
        switch operator {
        case '>':
            return leftValue.Float() > rightValue.Float(), nil
        case '<':
            return leftValue.Float() < rightValue.Float(), nil
        }
    case reflect.String:
        switch operator {
        case '>':
            return leftValue.String() > rightValue.String(), nil
        case '<':
            return leftValue.String() < rightValue.String(), nil
        }
    default:
        return nil, fmt.Errorf("unsupported operand type for comparison: %v", leftValue.Type())
    }

    return nil, fmt.Errorf("unsupported comparison operator: %c", operator)
}

// applyLogicalOperator applies the logical operator to the operands.
func applyLogicalOperator(left, right interface{}, operator rune) (interface{}, error) {
    leftValue := reflect.ValueOf(left)
    rightValue := reflect.ValueOf(right)

    if leftValue.Kind() != reflect.Bool || rightValue.Kind() != reflect.Bool {
        return nil, fmt.Errorf("operands must be boolean for logical operator")
    }

    switch operator {
    case '&':
        return leftValue.Bool() && rightValue.Bool(), nil
    case '|':
        return leftValue.Bool() || rightValue.Bool(), nil
    default:
        return nil, fmt.Errorf("unsupported logical operator: %c", operator)
    }
}
