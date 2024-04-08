package json

import (
	"errors"
	"strconv"
)

// LoadString loads a JSON string into the buffer.
func (b *buffer) LoadString(s string) error {
	if s == "" {
		return errors.New("empty JSON string")
	}

	b.data = []byte(s)
	b.length = len(b.data)
	b.index = 0
	b.last = GO
	b.state = GO

	return nil
}

// parseValue parses a JSON value from the buffer.
func (b *buffer) parseValue() (interface{}, error) {
	b.skipWhitespace()
	currToken, err := b.current()
	if err != nil {
		return nil, err
	}

	switch currToken {
	case doubleQuote:
		return b.parseString()
	case 't':
		if err := b.parseTrue(); err != nil {
			return nil, err
		}
		return true, nil
	case 'f':
		if err := b.parseFalse(); err != nil {
			return nil, err
		}
		return false, nil
	case 'n':
		if err := b.parseNull(); err != nil {
			return nil, err
		}
		return nil, nil
	default:
		if currToken == '-' || !notDigit(currToken) {
			return b.parseNumber()
		}

		return nil, errors.New("invalid token found while parsing JSON value")
	}
}

// parseString parses a string token from the buffer.
func (b *buffer) parseString() (string, error) {
	str, err := getString(b)
	if err != nil {
		return "", err
	}

	return *str, nil
}

// parseNumber parses a number token from the buffer.
func (b *buffer) parseNumber() (float64, error) {
	start := b.index

	if err := b.numeric(false); err != nil {
		return -1, err
	}

	numStr := string(b.data[start:b.index])
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return -1, err
	}

	return num, nil
}

// parseTrue parses a true literal token from the buffer.
func (b *buffer) parseTrue() error {
	if err := b.word(trueLiteral); err != nil {
		return errors.New("invalid token found while parsing boolean value")
	}

	return nil
}

// parseFalse parses a false literal token from the buffer.
func (b *buffer) parseFalse() error {
	if err := b.word(falseLiteral); err != nil {
		return errors.New("invalid token found while parsing boolean value")
	}

	return nil
}

// parseNull parses a null literal token from the buffer.
func (b *buffer) parseNull() error {
	if err := b.word(nullLiteral); err != nil {
		return errors.New("invalid token found while parsing null value")
	}

	return nil
}
