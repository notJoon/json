package json

import (
	"errors"
	"fmt"
	"io"
)

type buffer struct {
	data   []byte
	length int
	index  int

	last  States
	state States
	class Classes
}

// newBuffer creates a new buffer with the given data
func newBuffer(data []byte) *buffer {
	return &buffer{
		data:   data,
		length: len(data),
		last:   GO,
		state:  GO,
	}
}

// first retrieves the first non-whitespace (or other escaped) character in the buffer.
func (b *buffer) first() (byte, error) {
	for ; b.index < b.length; b.index++ {
		c := b.data[b.index]

		if !(c == whiteSpace || c == carriageReturn || c == newLine || c == tab) {
			return c, nil
		}
	}

	return 0, io.EOF
}

// current returns the byte of the current index.
func (b *buffer) current() (byte, error) {
	if b.index >= b.length {
		return 0, io.EOF
	}

	return b.data[b.index], nil
}

// next moves to the next byte and returns it.
func (b *buffer) next() (byte, error) {
	b.index++
	return b.current()
}

// step just moves to the next position.
func (b *buffer) step() error {
	_, err := b.next()
	return err
}

// move moves the index by the given position.
func (b *buffer) move(pos int) error {
	newIndex := b.index + pos

	if newIndex > b.length {
		return io.EOF
	}

	b.index = newIndex

	return nil
}

// slice returns the slice from the current index to the given position.
func (b *buffer) slice(pos int) ([]byte, error) {
	end := b.index + pos

	if end > b.length {
		return nil, io.EOF
	}

	return b.data[b.index:end], nil
}

// sliceFromIndices returns a slice of the buffer's data starting from 'start' up to (but not including) 'stop'.
func (b *buffer) sliceFromIndices(start, stop int) []byte {
	if start > b.length {
		start = b.length
	}

	if stop > b.length {
		stop = b.length
	}

	return b.data[start:stop]
}

// skip moves the index to skip the given byte.
func (b *buffer) skip(bs byte) error {
	for b.index < b.length {
		if b.data[b.index] == bs && !b.backslash() {
			return nil
		}

		b.index++
	}

	return io.EOF
}

// skipAny moves the index until it encounters one of the given set of bytes.
func (b *buffer) skipAny(endTokens map[byte]bool) error {
	for b.index < b.length {
		if _, exists := endTokens[b.data[b.index]]; exists {
			return nil
		}

		b.index++
	}

	return io.EOF
}

// skipAndReturnIndex moves the buffer index forward by one and returns the new index.
func (b *buffer) skipAndReturnIndex() (int, error) {
	err := b.step()
	if err != nil {
		return 0, err
	}

	return b.index, nil
}

// skipUntil moves the buffer index forward until it encounters a byte contained in the endTokens set.
func (b *buffer) skipUntil(endTokens map[byte]bool) (int, error) {
	for b.index < b.length {
		currentByte, err := b.current()
		if err != nil {
			return b.index, err
		}

		// Check if the current byte is in the set of end tokens.
		if _, exists := endTokens[currentByte]; exists {
			return b.index, nil
		}

		b.index++
	}

	return b.index, io.EOF
}

// significantTokens is a map where the keys are the significant characters in a JSON path.
// The values in the map are all true, which allows us to use the map as a set for quick lookups.
var significantTokens = map[byte]bool{
	dot:          true, // access properties of an object
	dollarSign:   true, // root object
	atSign:       true, // current object
	bracketOpen:  true, // start of an array index or filter expression
	bracketClose: true, // end of an array index or filter expression
}

// filterTokens stores the filter expression tokens.
var filterTokens = map[byte]bool{
	aesterisk: true, // wildcard
	andSign:   true,
	orSign:    true,
}

// skipToNextSignificantToken advances the buffer index to the next significant character.
// Significant characters are defined based on the JSON path syntax.
func (b *buffer) skipToNextSignificantToken() {
	for b.index < b.length {
		current := b.data[b.index]

		if _, ok := significantTokens[current]; ok {
			break
		}

		b.index++
	}
}

// backslash checks to see if the number of backslashes before the current index is odd.
//
// This is used to check if the current character is escaped. However, unlike the "unescape" function,
// "backslash" only serves to check the number of backslashes.
func (b *buffer) backslash() bool {
	if b.index == 0 {
		return false
	}

	count := 0
	for i := b.index - 1; ; i-- {
		if i >= b.length || b.data[i] != backSlash {
			break
		}

		count++

		if i == 0 {
			break
		}
	}

	return count%2 != 0
}

// numIndex holds a map of valid numeric characters
var numIndex = map[byte]bool{
	'0': true,
	'1': true,
	'2': true,
	'3': true,
	'4': true,
	'5': true,
	'6': true,
	'7': true,
	'8': true,
	'9': true,
	'.': true,
	'e': true,
	'E': true,
}

// pathToken checks if the current token is a valid JSON path token.
func (b *buffer) pathToken() error {
	var stack []byte

	inToken := false
	inNumber := false
	first := b.index

	for b.index < b.length {
		c := b.data[b.index]

		switch {
		case c == singleQuote:
			fallthrough

		case c == doubleQuote:
			inToken = true
			if err := b.step(); err != nil {
				return errors.New("error stepping through buffer")
			}

			if err := b.skip(c); err != nil {
				return errors.New("unmatched quote in path")
			}

			if b.index >= b.length {
				return errors.New("unmatched quote in path")
			}

		case c == bracketOpen || c == parenOpen:
			inToken = true
			stack = append(stack, c)

		case c == bracketClose || c == parenClose:
			inToken = true
			if len(stack) == 0 || (c == bracketClose && stack[len(stack)-1] != bracketOpen) || (c == parenClose && stack[len(stack)-1] != parenOpen) {
				return errors.New("mismatched bracket or parenthesis")
			}

			stack = stack[:len(stack)-1]

		case pathStateContainsValidPathToken(c):
			inToken = true

		case c == plus || c == minus:
			if inNumber || (b.index > 0 && numIndex[b.data[b.index-1]]) {
				inToken = true
			} else if !inToken && (b.index+1 < b.length && numIndex[b.data[b.index+1]]) {
				inToken = true
				inNumber = true
			} else if !inToken {
				return errors.New("unexpected operator at start of token")
			}

		default:
			if len(stack) != 0 || inToken {
				inToken = true
			} else {
				goto end
			}
		}

		b.index++
	}

end:
	if len(stack) != 0 {
		return errors.New("unclosed bracket or parenthesis at end of path")
	}

	if first == b.index {
		return errors.New("no token found")
	}

	if inNumber && !numIndex[b.data[b.index-1]] {
		inNumber = false
	}

	return nil
}

func pathStateContainsValidPathToken(c byte) bool {
	if _, ok := significantTokens[c]; ok {
		return true
	}

	if _, ok := filterTokens[c]; ok {
		return true
	}

	if _, ok := numIndex[c]; ok {
		return true
	}

	if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' {
		return true
	}

	return false
}

func (b *buffer) numeric(token bool) error {
	if token {
		b.last = GO
	}

	for ; b.index < b.length; b.index++ {
		b.class = b.getClasses(doubleQuote)
		if b.class == __ {
			return errors.New("invalid token found while parsing path")
		}

		b.state = StateTransitionTable[b.last][b.class]
		if b.state == __ {
			if token {
				break
			}

			return errors.New("invalid token found while parsing path")
		}

		if b.state < __ {
			return nil
		}

		if b.state < MI || b.state > E3 {
			return nil
		}

		b.last = b.state
	}

	if b.last != ZE && b.last != IN && b.last != FR && b.last != E3 {
		return errors.New("invalid token found while parsing path")
	}

	return nil
}

func (b *buffer) getClasses(c byte) Classes {
	if b.data[b.index] >= 128 {
		return C_ETC
	}

	if c == singleQuote {
		return QuoteAsciiClasses[b.data[b.index]]
	}

	return AsciiClasses[b.data[b.index]]
}

func (b *buffer) getState() States {
	b.last = b.state

	b.class = b.getClasses(doubleQuote)
	if b.class == __ {
		return __
	}

	b.state = StateTransitionTable[b.last][b.class]

	return b.state
}

// string parses a string token from the buffer.
//
// The function takes a byte 'search' and a boolean 'token'.
// 'search' is the byte that the function is looking for in the buffer.
// If 'token' is set to true, the buffer will parse the string as a token.
// Otherwise, it will parse the string as a value.
//
// The function iterates over the buffer until it finds the 'search' byte or reaches the end of the buffer.
// For each byte, it determines its class using the 'getClasses' method and updates the 'class' property of the buffer.
//
// It then uses the 'StateTransitionTable' to determine the next state based on the current 'last' state and the 'class' of the current byte.
// If an invalid state is encountered, it returns an error.
//
// If a valid state less than '__' is encountered, it breaks the loop and updates the 'last' state of the buffer.
//
// The function returns nil if it successfully parses the string.
func (b *buffer) string(search byte, token bool) error {
	if token {
		b.last = GO
	}

	for ; b.index < b.length; b.index++ {
		b.class = b.getClasses(search)

		if b.class == __ {
			return errors.New("invalid token found while parsing path")
		}

		b.state = StateTransitionTable[b.last][b.class]
		if b.state == __ {
			return errors.New("invalid token found while parsing path")
		}

		if b.state < __ {
			break
		}

		b.last = b.state
	}

	return nil
}

// word is a method on the buffer struct that checks if the next sequence of bytes in the buffer matches the provided byte slice 'bs'.
//
// The function iterates over the buffer from the current index until the end of the buffer or until it has checked all bytes in 'bs'.
// For each byte in the buffer, it checks if it matches the corresponding byte in 'bs'.
// If it encounters a byte that does not match, it returns an error indicating an invalid token was found while parsing the path.
//
// If it successfully checks all bytes in 'bs' without encountering a mismatch, it breaks the loop.
// After the loop, it checks if it has checked all bytes in 'bs'. If not, it returns an error indicating an invalid token was found while parsing the path.
//
// If it has checked all bytes in 'bs' and all of them matched, it returns nil indicating successful parsing.
func (b *buffer) word(bs []byte) error {
	var c byte

	max := len(bs)
	index := 0

	for ; b.index < b.length; b.index++ {
		c = b.data[b.index]

		if c != bs[index] {
			return errors.New("invalid token found while parsing path")
		}

		index++
		if index >= max {
			break
		}
	}

	if index != max {
		return errors.New("invalid token found while parsing path")
	}

	return nil
}

func numberKind2f64(value interface{}) (result float64, err error) {
	switch typed := value.(type) {
	case float64:
		result = typed
	case float32:
		result = float64(typed)
	case int:
		result = float64(typed)
	case int8:
		result = float64(typed)
	case int16:
		result = float64(typed)
	case int32:
		result = float64(typed)
	case int64:
		result = float64(typed)
	case uint:
		result = float64(typed)
	case uint8:
		result = float64(typed)
	case uint16:
		result = float64(typed)
	case uint32:
		result = float64(typed)
	case uint64:
		result = float64(typed)
	default:
		err = fmt.Errorf("invalid number type: %T", value)
	}

	return
}
