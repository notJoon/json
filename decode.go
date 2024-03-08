package json

import (
	"fmt"
	"io"
)

// State machine unmarshal taken from: https://github.com/spyzhov/ajson/blob/master/internal/state.go
func Unmarshal(data []byte) (*Node, error) {
	buf := newBuffer(data)

	var (
		state   States
		key     *string
		current *Node
		useKey  = func() **string {
			tmp := cptrs(key)
			key = nil
			return &tmp
		}
	)

	_, err := buf.first()
	if err != nil {
		return nil, io.EOF
	}

	for {
		state = buf.getState()
		if state == __ {
			return nil, unexpectedTokenError(buf.data, buf.index)
		}

		// region state machine
		if state >= GO {
			switch buf.state {
			case ST: // string
				if current != nil && current.IsObject() && key == nil {
					// key detected
					key, err = getString(buf)
					buf.state = CO
				} else {
					// string detected
					current, err = NewNode(current, buf, String, useKey())
					if err != nil {
						break
					}

					err = buf.string(DoublyQuoteToken, false)

					current.borders[1] = buf.index + 1
					buf.state = OK

					if current.prev != nil {
						current = current.prev
					}
				}

			case MI, ZE, IN: // number
				current, err = NewNode(current, buf, Number, useKey())
				if err != nil {
					break
				}

				err = buf.numeric(false)
				current.borders[1] = buf.index

				buf.index -= 1
				buf.state = OK

				if current.prev != nil {
					current = current.prev
				}

			case T1, F1: // boolean
				current, err = NewNode(current, buf, Boolean, useKey())
				if err != nil {
					break
				}

				if buf.state == T1 {
					err = buf.word(trueLiteral)
				} else {
					err = buf.word(falseLiteral)
				}

				current.borders[1] = buf.index + 1
				buf.state = OK

				if current.prev != nil {
					current = current.prev
				}

			case N1: // null
				current, err = NewNode(current, buf, Null, useKey())
				if err != nil {
					break
				}

				err = buf.word(nullLiteral)
				current.borders[1] = buf.index + 1
				buf.state = OK

				if current.prev != nil {
					current = current.prev
				}
			}
		} else {
			// region action
			switch state {
			case ec: // <empty> }
				if key != nil {
					err = unexpectedTokenError(buf.data, buf.index)
				}

				fallthrough

			case cc: // }
				if current != nil && current.IsObject() && !current.ready() {
					current.borders[1] = buf.index + 1

					if current.prev != nil {
						current = current.prev
					}
				} else {
					err = unexpectedTokenError(buf.data, buf.index)
				}

				buf.state = OK

			case bc: // ]
				if current != nil && current.IsArray() && !current.ready() {
					current.borders[1] = buf.index + 1
					if current.prev != nil {
						current = current.prev
					}
				} else {
					err = unexpectedTokenError(buf.data, buf.index)
				}

				buf.state = OK

			case co: // {
				current, err = NewNode(current, buf, Object, useKey())
				buf.state = OB

			case bo: // [
				current, err = NewNode(current, buf, Array, useKey())
				buf.state = AR

			case cm: // ,
				if current == nil {
					return nil, unexpectedTokenError(buf.data, buf.index)
				}

				if current.IsObject() {
					buf.state = KE // key expected
				} else if current.IsArray() {
					buf.state = VA // value expected
				} else {
					err = unexpectedTokenError(buf.data, buf.index)
				}

			case cl: // :
				if current == nil || !current.IsObject() || key == nil {
					err = unexpectedTokenError(buf.data, buf.index)
				} else {
					buf.state = VA
				}

			default:
				err = unexpectedTokenError(buf.data, buf.index)
			}
		}

		if err != nil {
			return nil, err
		}

		if buf.step() != nil {
			break
		}

		if _, err = buf.first(); err != nil {
			err = nil
			break
		}
	}

	var root *Node
	if current == nil || buf.state != OK {
		return nil, io.EOF
	} else {
		root = current.root()
		if !root.ready() {
			err = io.EOF
			root = nil
		}
	}

	return root, err
}

// getString extracts a string from the buffer and advances the buffer index past the string.
func getString(b *buffer) (*string, error) {
	start := b.index
	if err := b.string(DoublyQuoteToken, false); err != nil {
		return nil, err
	}

	value, ok := Unquote(b.data[start:b.index+1], DoublyQuoteToken)
	if !ok {
		return nil, unexpectedTokenError(b.data, start)
	}

	return &value, nil
}

func unexpectedTokenError(data []byte, index int) error {
	return fmt.Errorf("unexpected token at index %d. data %b", index, data)
}

func cptrs(cpy *string) *string {
	if cpy == nil {
		return nil
	}

	val := *cpy

	return &val
}

func UnmarshalSafe(data []byte) (*Node, error) {
	var safe []byte
	safe = append(safe, data...)
	return Unmarshal(safe)
}

func Must(root *Node, expect error) *Node {
	if expect != nil {
		panic(expect)
	}

	return root
}
