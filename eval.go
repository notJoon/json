package json

import (
	"errors"
	"fmt"
	"strings"
)

// func Eval(node *Node, path string) (*Node, error) {
// 	rpn, err := parseFilterExpression(path)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return eval(node, rpn)
// }

func eval(node *Node, rpn []string, cmd string) (*Node, error) {
	stack := make([]*Node, 0)

	for _, expr := range rpn {
		if isOperation(expr) {
			if len(stack) < 2 {
				return nil, fmt.Errorf("insufficient operands for operator: %s", expr)
			}

			op2 := stack[len(stack)-1]
			op1 := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			result, err := applyOperation(expr, op1, op2)
			if err != nil {
				return nil, err
			}

			stack = append(stack, result)
		} else {
			if strings.HasPrefix(expr, "@") || strings.HasPrefix(expr, "$") {
				cmds, err := parsePath(expr)
				if err != nil {
					return nil, err
				}

				slice, err := applyPath(node, cmds)
				if err != nil {
					return nil, err
				}

				if len(slice) == 1 {
					stack = append(stack, slice[0])
				} else {
					return nil, nil
				}
			} else {
				bstr := []byte(expr)

				if len(bstr) > 1 && bstr[0] == '"' && bstr[len(bstr)-1] == '"' {
					bstr = bstr[1 : len(bstr)-1]
				}

				temp, err := Unmarshal(bstr)
				if err != nil {
					return nil, err
				}

				stack = append(stack, temp)
			}
		}
	}

	switch len(stack) {
	case 1:
		return stack[0], nil
	case 0:
		return nil, nil
	default:
		return nil, errors.New("invalid expression")
	}
}

func isOperation(expr string) bool {
	_, ok := operations[expr]
	return ok
}

func applyOperation(expr string, op1, op2 *Node) (*Node, error) {
	operation, ok := operations[expr]
	if !ok {
		return nil, fmt.Errorf("unknown operation: %s", expr)
	}

	result, err := operation(op1, op2)
	if err != nil {
		return nil, err
	}

	return result, nil
}

var operations = map[string]func(*Node, *Node) (*Node, error){
	"==": equalOperation,
	"!=": notEqualOperation,
	">":  greaterThanOperation,
	">=": greaterThanEqualOperation,
	"<":  lessThanOperation,
	"<=": lessThanOrEqualOperation,
	"&&": andOperation,
	"||": orOperation,
}

func equalOperation(lhs, rhs *Node) (*Node, error) {
	res, err := lhs.Eq(rhs)
	if err != nil {
		return nil, err
	}
	return BoolNode("", res), nil
}

func notEqualOperation(lhs, rhs *Node) (*Node, error) {
	res, err := lhs.Neq(rhs)
	if err != nil {
		return nil, err
	}
	return BoolNode("", res), nil
}

func greaterThanOperation(lhs, rhs *Node) (*Node, error) {
	res, err := lhs.Gt(rhs)
	if err != nil {
		return nil, err
	}
	return BoolNode("", res), nil
}

func greaterThanEqualOperation(lhs, rhs *Node) (*Node, error) {
	res, err := lhs.Gte(rhs)
	if err != nil {
		return nil, err
	}

	return BoolNode("", res), nil
}

func lessThanOperation(lhs, rhs *Node) (*Node, error) {
	res, err := lhs.Lt(rhs)
	if err != nil {
		return nil, err
	}
	return BoolNode("", res), nil
}

func lessThanOrEqualOperation(lhs, rhs *Node) (*Node, error) {
	res, err := lhs.Lte(rhs)
	if err != nil {
		return nil, err
	}
	return BoolNode("", res), nil
}

func andOperation(lhs, rhs *Node) (*Node, error) {
	l, err := boolean(lhs)
	if err != nil {
		return nil, err
	}
	if l {
		r, err := boolean(rhs)
		if err != nil {
			return nil, err
		}
		return BoolNode("", r), nil
	}
	return BoolNode("", false), nil
}

func orOperation(lhs, rhs *Node) (*Node, error) {
	l, err := boolean(lhs)
	if err != nil {
		return nil, err
	}
	if !l {
		r, err := boolean(rhs)
		if err != nil {
			return nil, err
		}
		return BoolNode("", r), nil
	}
	return BoolNode("", true), nil
}

var functions = map[string]func(*Node) (*Node, error){
	"length": lengthFunction,
}

func lengthFunction(node *Node) (*Node, error) {
	if node.isContainer() {
		return NumberNode("", float64(node.Size())), nil
	}

	if node.IsString() {
		res, err := node.GetString()
		if err != nil {
			return nil, err
		}

		return NumberNode("", float64(len(res))), nil
	}

	return NumberNode("", 1), nil
}

////////////* Node Compare *///////////////

func (n *Node) Eq(node *Node) (bool, error) {
	if n == nil || node == nil {
		return false, errors.New("node is nil")
	}

	if n.Type() == node.Type() {
		switch n.Type() {
		case Boolean:
			l, err := n.GetBool()
			if err != nil {
				return false, err
			}
			r, err := node.GetBool()
			if err != nil {
				return false, err
			}
			return l == r, nil
		case Number:
			l, err := n.GetNumeric()
			if err != nil {
				return false, err
			}
			r, err := node.GetNumeric()
			if err != nil {
				return false, err
			}
			return l == r, nil
		case String:
			l, err := n.GetString()
			if err != nil {
				return false, err
			}
			r, err := node.GetString()
			if err != nil {
				return false, err
			}
			return l == r, nil
		case Null:
			return true, nil
		case Array:
			l, err := n.GetArray()
			if err != nil {
				return false, err
			}
			r, err := node.GetArray()
			if err != nil {
				return false, err
			}
			if len(l) == len(r) {
				for i := range l {
					result, err := l[i].Eq(r[i])
					if err != nil || !result {
						return false, err
					}
				}
			}
		case Object:
			l, err := n.GetObject()
			if err != nil {
				return false, err
			}
			r, err := n.GetObject()
			if err != nil {
				return false, err
			}
			if len(l) == len(r) {
				result := true
				for i := range l {
					elem, ok := r[i]
					if !ok {
						return false, nil
					}
					result, err = l[i].Eq(elem)
					if err != nil || !result {
						return false, err
					}
				}
			}
		}
	}
	return true, nil
}

func (n *Node) Neq(node *Node) (bool, error) {
	result, err := n.Eq(node)
	if err != nil {
		return false, err
	}
	return !result, nil
}

func (n *Node) Gt(node *Node) (bool, error) {
	if n == nil || node == nil {
		return false, errors.New("node is nil")
	}

	if n.Type() == node.Type() {
		switch n.Type() {
		case Number:
			lhs, rhs, err := getFloats(n, node)
			if err != nil {
				return false, err
			}
			return lhs > rhs, nil
		case String:
			lhs, rhs, err := getStrings(n, node)
			if err != nil {
				return false, err
			}
			return lhs > rhs, nil
		default:
			return false, errors.New("unsupported type")
		}
	}
	return false, nil
}

func (n *Node) Gte(node *Node) (bool, error) {
	if n == nil || node == nil {
		return false, errors.New("node is nil")
	}

	if n.Type() == node.Type() {
		switch n.Type() {
		case Number:
			lhs, rhs, err := getFloats(n, node)
			if err != nil {
				return false, err
			}
			return lhs >= rhs, nil
		case String:
			lhs, rhs, err := getStrings(n, node)
			if err != nil {
				return false, err
			}
			return lhs >= rhs, nil
		default:
			return false, errors.New("unsupported type")
		}
	}
	return false, nil
}

func (n *Node) Lt(node *Node) (bool, error) {
	if n == nil || node == nil {
		return false, errors.New("node is nil")
	}

	if n.Type() == node.Type() {
		switch n.Type() {
		case Number:
			lhs, rhs, err := getFloats(n, node)
			if err != nil {
				return false, err
			}
			return lhs < rhs, nil
		case String:
			lhs, rhs, err := getStrings(n, node)
			if err != nil {
				return false, err
			}
			return lhs < rhs, nil
		default:
			return false, errors.New("unsupported type")
		}
	}
	return false, nil
}

func (n *Node) Lte(node *Node) (bool, error) {
	if n == nil || node == nil {
		return false, errors.New("node is nil")
	}

	if n.Type() == node.Type() {
		switch n.Type() {
		case Number:
			lhs, rhs, err := getFloats(n, node)
			if err != nil {
				return false, err
			}
			return lhs <= rhs, nil
		case String:
			lhs, rhs, err := getStrings(n, node)
			if err != nil {
				return false, err
			}
			return lhs <= rhs, nil
		default:
			return false, errors.New("unsupported type")
		}
	}
	return false, nil
}

func getFloats(l, r *Node) (float64, float64, error) {
	lhs, err := l.GetNumeric()
	if err != nil {
		return 0, 0, err
	}
	rhs, err := r.GetNumeric()
	if err != nil {
		return 0, 0, err
	}
	return lhs, rhs, nil
}

func getStrings(l, r *Node) (string, string, error) {
	lhs, err := l.GetString()
	if err != nil {
		return "", "", err
	}
	rhs, err := r.GetString()
	if err != nil {
		return "", "", err
	}
	return lhs, rhs, nil
}

func boolean(node *Node) (bool, error) {
	switch node.Type() {
	case Boolean:
		return node.GetBool()
	case Number:
		if n, err := node.GetNumeric(); err == nil {
			return n != 0, nil
		}
	case String:
		if s, err := node.GetString(); err == nil {
			return s != "", nil
		}
	case Null:
		return false, nil
	case Array:
		fallthrough
	case Object:
		return !node.Empty(), nil
	}
	return false, nil
}
