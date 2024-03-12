package json

import (
	"bytes"
)

const (
	indentStr = "  "
)

const indentGrowthFactor = 2

// Indent applies the indent to the input JSON data.
func Indent(data []byte) ([]byte, error) {
	var out bytes.Buffer
	var level int

	for i := 0; i < len(data); i++ {
		c := data[i] // current character

		switch c {
		case curlyOpen, bracketOpen:
			// check if the empty object or array
			// we don't need to apply the indent when it's empty containers.
			if i+1 < len(data) && (data[i+1] == '}' || data[i+1] == ']') {
				if err := out.WriteByte(c); err != nil {
					return nil, err
				}

				i++ // skip next character
				if err := out.WriteByte(data[i]); err != nil {
					return nil, err
				}
			} else {
				if err := out.WriteByte(c); err != nil {
					return nil, err
				}

				level++
				if err := writeNewlineAndIndent(&out, level); err != nil {
					return nil, err
				}
			}

		case curlyClose, bracketClose:
			level--
			if err := writeNewlineAndIndent(&out, level); err != nil {
				return nil, err
			}
			if err := out.WriteByte(c); err != nil {
				return nil, err
			}

		case comma:
			if err := out.WriteByte(c); err != nil {
				return nil, err
			}
			if err := writeNewlineAndIndent(&out, level); err != nil {
				return nil, err
			}

		case colon:
			if err := out.WriteByte(c); err != nil {
				return nil, err
			}
			if err := out.WriteByte(' '); err != nil { // add space after colon to separate key and value
				return nil, err
			}

		default:
			if err := out.WriteByte(c); err != nil {
				return nil, err
			}
		}
	}

	return out.Bytes(), nil
}

func writeNewlineAndIndent(out *bytes.Buffer, level int) error {
	if err := out.WriteByte('\n'); err != nil {
		return err
	}
	for i := 0; i < level; i++ {
		if _, err := out.WriteString(indentStr); err != nil {
			return err
		}
	}
	return nil
}