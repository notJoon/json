package json

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Node struct {
	// prev is the parent node of the current node.
	prev     *Node
	// next is the child nodes of the current node. // this is only available for object and array type nodes.
	next     map[string]*Node
	// key holds the key of the current node in the parent node.
	key      *string
	// JSON data
	data     []byte
	// value holds the value of the current node.
	value    interface{}
	nodeType ValueType
	index    *int
	// borders stores the start and end index of the current node in the data.
	borders  [2]int
	// modified indicates the current node is changed or not.
	modified bool
}

// NewNode creates a new node instance with the given parent node, buffer, type, and key.
func NewNode(prev *Node, b *buffer, typ ValueType, key **string) (*Node, error) {
	curr := &Node{
		prev:     prev,
		data:     b.data,
		borders:  [2]int{b.index, 0},
		key:      *key,
		nodeType: typ,
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

// Load retrieves the value of the current node.
func (n *Node) Load() interface{} {
	return n.value
}

// Changed checks the current node is changed or not.
func (n *Node) Changed() bool {
	return n.modified
}

// Key returns the key of the current node.
func (n *Node) Key() string {
	if n == nil || n.key == nil {
		return ""
	}

	return *n.key
}

// HasKey checks the current node has the given key or not.
func (n *Node) HasKey(key string) bool {
	if n == nil {
		return false
	}

	_, ok := n.next[key]
	return ok
}

// GetKey returns the value of the given key from the current object node.
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

// MustKey returns the value of the given key from the current object node.
func (n *Node) MustKey(key string) *Node {
	val, err := n.GetKey(key)
	if err != nil {
		panic(err)
	}

	return val
}

// EachKey traverses the current JSON nodes and collects all the keys.
//
// TODO: retrieve the nested keys of the current node.
func (n *Node) EachKey() []string {
	if n == nil {
		return nil
	}

	result := make([]string, 0, len(n.next))
	for key := range n.next {
		result = append(result, key)
	}

	return result
}

// Empty returns true if the current node is empty.
func (n *Node) Empty() bool {
	if n == nil {
		return false
	}

	return len(n.next) == 0
}

// Type returns the type (ValueType) of the current node.
func (n *Node) Type() ValueType {
	return n.nodeType
}

// Value returns the value of the current node.
func (n *Node) Value() (value interface{}, err error) {
	value = n.Load()

	if value == nil {
		switch n.nodeType {
		case Null:
			return nil, nil

		case Number:
			value, err = ParseFloatLiteral(n.Source())
			if err != nil {
				return nil, err
			}

			n.value = value

		case String:
			var ok bool

			value, ok = Unquote(n.Source(), doubleQuote)
			if !ok {
				return "", errors.New("invalid string value")
			}

			n.value = value

		case Boolean:
			if len(n.Source()) == 0 {
				return nil, errors.New("empty boolean value")
			}

			b := n.Source()[0]
			value = b == 't' || b == 'T'
			n.value = value

		case Array:
			elems := make([]*Node, len(n.next))

			for _, e := range n.next {
				elems[*e.index] = e
			}

			value = elems
			n.value = value

		case Object:
			obj := make(map[string]*Node, len(n.next))

			for k, v := range n.next {
				obj[k] = v
			}

			value = obj
			n.value = value
		}
	}

	return value, nil
}

// Set sets the value of the current node.
//
// If the given value is nil, it will be updated to null type.
func (n *Node) Set(val interface{}) error {
	if val == nil {
		return n.SetNull()
	}

	switch result := val.(type) {
		// TODO: support bigint kind (uin256, int256, etc.) if possible.
		case float64, float32, int64, int32, int16, int8, int, uint64, uint32, uint16, uint8, uint:
			if fval, err := numberKind2f64(val); err != nil {
				return err
			} else {
				return n.SetNumber(fval)
			}

		case string:
			return n.SetString(result)

		case bool:
			return n.SetBool(result)

		case []*Node:
			return n.SetArray(result)

		case map[string]*Node:
			return n.SetObject(result)

		case *Node:
			return n.SetNode(result)

		default:
			return errors.New("invalid value type")
	}
}

// SetNull sets the current node to null type.
//
// If the current node is not null type, it will be updated to null type.
func (n *Node) SetNull() error {
	return n.update(Null, nil)
}

// SetNumber sets the current node to number type.
//
// If the current node is not number type, it will be updated to number type.
func (n *Node) SetNumber(val float64) error {
	return n.update(Number, val)
}

// SetString sets the current node to string type.
//
// If the current node is not string type, it will be updated to string type.
func (n *Node) SetString(val string) error {
	return n.update(String, val)
}

// SetBool sets the current node to boolean type.
//
// If the current node is not boolean type, it will be updated to boolean type.
func (n *Node) SetBool(val bool) error {
	return n.update(Boolean, val)
}

// SetArray sets the current node to array type.
//
// If the current node is not array type, it will be updated to array type.
func (n *Node) SetArray(val []*Node) error {
	return n.update(Array, val)
}

// SetObject sets the current node to object type.
//
// If the current node is not object type, it will be updated to object type.
func (n *Node) SetObject(val map[string]*Node) error {
	return n.update(Object, val)
}

// SetNode sets the current node to the given value type node.
//
// If the current node is not the same type as the given node, it will be updated to the given node type.
func (n *Node) SetNode(val *Node) error {
	if n == nil {
		return errors.New("node is nil")
	}

	if n.isSameOrParentNode(val) {
		return errors.New("can't append same or parent node")
	}

	node := val.Clone()
	node.setRef(n.prev, n.key, n.index)
	n.setRef(nil, nil, nil)

	*n = *node
	if n.prev != nil {
		n.prev.mark()
	}

	return nil
}

// Delete removes the current node from the parent node.
func (n *Node) Delete() error {
	if n == nil {
		return errors.New("can't delete nil node")
	}

	if n.prev == nil {
		return nil
	}

	return n.prev.remove(n)
}

// Size returns the number of sub-nodes of the current Array node.
func (n *Node) Size() int {
	if n == nil {
		return 0
	}

	return len(n.next)
}

// Index returns the index of the current node in the parent array node.
func (n *Node) Index() int {
	if n == nil {
		return -1
	}

	if n.index == nil {
		return -1
	}

	return *n.index
}

// MustIndex returns the array element at the given index.
//
// If the index is negative, it returns the index is from the end of the array.
// Also, it panics if the index is not found.
func (n *Node) MustIndex(expectIdx int) *Node {
	val, err := n.GetIndex(expectIdx)
	if err != nil {
		panic(err)
	}

	return val
}

// GetIndex returns the array element at the given index.
//
// if the index is negative, it returns the index is from the end of the array.
func (n *Node) GetIndex(idx int) (*Node, error) {
	if n == nil {
		return nil, errors.New("node is nil")
	}

	if n.nodeType != Array {
		return nil, errors.New("node is not array")
	}

	if idx < 0 {
		idx += len(n.next)
	}

	child, ok := n.next[strconv.Itoa(idx)]
	if !ok {
		return nil, errors.New("index not found")
	}

	return child, nil
}

// DeleteIndex removes the array element at the given index.
func (n *Node) DeleteIndex(idx int) error {
	node, err := n.GetIndex(idx)
	if err != nil {
		return err
	}

	return node.remove(node)
}

// NullNode creates a new null type node.
func NullNode(key string) *Node {
	return &Node{
		key:      &key,
		value:    nil,
		nodeType: Null,
		modified: true,
	}
}

// NumberNode creates a new number type node.
func NumberNode(key string, value float64) *Node {
	return &Node{
		key:      &key,
		value:    value,
		nodeType: Number,
		modified: true,
	}
}

// StringNode creates a new string type node.
func StringNode(key string, value string) *Node {
	return &Node{
		key:      &key,
		value:    value,
		nodeType: String,
		modified: true,
	}
}

// BoolNode creates a new boolean type node.
func BoolNode(key string, value bool) *Node {
	return &Node{
		key:      &key,
		value:    value,
		nodeType: Boolean,
		modified: true,
	}
}

// ArrayNode creates a new array type node.
//
// If the given value is nil, it creates an empty array node.
func ArrayNode(key string, value []*Node) *Node {
	curr := &Node{
		key:      &key,
		nodeType: Array,
		modified: true,
	}

	curr.next = make(map[string]*Node, len(value))
	if value != nil {
		curr.value = value

		for i, v := range value {
			var idx = i
			curr.next[strconv.Itoa(i)] = v

			v.prev = curr
			v.index = &idx
		}
	}

	return curr
}

// ObjectNode creates a new object type node.
//
// If the given value is nil, it creates an empty object node.
//
// next is a map of key and value pairs of the object.
func ObjectNode(key string, value map[string]*Node) *Node {
	curr := &Node{
		nodeType: Object,
		key:      &key,
		next:     value,
		modified: true,
	}

	if value != nil {
		curr.value = value

		for key, val := range value {
			vkey := key
			val.prev = curr
			val.key = &vkey
		}
	} else {
		curr.next = make(map[string]*Node)
	}

	return curr
}

// IsArray returns true if the current node is array type.
func (n *Node) IsArray() bool {
	return n.nodeType == Array
}

// IsObject returns true if the current node is object type.
func (n *Node) IsObject() bool {
	return n.nodeType == Object
}

// IsNull returns true if the current node is null type.
func (n *Node) IsBool() bool {
	return n.nodeType == Boolean
}

func (n *Node) ready() bool {
	return n.borders[1] != 0
}

// Source returns the source of the current node.
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

// GetNull returns the null value if current node is null type.
func (n *Node) GetNull() (interface{}, error) {
	if n == nil {
		return nil, errors.New("node is nil")
	}

	if n.nodeType != Null {
		return nil, errors.New("node is not null")
	}

	return nil, nil
}

// MustNull returns the null value if current node is null type.
//
// It panics if the current node is not null type.
func (n *Node) MustNull() interface{} {
	v, err := n.GetNull()
	if err != nil {
		panic(err)
	}

	return v
}

// GetNumeric returns the numeric (int/float) value if current node is number type.
func (n *Node) GetNumeric() (float64, error) {
	if n == nil {
		return 0, errors.New("node is nil")
	}

	if n.nodeType != Number {
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

// MustNumeric returns the numeric (int/float) value if current node is number type.
//
// It panics if the current node is not number type.
func (n *Node) MustNumeric() float64 {
	v, err := n.GetNumeric()
	if err != nil {
		panic(err)
	}

	return v
}

// GetString returns the string value if current node is string type.
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

// MustString returns the string value if current node is string type.
//
// It panics if the current node is not string type.
func (n *Node) MustString() string {
	v, err := n.GetString()
	if err != nil {
		panic(err)
	}

	return v
}

// GetBool returns the boolean value if current node is boolean type.
func (n *Node) GetBool() (bool, error) {
	if n == nil {
		return false, errors.New("node is nil")
	}

	if n.nodeType != Boolean {
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

// MustBool returns the boolean value if current node is boolean type.
//
// It panics if the current node is not boolean type.
func (n *Node) MustBool() bool {
	v, err := n.GetBool()
	if err != nil {
		panic(err)
	}

	return v
}

// GetArray returns the array value if current node is array type.
func (n *Node) GetArray() ([]*Node, error) {
	if n == nil {
		return nil, errors.New("node is nil")
	}

	if n.nodeType != Array {
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

// MustArray returns the array value if current node is array type.
//
// It panics if the current node is not array type.
func (n *Node) MustArray() []*Node {
	v, err := n.GetArray()
	if err != nil {
		panic(err)
	}

	return v
}

// AppendArray appends the given values to the current array node.
//
// If the current node is not array type, it returns an error.
func (n *Node) AppendArray(value ...*Node) error {
	if !n.IsArray() {
		return errors.New("can't append value to non-array node")
	}

	for _, val := range value {
		if err := n.append(nil, val); err != nil {
			return err
		}
	}

	n.mark()
	return nil
}

// GetObject returns the object value if current node is object type.
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

// MustObject returns the object value if current node is object type.
//
// It panics if the current node is not object type.
func (n *Node) MustObject() map[string]*Node {
	v, err := n.GetObject()
	if err != nil {
		panic(err)
	}

	return v
}

// AppendObject appends the given key and value to the current object node.
//
// If the current node is not object type, it returns an error.
func (n *Node) AppendObject(key string, value *Node) error {
	if !n.IsObject() {
		return errors.New("can't append value to non-object node")
	}

	if err := n.append(&key, value); err != nil {
		return err
	}

	n.mark()
	return nil
}

// String converts the node to a string representation.
func (n *Node) String() string {
	if n == nil {
		return ""
	}

	if n.ready() && !n.modified {
		return string(n.Source())
	}

	val, err := Marshal(n)
	if err != nil {
		return "error: " + err.Error()
	}

	return string(val)
}

// Path builds the path of the current node.
//
// For example:
//
//	{ "key": { "sub": [ "val1", "val2" ] }}
//
// The path of "val2" is: $.key.sub[1]
func (n *Node) Path() string {
	if n == nil {
		return ""
	}

	var sb strings.Builder

	if n.prev == nil {
		sb.WriteString("$")
	} else {
		sb.WriteString(n.prev.Path())

		if n.key != nil {
			sb.WriteString("['" + n.Key() + "']")
		} else {
			sb.WriteString("[" + strconv.Itoa(n.Index()) + "]")
		}
	}

	return sb.String()
}

// update updates the current node value with the given type and value.
func (n *Node) update(vt ValueType, val interface{}) error {
	if err := n.validate(vt, val); err != nil {
		return err
	}

	n.mark()
	n.clear()

	n.nodeType = vt
	n.value = nil

	if val != nil {
		switch vt {
		case Array:
			nodes := val.([]*Node)
			n.next = make(map[string]*Node, len(nodes))

			for _, node := range nodes {
				tn := node
				if err := n.append(nil, tn); err != nil {
					return err
				}
			}

		case Object:
			obj := val.(map[string]*Node)
			n.next = make(map[string]*Node, len(obj))

			for key, node := range obj {
				tk := key
				tn := node

				if err := n.append(&tk, tn); err != nil {
					return err
				}
			}

		default:
			n.value = val
		}
	}

	return nil
}

// mark marks the current node as modified.
func (n *Node) mark() {
	node := n
	for node != nil && !node.modified {
		node.modified = true
		node = node.prev
	}
}

// clear clears the current node's value
func (n *Node) clear() {
	n.data = nil
	n.borders[1] = 0

	for key := range n.next {
		n.next[key].prev = nil
	}

	n.next = nil
}

// validate checks the current node value is matching the given type.
func (n *Node) validate(t ValueType, v interface{}) error {
	if n == nil {
		return errors.New("node is nil")
	}

	switch t {
	case Null:
		if v != nil {
			return errors.New("invalid null value")
		}

	// TODO: support uint256, int256 type later.
	case Number:
		switch v.(type) {
		case float64, int, uint:
			return nil
		default:
			return errors.New("invalid number value")
		}

	case String:
		if _, ok := v.(string); !ok {
			return errors.New("invalid string value")
		}

	case Boolean:
		if _, ok := v.(bool); !ok {
			return errors.New("invalid boolean value")
		}

	case Array:
		if _, ok := v.([]*Node); !ok {
			return errors.New("invalid array value")
		}

	case Object:
		if _, ok := v.(map[string]*Node); !ok {
			return errors.New("invalid object value")
		}
	}

	return nil
}

// isContainer checks the current node type is array or object.
func (n *Node) isContainer() bool {
	return n.IsArray() || n.IsObject()
}

// remove removes the value from the current container type node.
func (n *Node) remove(v *Node) error {
	if !n.isContainer() {
		return fmt.Errorf("can't remove value from non-array or non-object node. got=%s", n.Type().String())
	}

	if v.prev != n {
		return errors.New("invalid parent node")
	}

	n.mark()
	if n.IsArray() {
		delete(n.next, strconv.Itoa(*v.index))
		n.dropIndex(*v.index)
	} else {
		delete(n.next, *v.key)
	}

	v.prev = nil
	return nil
}

// dropIndex rebase the index of current array node values.
func (n *Node) dropIndex(idx int) {
	for i := idx + 1; i <= len(n.next); i++ {
		prv := i - 1
		if curr, ok := n.next[strconv.Itoa(i)]; ok {
			curr.index = &prv
			n.next[strconv.Itoa(prv)] = curr
		}

		delete(n.next, strconv.Itoa(i))
	}
}

// append is a helper function to append the given value to the current container type node.
func (n *Node) append(key *string, val *Node) error {
	if n.isSameOrParentNode(val) {
		return errors.New("can't append same or parent node")
	}

	if val.prev != nil {
		if err := val.prev.remove(val); err != nil {
			return err
		}
	}

	val.prev = n
	val.key = key

	if key == nil {
		size := len(n.next)
		val.index = &size
		n.next[strconv.Itoa(size)] = val
	} else {
		if old, ok := n.next[*key]; ok {
			if err := n.remove(old); err != nil {
				return err
			}
		}
		n.next[*key] = val
	}

	return nil
}

func (n *Node) isSameOrParentNode(nd *Node) bool {
	return n == nd || n.isParentNode(nd)
}

func (n *Node) isParentNode(nd *Node) bool {
	if n == nil {
		return false
	}

	for curr := nd.prev; curr != nil; curr = curr.prev {
		if curr == n {
			return true
		}
	}

	return false
}

func (n *Node) setRef(prev *Node, key *string, idx *int) {
	n.prev = prev

	if key == nil {
		n.key = nil
	} else {
		tmp := *key
		n.key = &tmp
	}

	n.index = idx
}

// Clone creates a new node instance with the same value of the current node.
func (n *Node) Clone() *Node {
	node := n.clone()
	node.setRef(nil, nil, nil)

	return node
}

func (n *Node) clone() *Node {
	node := &Node{
		prev:     n.prev,
		next:     make(map[string]*Node, len(n.next)),
		key:      n.key,
		data:     n.data,
		value:    n.value,
		nodeType: n.nodeType,
		index:    n.index,
		borders:  n.borders,
		modified: n.modified,
	}

	for k, v := range n.next {
		node.next[k] = v
	}

	return node
}
