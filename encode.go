package json

import (
	"bytes"
	"fmt"
	"strconv"
)

func Marshal(node *Node) ([]byte, error) {
	var buf bytes.Buffer
	var (
		sVal string
		bVal bool
		nVal float64
		oVal []byte
		err  error
	)

	if node == nil {
		return nil, fmt.Errorf("node is nil")
	} else if node.modified {
		switch node.nodeType {
		case Null:
			buf.Write(nullLiteral)

		case Number:
			nVal, err = node.GetNumeric()
			if err != nil {
				return nil, err
			}

			num := fmt.Sprintf("%g", nVal)
			buf.WriteString(num)

		case String:
			sVal, err = node.GetString()
			if err != nil {
				return nil, err
			}

			quoted := fmt.Sprint(strconv.Quote(sVal))
			buf.WriteString(quoted)

		case Boolean:
			bVal, err = node.GetBool()
			if err != nil {
				return nil, err
			}

			bStr := fmt.Sprintf("%t", bVal)
			buf.WriteString(bStr)

		case Array:
			buf.WriteByte(bracketOpen)

			for i := 0; i < len(node.next); i++ {
				if i != 0 {
					buf.WriteByte(comma)
				}

				elem, ok := node.next[strconv.Itoa(i)]
				if !ok {
					return nil, fmt.Errorf("array element %d is not found", i)
				}

				oVal, err = Marshal(elem)
				if err != nil {
					return nil, err
				}

				buf.Write(oVal)
			}

			buf.WriteByte(bracketClose)

		case Object:
			buf.WriteByte(curlyOpen)

			bVal = false
			for k, v := range node.next {
				if bVal {
					buf.WriteByte(comma)
				} else {
					bVal = true
				}

				key := fmt.Sprint(strconv.Quote(k))
				buf.WriteString(key)
				buf.WriteByte(colon)

				oVal, err = Marshal(v)
				if err != nil {
					return nil, err
				}

				buf.Write(oVal)
			}

			buf.WriteByte(curlyClose)
		}
	} else if node.ready() {
		buf.Write(node.Source())
	} else {
		return nil, fmt.Errorf("node is not modified")
	}

	return buf.Bytes(), nil
}
