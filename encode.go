package json

import (
	// "bytes"
	"fmt"
	"strconv"
)

// TODO: build bytes with bytes.Buffer
func Marshal(node *Node) ([]byte, error) {
	result := make([]byte, 0)
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
			result = append(result, nullLiteral...)

		case Number:
			nVal, err = node.GetNumeric()
			if err != nil {
				return nil, err
			}

			num := fmt.Sprintf("%g", nVal)
			result = append(result, num...)

		case String:
			sVal, err = node.GetString()
			if err != nil {
				return nil, err
			}

			quoted := fmt.Sprint(strconv.Quote(sVal))
			result = append(result, quoted...)

		case Boolean:
			bVal, err = node.GetBool()
			if err != nil {
				return nil, err
			}

			bStr := fmt.Sprintf("%t", bVal)
			result = append(result, bStr...)

		case Array:
			result = append(result, SquareOpenToken)

			for i := 0; i < len(node.next); i++ {
				if i != 0 {
					result = append(result, CommaToken)
				}

				elem, ok := node.next[strconv.Itoa(i)]
				if !ok {
					return nil, fmt.Errorf("array element %d is not found", i)
				}

				oVal, err = Marshal(elem)
				if err != nil {
					return nil, err
				}

				result = append(result, oVal...)
			}

			result = append(result, SquareCloseToken)

		case Object:
			result = append(result, CurlyOpenToken)

			bVal = false
			for k, v := range node.next {
				if bVal {
					result = append(result, CommaToken)
				} else {
					bVal = true
				}

				key := fmt.Sprint(strconv.Quote(k))
				result = append(result, key...)
				result = append(result, ColonToken)

				oVal, err = Marshal(v)
				if err != nil {
					return nil, err
				}

				result = append(result, oVal...)
			}

			result = append(result, CurlyCloseToken)
		}
	} else if node.ready() {
		return nil, fmt.Errorf("node is not ready")
	} else {
		return nil, fmt.Errorf("node is not modified")
	}

	return result, nil
}
