package json

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"
)

var (
	nilKey   *string
	dummyKey = "key"
)

type _args struct {
	prev *Node
	buf  *buffer
	typ  ValueType
	key  **string
}

type simpleNode struct {
	name string
	node *Node
}

func TestNode_CreateNewNode(t *testing.T) {
	rel := &dummyKey

	tests := []struct {
		name        string
		args        _args
		expectCurr  *Node
		expectErr   bool
		expectPanic bool
	}{
		{
			name: "object with empty key",
			args: _args{
				prev: ObjectNode("", make(map[string]*Node)),
				buf:  newBuffer(make([]byte, 10)),
				typ:  Boolean,
				key:  &nilKey,
			},
			expectCurr:  nil,
			expectPanic: true,
		},
		{
			name: "child for non container type",
			args: _args{
				prev: BoolNode("", true),
				buf:  newBuffer(make([]byte, 10)),
				typ:  Boolean,
				key:  &rel,
			},
			expectCurr: nil,
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if tt.expectPanic {
						return
					}
					t.Errorf("%s panic occurred when not expected: %v", tt.name, r)
				} else if tt.expectPanic {
					t.Errorf("%s expected panic but didn't occur", tt.name)
				}
			}()

			got, err := NewNode(tt.args.prev, tt.args.buf, tt.args.typ, tt.args.key)
			if (err != nil) != tt.expectErr {
				t.Errorf("%s error = %v, expect error %v", tt.name, err, tt.expectErr)
				return
			}

			if tt.expectErr {
				return
			}

			if !compareNodes(got, tt.expectCurr) {
				t.Errorf("%s got = %v, want %v", tt.name, got, tt.expectCurr)
			}
		})
	}
}

func TestNode_Value(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		_type       ValueType
		expected    interface{}
		errExpected bool
	}{
		{name: "null", data: []byte("null"), _type: Null, expected: nil},
		{name: "1", data: []byte("1"), _type: Number, expected: float64(1)},
		{name: ".1", data: []byte(".1"), _type: Number, expected: float64(.1)},
		{name: "-.1e1", data: []byte("-.1e1"), _type: Number, expected: float64(-1)},
		{name: "string", data: []byte("\"foo\""), _type: String, expected: "foo"},
		{name: "space", data: []byte("\"foo bar\""), _type: String, expected: "foo bar"},
		{name: "true", data: []byte("true"), _type: Boolean, expected: true},
		{name: "invalid true", data: []byte("tru"), _type: Unknown, errExpected: true},
		{name: "invalid false", data: []byte("fals"), _type: Unknown, errExpected: true},
		{name: "false", data: []byte("false"), _type: Boolean, expected: false},
		{name: "e1", data: []byte("e1"), _type: Unknown, errExpected: true},
		{name: "1a", data: []byte("1a"), _type: Unknown, errExpected: true},
		{name: "string error", data: []byte("\"foo\nbar\""), _type: String, errExpected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			curr := &Node{
				data:     tt.data,
				nodeType: tt._type,
				borders:  [2]int{0, len(tt.data)},
			}

			got, err := curr.Value()
			if err != nil {
				if !tt.errExpected {
					t.Errorf("%s error = %v, expect error %v", tt.name, err, tt.errExpected)
				}
				return
			}

			if got != tt.expected {
				t.Errorf("%s got = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestNode_Set(t *testing.T) {
	node := func(data string) *Node {
		bs := []byte(data)
		return Must(Unmarshal(bs))
	}

	tests := []struct {
		name   string
		node   *Node
		getter func(root *Node) *Node
		value  interface{}
		expect string
		isErr  bool
	}{
		{
			name:   "null -> f64",
			node:   node("null"),
			value:  float64(3.141592),
			expect: "3.141592",
			isErr:  false,
		},
		{
			name:   "null -> f32 with integer value",
			node:   node("null"),
			value:  float32(42),
			expect: "42",
			isErr:  false,
		},
		{
			name:   "null -> i64",
			node:   node("null"),
			value:  int64(42),
			expect: "42",
		},
		{
			name:   "null -> i32",
			node:   node("null"),
			value:  int32(42),
			expect: "42",
		},
		{
			name:   "null -> i16",
			node:   node("null"),
			value:  int16(42),
			expect: "42",
		},
		{
			name:   "null -> i8",
			node:   node("null"),
			value:  int8(42),
			expect: "42",
		},
		{
			name:   "null -> int",
			node:   node("null"),
			value:  int(42),
			expect: "42",
		},
		{
			name:   "null -> uint",
			node:   node("null"),
			value:  uint(42),
			expect: "42",
		},
		{
			name:   "null -> uint64",
			node:   node("null"),
			value:  uint64(42),
			expect: "42",
		},
		{
			name:   "null -> uint32",
			node:   node("null"),
			value:  uint32(42),
			expect: "42",
		},
		{
			name:   "null -> bool (true)",
			node:   node("null"),
			value:  true,
			expect: "true",
		},
		{
			name:   "null -> bool (false)",
			node:   node("null"),
			value:  false,
			expect: "false",
		},
		{
			name:   "array -> string",
			node:   node("[1, 2, 3]"),
			value:  "foobar",
			expect: `"foobar"`,
		},
		{
			name:   "object -> bool",
			node:   node(`{"foo": ["bar"], "baz": 42}`),
			value:  true,
			expect: "true",
		},
		{
			name: "object[value] -> null",
			node: node(`{"foo": ["bar"]}`),
			getter: func(root *Node) *Node {
				return root.MustKey("foo").MustIndex(0)
			},
			value:  nil,
			expect: `{"foo":[null]}`,
			isErr:  false,
		},
		{
			name: "array[v] -> array[Node]",
			node: node(`[null]`),
			getter: func(root *Node) *Node {
				return root.MustIndex(0)
			},
			value:  map[string]*Node{},
			expect: `[{}]`,
			isErr:  false,
		},
		{
			name: "array[v] -> array[array[]]",
			node: node(`[null]`),
			getter: func(root *Node) *Node {
				return root.MustIndex(0)
			},
			value:  []*Node{node(`1`)},
			expect: `[[1]]`,
			isErr:  false,
		},
		{
			name:  "nil",
			node:  nil,
			value: nil,
			isErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := tt.node
			if tt.getter != nil {
				val = tt.getter(tt.node)
			}

			if err := val.Set(tt.value); (err != nil) != tt.isErr {
				t.Errorf("%s error = %v, wantErr %v", tt.name, err, tt.isErr)
			}

			if tt.isErr {
				return
			}

			if tt.node.String() != tt.expect {
				t.Errorf("%s got = %v, want %v", tt.name, tt.node.String(), tt.expect)
			}
		})
	}
}

func TestNode_Delete(t *testing.T) {
	root := Must(Unmarshal([]byte(`{"foo":"bar"}`)))
	if err := root.Delete(); err != nil {
		t.Errorf("Delete returns error: %v", err)
	}

	if value, err := Marshal(root); err != nil {
		t.Errorf("Marshal returns error: %v", err)
	} else if string(value) != `{"foo":"bar"}` {
		t.Errorf("Marshal returns wrong value: %s", string(value))
	}

	foo := root.MustKey("foo")
	if err := foo.Delete(); err != nil {
		t.Errorf("Delete returns error while handling foo: %v", err)
	}

	if value, err := Marshal(root); err != nil {
		t.Errorf("Marshal returns error: %v", err)
	} else if string(value) != `{}` {
		t.Errorf("Marshal returns wrong value: %s", string(value))
	}

	if value, err := Marshal(foo); err != nil {
		t.Errorf("Marshal returns error: %v", err)
	} else if string(value) != `"bar"` {
		t.Errorf("Marshal returns wrong value: %s", string(value))
	}

	if foo.prev != nil {
		t.Errorf("foo.prev should be nil")
	}
}

func TestNode_ObjectNode(t *testing.T) {
	objs := map[string]*Node{
		"key1": NullNode("null"),
		"key2": NumberNode("answer", 42),
		"key3": StringNode("string", "foobar"),
		"key4": BoolNode("bool", true),
	}

	node := ObjectNode("test", objs)

	if len(node.next) != len(objs) {
		t.Errorf("ObjectNode: want %v got %v", len(objs), len(node.next))
	}

	for k, v := range objs {
		if node.next[k] == nil {
			t.Errorf("ObjectNode: want %v got %v", v, node.next[k])
		}
	}
}

func TestNode_AppendObject(t *testing.T) {
	if err := Must(Unmarshal([]byte(`{"foo":"bar","baz":null}`))).AppendObject("biz", NullNode("")); err != nil {
		t.Errorf("AppendArray should return error")
	}

	root := Must(Unmarshal([]byte(`{"foo":"bar"}`)))
	if err := root.AppendObject("baz", NullNode("")); err != nil {
		t.Errorf("AppendObject should not return error: %s", err)
	}

	if value, err := Marshal(root); err != nil {
		t.Errorf("Marshal returns error: %v", err)
	} else if isSameObject(string(value), `"{"foo":"bar","baz":null}"`) {
		t.Errorf("Marshal returns wrong value: %s", string(value))
	}

	// FIXME: this may fail if execute test in more than 3 times in a row.
	if err := root.AppendObject("biz", NumberNode("", 42)); err != nil {
		t.Errorf("AppendObject returns error: %v", err)
	}

	val, err := Marshal(root)

	if err != nil {
		t.Errorf("Marshal returns error: %v", err)
	}

	// FIXME: this may fail if execute test in more than 3 times in a row.
	if isSameObject(string(val), `"{"foo":"bar","baz":null,"biz":42}"`) {
		t.Errorf("Marshal returns wrong value: %s", string(val))
	}
}

func TestNode_ArrayNode(t *testing.T) {
	arr := []*Node{
		NullNode("nil"),
		NumberNode("num", 42),
		StringNode("str", "foobar"),
		BoolNode("bool", true),
	}

	node := ArrayNode("test", arr)

	if len(node.next) != len(arr) {
		t.Errorf("ArrayNode: want %v got %v", len(arr), len(node.next))
	}

	for i, v := range arr {
		if node.next[strconv.Itoa(i)] == nil {
			t.Errorf("ArrayNode: want %v got %v", v, node.next[strconv.Itoa(i)])
		}
	}
}

func TestNode_AppendArray(t *testing.T) {
	if err := Must(Unmarshal([]byte(`[{"foo":"bar"}]`))).AppendArray(NullNode("")); err != nil {
		t.Errorf("should return error")
	}

	root := Must(Unmarshal([]byte(`[{"foo":"bar"}]`)))
	if err := root.AppendArray(NullNode("")); err != nil {
		t.Errorf("should not return error: %s", err)
	}

	if value, err := Marshal(root); err != nil {
		t.Errorf("Marshal returns error: %v", err)
	} else if string(value) != `[{"foo":"bar"},null]` {
		t.Errorf("Marshal returns wrong value: %s", string(value))
	}

	if err := root.AppendArray(
		NumberNode("", 1),
		StringNode("", "foo"),
		Must(Unmarshal([]byte(`[0,1,null,true,"example"]`))),
		Must(Unmarshal([]byte(`{"foo": true, "bar": null, "baz": 123}`))),
	); err != nil {
		t.Errorf("AppendArray returns error: %v", err)
	}

	if value, err := Marshal(root); err != nil {
		t.Errorf("Marshal returns error: %v", err)
	} else if string(value) != `[{"foo":"bar"},null,1,"foo",[0,1,null,true,"example"],{"foo": true, "bar": null, "baz": 123}]` {
		t.Errorf("Marshal returns wrong value: %s", string(value))
	}
}

/******** value getter ********/

func TestNode_GetBool(t *testing.T) {
	root, err := Unmarshal([]byte(`true`))
	if err != nil {
		t.Errorf("Error on Unmarshal(): %s", err.Error())
		return
	}

	value, err := root.GetBool()
	if err != nil {
		t.Errorf("Error on root.GetBool(): %s", err.Error())
	}

	if !value {
		t.Errorf("root.GetBool() is corrupted")
	}
}

func TestNode_GetBool_Fail(t *testing.T) {
	tests := []simpleNode{
		{"nil node", (*Node)(nil)},
		{"literally null node", NullNode("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.node.GetBool(); err == nil {
				t.Errorf("%s should be an error", tt.name)
			}
		})
	}
}

// countBooleans test function that counts the number of true and false values in a slice of booleans.
// because map doesn't maintain order, we can't use it to test the order of the values.
func countBooleans(bools []bool) (trueCount, falseCount int) {
	for _, b := range bools {
		if b {
			trueCount++
		} else {
			falseCount++
		}
	}
	return
}

func TestNode_GetAllBools(t *testing.T) {
	tests := []struct {
		name           string
		json           string
		expectedTrues  int
		expectedFalses int
	}{
		{
			name:           "no booleans",
			json:           `{"foo": "bar", "number": 123}`,
			expectedTrues:  0,
			expectedFalses: 0,
		},
		{
			name:           "single boolean true",
			json:           `{"isTrue": true}`,
			expectedTrues:  1,
			expectedFalses: 0,
		},
		{
			name:           "single boolean false",
			json:           `{"isFalse": false}`,
			expectedTrues:  0,
			expectedFalses: 1,
		},
		{
			name:           "multiple booleans",
			json:           `{"first": true, "info": {"second": false, "third": true}, "array": [false, true, false]}`,
			expectedTrues:  3,
			expectedFalses: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root, err := Unmarshal([]byte(tc.json))
			if err != nil {
				t.Errorf("Failed to unmarshal JSON: %v", err)
				return
			}

			bools := root.GetBools()
			trueCount, falseCount := countBooleans(bools)
			if trueCount != tc.expectedTrues || falseCount != tc.expectedFalses {
				t.Errorf("%s: expected %d true and %d false booleans, got %d true and %d false", tc.name, tc.expectedTrues, tc.expectedFalses, trueCount, falseCount)
			}
		})
	}
}

// GetBools function has implemented to use recursion before, but after the benchmark test, it was found that it's slower than the current (DFS) implementation.
// So, the recursion implementation was removed.
//
// Also, this change has affected to other multiple value getter functions.
func BenchmarkNode_GetAllBools(b *testing.B) {
	root, err := Unmarshal([]byte(`{"first": true, "info": {"second": false, "third": true}, "array": [false, true, false]}`))
	if err != nil {
		b.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		root.GetBools()
	}
}

func TestNode_IsBool(t *testing.T) {
	tests := []simpleNode{
		{"true", BoolNode("", true)},
		{"false", BoolNode("", false)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.node.IsBool() {
				t.Errorf("%s should be a bool", tt.name)
			}
		})
	}
}

func TestNode_IsBool_With_Unmarshal(t *testing.T) {
	tests := []struct {
		name string
		json []byte
		want bool
	}{
		{"true", []byte("true"), true},
		{"false", []byte("false"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Unmarshal(tt.json)
			if err != nil {
				t.Errorf("Error on Unmarshal(): %s", err.Error())
			}

			if root.IsBool() != tt.want {
				t.Errorf("%s should be a bool", tt.name)
			}
		})
	}
}

var nullJson = []byte(`null`)

func TestNode_GetNull(t *testing.T) {
	root, err := Unmarshal(nullJson)
	if err != nil {
		t.Errorf("Error on Unmarshal(): %s", err.Error())
	}

	value, err := root.GetNull()
	if err != nil {
		t.Errorf("error occurred while getting null, %s", err)
	}

	if value != nil {
		t.Errorf("value is not matched. expected: nil, got: %v", value)
	}
}

func TestNode_GetNull_Fail(t *testing.T) {
	tests := []simpleNode{
		{"nil node", (*Node)(nil)},
		{"number node is null", NumberNode("", 42)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.node.GetNull(); err == nil {
				t.Errorf("%s should be an error", tt.name)
			}
		})
	}
}

func TestNode_MustNull(t *testing.T) {
	root, err := Unmarshal(nullJson)
	if err != nil {
		t.Errorf("Error on Unmarshal(): %s", err.Error())
	}

	value := root.MustNull()
	if value != nil {
		t.Errorf("value is not matched. expected: nil, got: %v", value)
	}
}

func TestNode_GetNumeric_Float(t *testing.T) {
	root, err := Unmarshal([]byte(`123.456`))
	if err != nil {
		t.Errorf("Error on Unmarshal(): %s", err)
		return
	}

	value, err := root.GetNumeric()
	if err != nil {
		t.Errorf("Error on root.GetNumeric(): %s", err)
	}

	if value != float64(123.456) {
		t.Errorf(fmt.Sprintf("value is not matched. expected: 123.456, got: %v", value))
	}
}

func TestNode_GetNumeric_Scientific_Notation(t *testing.T) {
	root, err := Unmarshal([]byte(`1e3`))
	if err != nil {
		t.Errorf("Error on Unmarshal(): %s", err)
		return
	}

	value, err := root.GetNumeric()
	if err != nil {
		t.Errorf("Error on root.GetNumeric(): %s", err)
	}

	if value != float64(1000) {
		t.Errorf(fmt.Sprintf("value is not matched. expected: 1000, got: %v", value))
	}
}

func TestNode_GetNumeric_With_Unmarshal(t *testing.T) {
	root, err := Unmarshal([]byte(`123`))
	if err != nil {
		t.Errorf("Error on Unmarshal(): %s", err)
		return
	}

	value, err := root.GetNumeric()
	if err != nil {
		t.Errorf("Error on root.GetNumeric(): %s", err)
	}

	if value != float64(123) {
		t.Errorf(fmt.Sprintf("value is not matched. expected: 123, got: %v", value))
	}
}

func TestNode_GetNumeric_Fail(t *testing.T) {
	tests := []simpleNode{
		{"nil node", (*Node)(nil)},
		{"null node", NullNode("")},
		{"string node", StringNode("", "123")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.node.GetNumeric(); err == nil {
				t.Errorf("%s should be an error", tt.name)
			}
		})
	}
}

func TestNode_GetAllIntsFromNode(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected []int
	}{
		{
			name:     "no numbers",
			json:     `{"foo": "bar", "baz": "qux"}`,
			expected: []int{},
		},
		{
			name:     "mixed numbers",
			json:     `{"int": 42, "float": 3.14, "nested": {"int2": 24, "array": [1, 2.34, 3]}}`,
			expected: []int{42, 24, 1, 3},
		},
		{
			name:     "all floats",
			json:     `{"float1": 0.1, "float2": 4.5}`,
			expected: []int{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root, err := Unmarshal([]byte(tc.json))
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			got := root.GetInts()
			for _, exp := range tc.expected {
				if !containsInts(got, exp) {
					t.Errorf("Expected integer %d is missing", exp)
				}
			}
		})
	}
}

// containsInts checks if a slice of integers contains a specific value.
func containsInts(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func TestNode_GetAllFloatsFromNode(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected []float64
	}{
		{
			name:     "no numbers",
			json:     `{"foo": "bar", "baz": "qux"}`,
			expected: []float64{},
		},
		{
			name:     "mixed numbers",
			json:     `{"int": 42, "float": 3.14, "nested": {"float2": 2.34, "array": [1, 2.34, 3]}}`,
			expected: []float64{3.14, 2.34, 2.34},
		},
		{
			name:     "all ints",
			json:     `{"int1": 1, "int2": 2}`,
			expected: []float64{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root, err := Unmarshal([]byte(tc.json))
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			got := root.GetFloats()
			for _, exp := range tc.expected {
				if !containsFloats(got, exp) {
					t.Errorf("Expected float %f is missing", exp)
				}
			}
		})
	}
}

// containsFloats checks if a slice of floats contains a specific value.
func containsFloats(slice []float64, val float64) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func TestNode_GetString(t *testing.T) {
	root, err := Unmarshal([]byte(`"123foobar 3456"`))
	if err != nil {
		t.Errorf("Error on Unmarshal(): %s", err)
	}

	value, err := root.GetString()
	if err != nil {
		t.Errorf("Error on root.GetString(): %s", err)
	}

	if value != "123foobar 3456" {
		t.Errorf(fmt.Sprintf("value is not matched. expected: 123, got: %s", value))
	}
}

func TestNode_GetString_Fail(t *testing.T) {
	tests := []simpleNode{
		{"nil node", (*Node)(nil)},
		{"null node", NullNode("")},
		{"number node", NumberNode("", 123)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.node.GetString(); err == nil {
				t.Errorf("%s should be an error", tt.name)
			}
		})
	}
}

func TestNode_GetAllStringsFromNode(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected []string
	}{
		{
			name:     "no strings",
			json:     `{"int": 42, "float": 3.14}`,
			expected: []string{},
		},
		{
			name:     "multiple strings",
			json:     `{"str": "hello", "nested": {"str2": "world"}, "array": ["foo", "bar"]}`,
			expected: []string{"hello", "world", "foo", "bar"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root, err := Unmarshal([]byte(tc.json))
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			got := root.GetStrings()
			if len(got) != len(tc.expected) {
				t.Errorf("Expected %d strings, got %d", len(tc.expected), len(got))
				return
			}
			for _, exp := range tc.expected {
				found := false
				for _, str := range got {
					if str == exp {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected string '%s' not found", exp)
				}
			}
		})
	}
}

func TestNode_MustString(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"foo", []byte(`"foo"`)},
		{"foo bar", []byte(`"foo bar"`)},
		{"", []byte(`""`)},
		{"안녕하세요", []byte(`"안녕하세요"`)},
		{"こんにちは", []byte(`"こんにちは"`)},
		{"你好", []byte(`"你好"`)},
		{"one \"encoded\" string", []byte(`"one \"encoded\" string"`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Unmarshal(tt.data)
			if err != nil {
				t.Errorf("Error on Unmarshal(): %s", err)
			}

			value := root.MustString()
			if value != tt.name {
				t.Errorf("value is not matched. expected: %s, got: %s", tt.name, value)
			}
		})
	}
}

func TestUnmarshal_Array(t *testing.T) {
	root, err := Unmarshal([]byte(" [1,[\"1\",[1,[1,2,3]]]]\r\n"))

	if err != nil {
		t.Errorf("Error on Unmarshal: %s", err.Error())
	}

	if root == nil {
		t.Errorf("Error on Unmarshal: root is nil")
	}

	if root.Type() != Array {
		t.Errorf("Error on Unmarshal: wrong type")
	}

	array, err := root.GetArray()
	if err != nil {
		t.Errorf("error occurred while getting array, %s", err)
	} else if len(array) != 2 {
		t.Errorf("expected 2 elements, got %d", len(array))
	} else if val, err := array[0].GetNumeric(); err != nil {
		t.Errorf("value of array[0] is not numeric. got: %v", array[0].value)
	} else if val != 1 {
		t.Errorf("Error on array[0].GetNumeric(): expected to be '1', got: %v", val)
	} else if val, err := array[1].GetArray(); err != nil {
		t.Errorf("error occurred while getting array, %s", err.Error())
	} else if len(val) != 2 {
		t.Errorf("Error on array[1].GetArray(): expected 2 elements, got %d", len(val))
	} else if el, err := val[0].GetString(); err != nil {
		t.Errorf("error occurred while getting string, %s", err.Error())
	} else if el != "1" {
		t.Errorf("Error on val[0].GetString(): expected to be '1', got: %s", el)
	}
}

var sampleArr = []byte(`[-1, 2, 3, 4, 5, 6]`)

func TestNode_GetArray(t *testing.T) {
	root, err := Unmarshal(sampleArr)
	if err != nil {
		t.Errorf("Error on Unmarshal(): %s", err)
		return
	}

	array, err := root.GetArray()
	if err != nil {
		t.Errorf("Error on root.GetArray(): %s", err)
	}

	if len(array) != 6 {
		t.Errorf(fmt.Sprintf("length is not matched. expected: 3, got: %d", len(array)))
	}

	for i, node := range array {
		for j, val := range []int{-1, 2, 3, 4, 5, 6} {
			if i == j {
				if v, err := node.GetNumeric(); err != nil {
					t.Errorf(fmt.Sprintf("Error on node.GetNumeric(): %s", err))
				} else if v != float64(val) {
					t.Errorf(fmt.Sprintf("value is not matched. expected: %d, got: %v", val, v))
				}
			}
		}
	}
}

func TestNode_GetArray_Fail(t *testing.T) {
	tests := []simpleNode{
		{"nil node", (*Node)(nil)},
		{"null node", NullNode("")},
		{"number node", NumberNode("", 123)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.node.GetArray(); err == nil {
				t.Errorf("%s should be an error", tt.name)
			}
		})
	}
}

func TestNode_IsArray(t *testing.T) {
	root, err := Unmarshal(sampleArr)
	if err != nil {
		t.Errorf("Error on Unmarshal(): %s", err)
		return
	}

	if root.Type() != Array {
		t.Errorf(fmt.Sprintf("Must be an array. got: %s", root.Type().String()))
	}
}

func TestNode_ArrayEach(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected []int
	}{
		{
			name:     "empty array",
			json:     `[]`,
			expected: []int{},
		},
		{
			name:     "single element",
			json:     `[42]`,
			expected: []int{42},
		},
		{
			name:     "multiple elements",
			json:     `[1, 2, 3, 4, 5]`,
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "multiple elements but all values are same",
			json:     `[1, 1, 1, 1, 1]`,
			expected: []int{1, 1, 1, 1, 1},
		},
		{
			name:     "multiple elements with non-numeric values",
			json:     `["a", "b", "c", "d", "e"]`,
			expected: []int{},
		},
		{
			name:     "non-array node",
			json:     `{"not": "an array"}`,
			expected: []int{},
		},
		{
			name:     "array containing numeric and non-numeric elements",
			json:     `["1", 2, 3, "4", 5, "6"]`,
			expected: []int{2, 3, 5},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := Must(Unmarshal([]byte(tc.json)))

			var result []int // callback result
			root.ArrayEach(func(index int, element *Node) {
				if val, err := strconv.Atoi(element.String()); err == nil {
					result = append(result, val)
				}
			})

			if len(result) != len(tc.expected) {
				t.Errorf("%s: expected %d elements, got %d", tc.name, len(tc.expected), len(result))
				return
			}

			for i, val := range result {
				if val != tc.expected[i] {
					t.Errorf("%s: expected value at index %d to be %d, got %d", tc.name, i, tc.expected[i], val)
				}
			}
		})
	}
}

func TestNode_Key(t *testing.T) {
	root, err := Unmarshal([]byte(`{"foo": true, "bar": null, "baz": 123, "biz": [1,2,3]}`))
	if err != nil {
		t.Errorf("Error on Unmarshal(): %s", err.Error())
	}

	obj := root.MustObject()
	for key, node := range obj {
		if key != node.Key() {
			t.Errorf("Key() = %v, want %v", node.Key(), key)
		}
	}

	keys := []string{"foo", "bar", "baz", "biz"}
	for _, key := range keys {
		if obj[key].Key() != key {
			t.Errorf("Key() = %v, want %v", obj[key].Key(), key)
		}
	}

	if root.MustKey("foo").Clone().Key() != "" {
		t.Errorf("wrong key found for cloned key")
	}

	if (*Node)(nil).Key() != "" {
		t.Errorf("wrong key found for nil node")
	}
}

func TestNode_Size(t *testing.T) {
	root, err := Unmarshal(sampleArr)
	if err != nil {
		t.Errorf("error occurred while unmarshal")
	}

	size := root.Size()
	if size != 6 {
		t.Errorf(fmt.Sprintf("Size() must be 6. got: %v", size))
	}

	if (*Node)(nil).Size() != 0 {
		t.Errorf(fmt.Sprintf("Size() must be 0. got: %v", (*Node)(nil).Size()))
	}
}

func TestNode_Index(t *testing.T) {
	root, err := Unmarshal([]byte(`[1, 2, 3, 4, 5, 6]`))
	if err != nil {
		t.Error("error occurred while unmarshal")
	}

	arr := root.MustArray()
	for i, node := range arr {
		if i != node.Index() {
			t.Errorf(fmt.Sprintf("Index() must be nil. got: %v", i))
		}
	}
}

func TestNode_Index_Fail(t *testing.T) {
	tests := []struct {
		name string
		node *Node
		want int
	}{
		{"nil node", (*Node)(nil), -1},
		{"null node", NullNode(""), -1},
		{"object node", ObjectNode("", nil), -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.Index(); got != tt.want {
				t.Errorf("Index() = %v, want %v", got, tt.want)
			}
		})

	}
}

func TestNode_GetIndex(t *testing.T) {
	root, err := Unmarshal([]byte(`[1, 2, 3, 4, 5, 6]`))
	if err != nil {
		t.Errorf("error occurred while unmarshal")
	}

	expected := []int{1, 2, 3, 4, 5, 6}
	if len(expected) != root.Size() {
		t.Errorf("expected length and root size is not matched. expected: %d, got: %d", len(expected), root.Size())
	}

	for i, v := range expected {
		val, err := root.GetIndex(i)
		if err != nil {
			t.Errorf("error occurred while getting index %d, %s", i, err)
		}

		if val.MustNumeric() != float64(v) {
			t.Errorf("value is not matched. expected: %d, got: %v", v, val.MustNumeric())
		}
	}
}

func TestNode_GetIndex_InputIndex_Exceed_Original_Node_Index(t *testing.T) {
	root, err := Unmarshal([]byte(`[1, 2, 3, 4, 5, 6]`))
	if err != nil {
		t.Errorf("error occurred while unmarshal")
	}

	_, err = root.GetIndex(10)
	if err == nil {
		t.Errorf("GetIndex should return error")
	}
}

func TestNode_GetKey(t *testing.T) {
	root, err := Unmarshal([]byte(`{"foo": true, "bar": null}`))
	if err != nil {
		t.Error("error occurred while unmarshal")
	}

	value, err := root.GetKey("foo")
	if err != nil {
		t.Errorf("error occurred while getting key, %s", err)
	}

	if value.MustBool() != true {
		t.Errorf("value is not matched. expected: true, got: %v", value.MustBool())
	}

	value, err = root.GetKey("bar")
	if err != nil {
		t.Errorf("error occurred while getting key, %s", err)
	}

	_, err = root.GetKey("baz")
	if err == nil {
		t.Errorf("key baz is not exist. must be failed")
	}

	if value.MustNull() != nil {
		t.Errorf("value is not matched. expected: nil, got: %v", value.MustNull())
	}
}

func TestNode_GetKey_Fail(t *testing.T) {
	tests := []simpleNode{
		{"nil node", (*Node)(nil)},
		{"null node", NullNode("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.node.GetKey(""); err == nil {
				t.Errorf("%s should be an error", tt.name)
			}
		})
	}
}

func TestNode_EachKey(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected []string
	}{
		{
			name:     "simple foo/bar",
			json:     `{"foo": true, "bar": null}`,
			expected: []string{"foo", "bar"},
		},
		{
			name:     "empty object",
			json:     `{}`,
			expected: []string{},
		},
		{
			name: "nested object",
			json: `{
				"outer": {
					"inner": {
						"key": "value"
					},
					"array": [1, 2, 3]
				},
				"another": "item"
			}`,
			expected: []string{"outer", "inner", "key", "array", "another"},
		},
		{
			name: "complex object",
			json: `{
				"Image": {
					"Width": 800,
					"Height": 600,
					"Title": "View from 15th Floor",
					"Thumbnail": {
						"Url": "http://www.example.com/image/481989943",
						"Height": 125,
						"Width": 100
					},
					"Animated": false,
					"IDs": [116, 943, 234, 38793]
				}
			}`,
			expected: []string{"Image", "Width", "Height", "Title", "Thumbnail", "Url", "Animated", "IDs"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Unmarshal([]byte(tt.json))
			if err != nil {
				t.Errorf("error occurred while unmarshal")
			}

			value := root.UniqueKeys()
			if len(value) != len(tt.expected) {
				t.Errorf("%s length must be %v. got: %v. retrieved keys: %s", tt.name, len(tt.expected), len(value), value)
			}

			for _, key := range value {
				if !contains(tt.expected, key) {
					t.Errorf("EachKey() must be in %v. got: %v", tt.expected, key)
				}
			}
		})
	}
}

func TestNode_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		node     *Node
		expected bool
	}{
		{"nil node", (*Node)(nil), false}, // nil node is not empty.
		{"null node", NullNode(""), true},
		{"empty object", ObjectNode("", nil), true},
		{"empty array", ArrayNode("", nil), true},
		{"non-empty object", ObjectNode("", map[string]*Node{"foo": BoolNode("foo", true)}), false},
		{"non-empty array", ArrayNode("", []*Node{BoolNode("0", true)}), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.Empty(); got != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestNode_Index_EmptyList(t *testing.T) {
	root, err := Unmarshal([]byte(`[]`))
	if err != nil {
		t.Errorf("error occurred while unmarshal")
	}

	array := root.MustArray()
	for i, node := range array {
		if i != node.Index() {
			t.Errorf(fmt.Sprintf("Index() must be nil. got: %v", i))
		}
	}
}

func TestNode_GetObject(t *testing.T) {
	root, err := Unmarshal([]byte(`{"foo": true,"bar": null}`))
	if err != nil {
		t.Errorf("Error on Unmarshal(): %s", err.Error())
		return
	}

	value, err := root.GetObject()
	if err != nil {
		t.Errorf("Error on root.GetObject(): %s", err.Error())
	}

	if _, ok := value["foo"]; !ok {
		t.Errorf("root.GetObject() is corrupted: foo")
	}

	if _, ok := value["bar"]; !ok {
		t.Errorf("root.GetObject() is corrupted: bar")
	}
}

func TestNode_GetObject_Fail(t *testing.T) {
	tests := []simpleNode{
		{"nil node", (*Node)(nil)},
		{"get object from null node", NullNode("")},
		{"not object node", NumberNode("", 123)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.node.GetObject(); err == nil {
				t.Errorf("%s should be an error", tt.name)
			}
		})
	}
}

func TestNode_ObjectEach(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected map[string]int
	}{
		{
			name:     "empty object",
			json:     `{}`,
			expected: make(map[string]int),
		},
		{
			name:     "single key-value pair",
			json:     `{"key": 42}`,
			expected: map[string]int{"key": 42},
		},
		{
			name:     "multiple key-value pairs",
			json:     `{"one": 1, "two": 2, "three": 3}`,
			expected: map[string]int{"one": 1, "two": 2, "three": 3},
		},
		{
			name:     "multiple key-value pairs with some non-numeric values",
			json:     `{"one": 1, "two": "2", "three": 3, "four": "4"}`,
			expected: map[string]int{"one": 1, "three": 3},
		},
		{
			name:     "non-object node",
			json:     `["not", "an", "object"]`,
			expected: make(map[string]int),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root, err := Unmarshal([]byte(tc.json))
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			result := make(map[string]int)
			root.ObjectEach(func(key string, value *Node) {
				// extract integer values from the object
				if val, err := strconv.Atoi(value.String()); err == nil {
					result[key] = val
				}
			})

			if len(result) != len(tc.expected) {
				t.Errorf("%s: expected %d key-value pairs, got %d", tc.name, len(tc.expected), len(result))
				return
			}

			for key, val := range tc.expected {
				if result[key] != val {
					t.Errorf("%s: expected value for key %s to be %d, got %d", tc.name, key, val, result[key])
				}
			}
		})
	}
}

func TestNode_ExampleMust(t *testing.T) {
	data := []byte(`{
        "Image": {
            "Width":  800,
            "Height": 600,
            "Title":  "View from 15th Floor",
            "Thumbnail": {
                "Url":    "http://www.example.com/image/481989943",
                "Height": 125,
                "Width":  100
            },
            "Animated" : false,
            "IDs": [116, 943, 234, 38793]
        }
    }`)

	root := Must(Unmarshal(data))
	if root.Size() != 1 {
		t.Errorf("root.Size() must be 1. got: %v", root.Size())
	}

	fmt.Printf("Object has %d inheritors inside", root.Size())
	// Output:
	// Object has 1 inheritors inside
}

// Calculate AVG price from different types of objects, JSON from: https://goessner.net/articles/JsonPath/index.html#e3
func TestExampleUnmarshal(t *testing.T) {
	data := []byte(`{ "store": {
    "book": [ 
      { "category": "reference",
        "author": "Nigel Rees",
        "title": "Sayings of the Century",
        "price": 8.95
      },
      { "category": "fiction",
        "author": "Evelyn Waugh",
        "title": "Sword of Honour",
        "price": 12.99
      },
      { "category": "fiction",
        "author": "Herman Melville",
        "title": "Moby Dick",
        "isbn": "0-553-21311-3",
        "price": 8.99
      },
      { "category": "fiction",
        "author": "J. R. R. Tolkien",
        "title": "The Lord of the Rings",
        "isbn": "0-395-19395-8",
        "price": 22.99
      }
    ],
    "bicycle": { "color": "red",
      "price": 19.95
    },
    "tools": null
  }
}`)

	root, err := Unmarshal(data)
	if err != nil {
		t.Errorf("error occurred when unmarshal")
	}

	store := root.MustKey("store").MustObject()

	var prices float64
	size := 0
	for _, objects := range store {
		if objects.IsArray() && objects.Size() > 0 {
			size += objects.Size()
			for _, object := range objects.MustArray() {
				prices += object.MustKey("price").MustNumeric()
			}
		} else if objects.IsObject() && objects.HasKey("price") {
			size++
			prices += objects.MustKey("price").MustNumeric()
		}
	}

	result := int(prices / float64(size))
	fmt.Println("AVG price:", result)
}

func TestNode_ExampleMust_panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	data := []byte(`{]`)
	root := Must(Unmarshal(data))
	fmt.Printf("Object has %d inheritors inside", root.Size())
}

func compareNodes(n1, n2 *Node) bool {
	if n1 == nil || n2 == nil {
		return n1 == n2
	}

	if n1.key != n2.key {
		return false
	}

	if !bytes.Equal(n1.data, n2.data) {
		return false
	}

	if n1.index != n2.index {
		return false
	}

	if n1.borders != n2.borders {
		return false
	}

	if n1.modified != n2.modified {
		return false
	}

	if n1.nodeType != n2.nodeType {
		return false
	}

	if !compareNodes(n1.prev, n2.prev) {
		return false
	}

	if len(n1.next) != len(n2.next) {
		return false
	}

	for k, v := range n1.next {
		if !compareNodes(v, n2.next[k]) {
			return false
		}
	}

	return true
}

func TestNode_Path(t *testing.T) {
	data := []byte(`{
        "Image": {
            "Width":  800,
            "Height": 600,
            "Title":  "View from 15th Floor",
            "Thumbnail": {
                "Url":    "http://www.example.com/image/481989943",
                "Height": 125,
                "Width":  100
            },
            "Animated" : false,
            "IDs": [116, 943, 234, 38793]
          }
      }`)

	root, err := Unmarshal(data)
	if err != nil {
		t.Errorf("Error on Unmarshal(): %s", err.Error())
		return
	}

	if root.Path() != "$" {
		t.Errorf("Wrong root.Path()")
	}

	element := root.MustKey("Image").MustKey("Thumbnail").MustKey("Url")
	if element.Path() != "$['Image']['Thumbnail']['Url']" {
		t.Errorf("Wrong path found: %s", element.Path())
	}

	if (*Node)(nil).Path() != "" {
		t.Errorf("Wrong (nil).Path()")
	}
}

func TestNode_Path2(t *testing.T) {
	tests := []struct {
		name string
		node *Node
		want string
	}{
		{
			name: "Node with key",
			node: &Node{
				prev: &Node{},
				key:  func() *string { s := "key"; return &s }(),
			},
			want: "$['key']",
		},
		{
			name: "Node with index",
			node: &Node{
				prev:  &Node{},
				index: func() *int { i := 1; return &i }(),
			},
			want: "$[1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.Path(); got != tt.want {
				t.Errorf("Path() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNode_Root(t *testing.T) {
	root := &Node{}
	child := &Node{prev: root}
	grandChild := &Node{prev: child}

	tests := []struct {
		name string
		node *Node
		want *Node
	}{
		{
			name: "Root node",
			node: root,
			want: root,
		},
		{
			name: "Child node",
			node: child,
			want: root,
		},
		{
			name: "Grandchild node",
			node: grandChild,
			want: root,
		},
		{
			name: "Node is nil",
			node: nil,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.root(); got != tt.want {
				t.Errorf("root() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNode_DeleteIndex(t *testing.T) {
	tests := []struct {
		json     string
		expected string
		index    int
		fail     bool
	}{
		{`null`, ``, 0, true},
		{`1`, ``, 0, true},
		{`{}`, ``, 0, true},
		{`{"foo":"bar"}`, ``, 0, true},
		{`true`, ``, 0, true},
		{`[]`, ``, 0, true},
		{`[]`, ``, -1, true},
		{`[1]`, `[]`, 0, false},
		{`[{}]`, `[]`, 0, false},
		{`[{}]`, `[]`, -1, false},
		{`[{},[],1]`, `[{},[]]`, -1, false},
		{`[{},[],1]`, `[{},1]`, 1, false},
		{`[{},[],1]`, ``, 10, true},
		{`[{},[],1]`, ``, -10, true},
	}
	for _, test := range tests {
		t.Run(test.json, func(t *testing.T) {
			root := Must(Unmarshal([]byte(test.json)))
			err := root.DeleteIndex(test.index)
			if test.fail {
				if err == nil {
					t.Errorf("Expected error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				result, err := Marshal(root)
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if string(result) != test.expected {
					t.Errorf("Unexpected result: %s", result)
				}
			}
		})
	}
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}

	return false
}

func TestNode_update(t *testing.T) {
	type _args struct {
		nt  ValueType
		val interface{}
	}

	tests := []struct {
		name  string
		args  _args
		isErr bool
	}{
		{
			name:  "null ok",
			args:  _args{nt: Null, val: nil},
			isErr: false,
		},
		{
			name:  "null fail",
			args:  _args{nt: Null, val: "foo"},
			isErr: true,
		},
		{
			name:  "string ok",
			args:  _args{nt: String, val: "foo"},
			isErr: false,
		},
		{
			name:  "string fail",
			args:  _args{nt: String, val: nil},
			isErr: true,
		},
		{
			name:  "number ok int type",
			args:  _args{nt: Number, val: 42},
			isErr: false,
		},
		{
			name:  "number ok float type",
			args:  _args{nt: Number, val: 42.0},
			isErr: false,
		},
		{
			name:  "number fail",
			args:  _args{nt: Number, val: nil},
			isErr: true,
		},
		{
			name:  "bool ok",
			args:  _args{nt: Boolean, val: true},
			isErr: false,
		},
		{
			name:  "bool fail",
			args:  _args{nt: Boolean, val: nil},
			isErr: true,
		},
		{
			name:  "array ok",
			args:  _args{nt: Array, val: []*Node{}},
			isErr: false,
		},
		{
			name:  "array fail",
			args:  _args{nt: Array, val: nil},
			isErr: true,
		},
		{
			name:  "object ok",
			args:  _args{nt: Object, val: map[string]*Node{}},
			isErr: false,
		},
		{
			name:  "object type, fail",
			args:  _args{nt: Object, val: nil},
			isErr: true,
		},
		{
			name:  "object fail: type mismatch",
			args:  _args{nt: Object, val: 42.0},
			isErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NullNode("")
			if err := node.update(tt.args.nt, tt.args.val); (err != nil) != tt.isErr {
				t.Errorf("%s error = %v, wantErr %v", tt.name, err, tt.isErr)
			}
		})
	}
}

// ignore the sequence of keys by ordering them.
// need to avoid import encoding/json and reflect package.
// because gno does not support them for now.
// TODO: use encoding/json to compare the result after if possible in gno.
func isSameObject(a, b string) bool {
	aPairs := strings.Split(strings.Trim(a, "{}"), ",")
	bPairs := strings.Split(strings.Trim(b, "{}"), ",")

	aMap := make(map[string]string)
	bMap := make(map[string]string)
	for _, pair := range aPairs {
		kv := strings.Split(pair, ":")
		key := strings.Trim(kv[0], `"`)
		value := strings.Trim(kv[1], `"`)
		aMap[key] = value
	}
	for _, pair := range bPairs {
		kv := strings.Split(pair, ":")
		key := strings.Trim(kv[0], `"`)
		value := strings.Trim(kv[1], `"`)
		bMap[key] = value
	}

	aKeys := make([]string, 0, len(aMap))
	bKeys := make([]string, 0, len(bMap))
	for k := range aMap {
		aKeys = append(aKeys, k)
	}

	for k := range bMap {
		bKeys = append(bKeys, k)
	}

	sort.Strings(aKeys)
	sort.Strings(bKeys)

	if len(aKeys) != len(bKeys) {
		return false
	}

	for i := range aKeys {
		if aKeys[i] != bKeys[i] {
			return false
		}

		if aMap[aKeys[i]] != bMap[bKeys[i]] {
			return false
		}
	}

	return true
}

func TestNode_Clone(t *testing.T) {
	node := NumberNode("", 1.1)
	null := NullNode("")
	array := ArrayNode("", []*Node{node, null})
	object := ObjectNode("", map[string]*Node{"array": array})

	tests := []struct {
		name string
		node *Node
		json string
	}{
		{
			name: "null",
			node: null,
			json: "null",
		},
		{
			name: "node",
			node: node,
			json: "1.1",
		},
		{
			name: "array",
			node: array,
			json: "[1.1,null]",
		},
		{
			name: "object",
			node: object,
			json: `{"array":[1.1,null]}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clone := test.node.Clone()
			if clone.prev != nil {
				t.Error("Clone().parent != nil")
			} else if clone.index != nil {
				t.Error("Clone().index != nil")
			} else if clone.key != nil {
				t.Error("Clone().key != nil")
			}

			if result, err := Marshal(clone); err != nil {
				t.Errorf("Marshal() error: %s", err)
			} else if string(result) != test.json {
				t.Errorf("Marshal() clone not match: \nExpected: %s\nActual: %s", test.json, result)
			} else if base, err := Marshal(test.node); err != nil {
				t.Errorf("Marshal() error: %s", err)
			} else if string(base) != test.json {
				t.Errorf("Marshal() base not match: \nExpected: %s\nActual: %s", test.json, base)
			}
		})
	}
}
