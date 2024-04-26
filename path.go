package json

import (
	"errors"
	"io"
)

func Path(data []byte, path string) ([]*Node, error) {
	buf := newBuffer([]byte(path))
	root, err := Unmarshal(data)
	if err != nil {
		return nil, err
	}

	result := make([]*Node, 0)
	// check first symbol
	b, err := buf.current()
	if err != nil {
		return nil, errors.New("path: invalid path")
	}
	if b != dollarSign {
		return nil, errors.New("path: invalid path (must start with $)")
	}

	var (
		start int
		found bool
		childEnd = map[byte]bool{dot: true, bracketOpen: true}
	)

	for {
		b, err := buf.current()
		if err != nil {
			break
		}
		switch b {
		case dollarSign:
			result = append(result, root)
			err = buf.step()
			found = true
		case dot: // child / descendant operator
			err = buf.step()
			if err != nil {
				break
			}
			start = buf.index
			err = buf.skipAny(childEnd)
			if err == io.EOF {
				err = nil
			}
			if err != nil {
				break
			}
			if buf.index-start == 0 {
				temp := make([]*Node, 0)
				for _, elem := range result {
					temp = append(temp, recursiveDescend(elem)...)
				}
				result = append(result, temp...)
			} else if buf.index-start == 1 && buf.data[start] == aesterisk {
				temp := make([]*Node, 0)
				for _, elem := range result {
					temp = append(temp, elem.Inheritors()...)
				}
				result = temp
			}
			found = true
		}
		if err != nil {
			break
		}
	}
	if err == io.EOF {
		if found {
			err = nil
		} else {
			err = errors.New("path: invalid path")
		}
	}
	return result, err
}

func recursiveDescend(node *Node) []*Node {
	result := make([]*Node, 0)
	if node.isContainer() {
		for _, elem := range node.Inheritors() {
			if elem.isContainer() {
				result = append(result, elem)
			}
		}
	}
	for _, elem := range result {
		result = append(result, recursiveDescend(elem)...)
	}
	return result
}