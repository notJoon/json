package json

import (
	"errors"
	"fmt"
	"strconv"
)

type NodeValue struct {
	value interface{}
	typ   ValueType
}

func newNodeValue(value interface{}) *NodeValue {
	return &NodeValue{
		value: value,
		typ:   typeOf(value),
	}
}

func (nv *NodeValue) Load() interface{} {
	return nv.value
}

type Node struct {
	prev     *Node
	next     map[string]*Node
	key      *string
	data     []byte
	value    *NodeValue
	index    *int
	borders  [2]int // start, end
	modified bool
}

func NewNode(prev *Node, b *buffer, typ ValueType, key **string) (*Node, error) {
	curr := &Node{
		prev:    prev,
		data:    b.data,
		borders: [2]int{b.index, 0},
		key:     *key,
		value: &NodeValue{
			value: nil,
			typ:   typ,
		},
		modified: false,
	}

	if typ == Object || typ == Array {
		curr.next = make(map[string]*Node)
	}

	if prev != nil {
		if prev.IsArray() {
			size := len(prev.next)
			curr.index = &size

			prev.next[strconv.Itoa(size)] = curr
		} else if prev.IsObject() {
			if key == nil {
				return nil, errors.New("key is required for object")
			}

			prev.next[**key] = curr
		} else {
			return nil, errors.New("invalid parent type")
		}
	}

	return curr, nil
}

func valueNode(prev *Node, _key string, typ ValueType, val interface{}) *Node {
	curr := &Node{
		prev:    prev,
		data:    nil,
		borders: [2]int{0, 0},
		value: &NodeValue{
			typ: typ,
		},
		modified: true,
	}

	if val != nil {
		curr.value = newNodeValue(val)
	}

	return curr
}

func (n *Node) Key() *string {
	if n == nil || n.key == nil {
		return nil
	}

	return n.key
}

func (n *Node) HasKey(key string) bool {
	if n == nil {
		return false
	}

	_, ok := n.next[key]
	return ok
}

func (n *Node) GetKey(key string) (*Node, error) {
	if n == nil {
		return nil, fmt.Errorf("node is nil")
	}

	if n.Type() != Object {
		return nil, fmt.Errorf("target node is not object type. got: %s", n.Type().String())
	}

	value, ok := n.next[key]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	return value, nil
}

func (n *Node) MustKey(key string) *Node {
	val, err := n.GetKey(key)
	if err != nil {
		panic(err)
	}

	return val
}

func (n *Node) Empty() bool {
	if n == nil {
		return false
	}

	return len(n.next) == 0
}

func (n *Node) Type() ValueType {
	return n.value.typ
}

func (n *Node) Value() (value interface{}, err error) {
	value = n.value.Load()

	if value == nil {
		switch n.value.typ {
		case Null:
			return nil, nil

		case Number:
			value, err = ParseFloatLiteral(n.Source())
			if err != nil {
				return nil, err
			}

			n.value = newNodeValue(value)

		case String:
			var ok bool

			value, ok = unquote(n.Source(), DoublyQuoteToken)
			if !ok {
				return "", errors.New("invalid string value")
			}

			n.value = newNodeValue(value)

		case Boolean:
			if len(n.Source()) == 0 {
				return nil, errors.New("empty boolean value")
			}

			b := n.Source()[0]
			value = b == 't' || b == 'T'
			n.value = newNodeValue(value)

		case Array:
			elems := make([]*Node, len(n.next))

			for _, e := range n.next {
				elems[*e.index] = e
			}

			value = elems
			n.value = newNodeValue(value)

		case Object:
			obj := make(map[string]*Node, len(n.next))

			for k, v := range n.next {
				obj[k] = v
			}

			value = obj
			n.value = newNodeValue(value)
		}
	}

	return value, nil
}

func (n *Node) Size() int {
	if n == nil {
		return 0
	}

	return len(n.next)
}

func (n *Node) Keys() []string {
	if n == nil {
		return nil
	}

	result := make([]string, 0, len(n.next))
	for key := range n.next {
		result = append(result, key)
	}

	return result
}

func (n *Node) Index() int {
	if n == nil {
		return -1
	}

	if n.index == nil {
		return -1
	}

	return *n.index
}

func NullNode(key string) *Node {
	return &Node{
		key:      &key,
		value:    &NodeValue{value: nil, typ: Null},
		modified: true,
	}
}

func NumberNode(key string, value float64) *Node {
	return &Node{
		key: &key,
		value: &NodeValue{
			value: value,
			typ:   Number, // treat Float and Number as Number type
		},
		modified: true,
	}
}

func StringNode(key string, value string) *Node {
	val := newNodeValue(value)

	return &Node{
		key:      &key,
		value:    val,
		modified: true,
	}
}

func BoolNode(key string, value bool) *Node {
	val := newNodeValue(value)

	return &Node{
		key:      &key,
		value:    val,
		modified: true,
	}
}

func ArrayNode(key string, value []*Node) *Node {
	curr := &Node{
		key:      &key,
		value:    &NodeValue{value: value, typ: Array},
		modified: true,
	}

	curr.next = make(map[string]*Node, len(value))
	if value != nil {
		curr.value = newNodeValue(value)

		for i, v := range value {
			var idx = i
			curr.next[strconv.Itoa(i)] = v

			v.prev = curr
			v.index = &idx
		}
	}

	return curr
}

func ObjectNode(key string, value map[string]*Node) *Node {
	curr := &Node{
		key: &key,
		value: &NodeValue{
			value: value,
			typ:   Object,
		},
		next:     value,
		modified: true,
	}

	if value != nil {
		curr.value = newNodeValue(value)
		for k, v := range value {
			v.prev = curr
			v.key = &k
		}
	} else {
		curr.next = make(map[string]*Node)
	}

	return curr
}

func (n *Node) IsArray() bool {
	return n.value.typ == Array
}

func (n *Node) IsObject() bool {
	return n.value.typ == Object
}

func (n *Node) IsBool() bool {
	return n.value.typ == Boolean
}

func (n *Node) ready() bool {
	return n.borders[1] != 0
}

func (n *Node) Source() []byte {
	if n == nil {
		return nil
	}

	if n.ready() && !n.modified && n.data != nil {
		return (n.data)[n.borders[0]:n.borders[1]]
	}

	return nil
}

func (n *Node) root() *Node {
	if n == nil {
		return nil
	}

	curr := n
	for curr.prev != nil {
		curr = curr.prev
	}

	return curr
}

func (n *Node) GetNull() (interface{}, error) {
	if n == nil {
		return nil, errors.New("node is nil")
	}

	if n.value.typ != Null {
		return nil, errors.New("node is not null")
	}

	return nil, nil
}

func (n *Node) GetNumeric() (float64, error) {
	if n == nil {
		return 0, errors.New("node is nil")
	}

	if n.value.typ != Number {
		return 0, errors.New("node is not number")
	}

	val, err := n.Value()
	if err != nil {
		return 0, err
	}

	v, ok := val.(float64)
	if !ok {
		return 0, errors.New("node is not number")
	}

	return v, nil
}

func (n *Node) GetString() (string, error) {
	if n == nil {
		return "", errors.New("string node is empty")
	}

	if n.Type() != String {
		return "", errors.New("node type is not string")
	}

	val, err := n.Value()
	if err != nil {
		return "", err
	}

	v, ok := val.(string)
	if !ok {
		return "", errors.New("node is not string")
	}

	return v, nil
}

func (n *Node) GetBool() (bool, error) {
	if n == nil {
		return false, errors.New("node is nil")
	}

	if n.value.typ != Boolean {
		return false, errors.New("node is not boolean")
	}

	val, err := n.Value()
	if err != nil {
		return false, err
	}

	v, ok := val.(bool)
	if !ok {
		return false, errors.New("node is not boolean")
	}

	return v, nil
}

// FIXME: currently, the array value is not stored in the node properly.
func (n *Node) GetArray() ([]*Node, error) {
	if n == nil {
		return nil, errors.New("node is nil")
	}

	if n.value.typ != Array {
		return nil, errors.New("node is not array")
	}

	val, err := n.Value()
	if err != nil {
		return nil, err
	}

	v, ok := val.([]*Node)
	if !ok {
		return nil, errors.New("node is not array")
	}

	return v, nil
}

func (n *Node) GetObject() (map[string]*Node, error) {
	if n == nil {
		return nil, errors.New("node is nil")
	}

	if n.Type() != Object {
		return nil, errors.New("node is not object")
	}

	val, err := n.Value()
	if err != nil {
		return nil, err
	}

	v, ok := val.(map[string]*Node)
	if !ok {
		return nil, errors.New("node is not object")
	}

	return v, nil
}

func (n *Node) MustNull() interface{} {
	v, err := n.GetNull()
	if err != nil {
		panic(err)
	}

	return v
}

func (n *Node) MustNumeric() float64 {
	v, err := n.GetNumeric()
	if err != nil {
		panic(err)
	}

	return v
}

func (n *Node) MustString() string {
	v, err := n.GetString()
	if err != nil {
		panic(err)
	}

	return v
}

func (n *Node) MustBool() bool {
	v, err := n.GetBool()
	if err != nil {
		panic(err)
	}

	return v
}

func (n *Node) MustArray() []*Node {
	v, err := n.GetArray()
	if err != nil {
		panic(err)
	}

	return v
}

func (n *Node) MustObject() map[string]*Node {
	v, err := n.GetObject()
	if err != nil {
		panic(err)
	}

	return v
}
