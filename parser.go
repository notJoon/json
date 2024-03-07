package json

import (
	"bytes"
	"errors"
	"strconv"
)

const (
	absMinInt64 = 1 << 63
	maxInt64    = absMinInt64 - 1
	maxUint64   = 1<<64 - 1
)

const unescapeStackBufSize = 64

// PaseStringLiteral parses a string from the given byte slice.
func ParseStringLiteral(data []byte) (string, error) {
	var buf [unescapeStackBufSize]byte

	bf, err := unescape(data, buf[:])
	if err != nil {
		return "", errors.New("invalid string input found while parsing string value")
	}

	return string(bf), nil
}

// ParseBoolLiteral parses a boolean value from the given byte slice.
func ParseBoolLiteral(data []byte) (bool, error) {
	switch {
	case bytes.Equal(data, trueLiteral):
		return true, nil
	case bytes.Equal(data, falseLiteral):
		return false, nil
	default:
		return false, errors.New("JSON Error: malformed boolean value found while parsing boolean value")
	}
}

// PaseFloatLiteral parses a float64 from the given byte slice.
//
// It utilizes double-precision (64-bit) floating-point format as defined
// by the IEEE 754 standard, providing a decimal precision of approximately 15 digits.
// TODO: support for 32-bit floating-point format
func ParseFloatLiteral(bytes []byte) (value float64, err error) {
	if len(bytes) == 0 {
		return -1, errors.New("JSON Error: empty byte slice found while parsing float value")
	}

	neg, bytes := trimNegativeSign(bytes)

	var exponentPart []byte
	for i, c := range bytes {
		if lower(c) == 'e' {
			exponentPart = bytes[i+1:]
			bytes = bytes[:i]
			break
		}
	}

	man, exp10, err := extractMantissaAndExp10(bytes)
	if err != nil {
		return -1, err
	}

	if len(exponentPart) > 0 {
		exp, err := strconv.Atoi(string(exponentPart))
		if err != nil {
			return -1, errors.New("JSON Error: invalid exponent value found while parsing float value")
		}
		exp10 += exp
	}

	// for fast float64 conversion
	f, success := eiselLemire64(man, exp10, neg)
	if !success {
		return 0, nil
	}

	return f, nil
}

func ParseIntLiteral(bytes []byte) (v int64, err error) {
	if len(bytes) == 0 {
		return 0, errors.New("JSON Error: empty byte slice found while parsing integer value")
	}

	neg, bytes := trimNegativeSign(bytes)

	var n uint64 = 0
	for _, c := range bytes {
		if notDigit(c) {
			return 0, errors.New("JSON Error: non-digit characters found while parsing integer value")
		}

		if n > maxUint64/10 {
			return 0, errors.New("JSON Error: numeric value exceeds the range limit")
		}

		n *= 10

		n1 := n + uint64(c-'0')
		if n1 < n {
			return 0, errors.New("JSON Error: numeric value exceeds the range limit")
		}

		n = n1
	}

	if n > maxInt64 {
		if neg && n == absMinInt64 {
			return -absMinInt64, nil
		}

		return 0, errors.New("JSON Error: numeric value exceeds the range limit")
	}

	if neg {
		return -int64(n), nil
	}

	return int64(n), nil
}

// extractMantissaAndExp10 parses a byte slice representing a decimal number and extracts the mantissa and the exponent of its base-10 representation.
// It iterates through the bytes, constructing the mantissa by treating each byte as a digit.
// If a decimal point is encountered, the function keeps track of the position of the decimal point to calculate the exponent.
// The function ensures that:
// - The number contains at most one decimal point.
// - All characters in the byte slice are digits or a single decimal point.
// - The resulting mantissa does not overflow a uint64.
func extractMantissaAndExp10(bytes []byte) (uint64, int, error) {
	var (
		man          uint64
		exp10        int
		decimalFound bool
	)

	for _, c := range bytes {
		if c == DotToken {
			if decimalFound {
				return 0, 0, errors.New("JSON Error: multiple decimal points found while parsing float value")
			}
			decimalFound = true
			continue
		}

		if notDigit(c) {
			return 0, 0, errors.New("JSON Error: non-digit characters found while parsing integer value")
		}

		digit := uint64(c - '0')

		if man > (maxUint64-digit)/10 {
			return 0, 0, errors.New("JSON Error: numeric value exceeds the range limit")
		}

		man = man*10 + digit

		if decimalFound {
			exp10--
		}
	}

	return man, exp10, nil
}

// typeOf check the type of the given value and returns the corresponding ValueType.
func typeOf(v interface{}) ValueType {
	switch v.(type) {
	case string:
		return String
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return Number
	case float32, float64:
		return Float
	case bool:
		return Boolean
	case nil:
		return Null
	case map[string]interface{}:
		return Object
	case []interface{}:
		return Array
	default:
		return Unknown
	}
}