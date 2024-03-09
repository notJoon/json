package json

import (
	"bytes"
	"fmt"
	"strconv"
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

func TestNode_MustString(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"foo", []byte(`"foo"`)},
		{"foo bar", []byte(`"foo bar"`)},
		{"", []byte(`""`)},
		{"こんにちは", []byte(`"こんにちは"`)},
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

func TestNode_Size(t *testing.T) {
	root, err := Unmarshal(sampleArr)
	if err != nil {
		t.Errorf("error occured while unmarshaling")
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
		t.Error("error occured while unmarshaling")
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

func TestNode_GetKey(t *testing.T) {
	root, err := Unmarshal([]byte(`{"foo": true, "bar": null}`))
	if err != nil {
		t.Error("error occured while unmarshaling")
	}

	value, err := root.GetKey("foo")
	if err != nil {
		t.Errorf("error occured while getting key, %s", err)
	}

	if value.MustBool() != true {
		t.Errorf("value is not matched. expected: true, got: %v", value.MustBool())
	}

	value, err = root.GetKey("bar")
	if err != nil {
		t.Errorf("error occured while getting key, %s", err)
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
		// {
		// 	name: "complex object",
		// 	json: `{
		// 		"Image": {
		// 			"Width":  800,
		// 			"Height": 600,
		// 			"Title":  "View from 15th Floor",
		// 			"Thumbnail": {
		// 				"Url":    "http://www.example.com/image/481989943",
		// 				"Height": 125,
		// 				"Width":  100
		// 			},
		// 			"Animated" : false,
		// 			"IDs": [116, 943, 234, 38793]
		// 		}
		// 	}`,
		// 	expected: []string{"Image", "Width", "Height", "Title", "Thumbnail", "Url", "Animated", "IDs"},
		// },
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root, err := Unmarshal([]byte(tc.json))
			if err != nil {
				t.Errorf("error occured while unmarshaling")
			}

			value := root.EachKey()
			if len(value) != len(tc.expected) {
				t.Errorf("EachKey() must be %v. got: %v", len(tc.expected), len(value))
			}

			for _, key := range value {
				if !contains(tc.expected, key) {
					t.Errorf("EachKey() must be in %v. got: %v", tc.expected, key)
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
		t.Errorf("error occured while unmarshaling")
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

func TestNodeString(t *testing.T) {
	rel := &dummyKey
	node := &Node{
		key:      rel,
		nodeType: String,
		index:    new(int),
		borders:  [2]int{0, 1},
		modified: true,
	}

	expected := "Node{key: key, nodeType: string, index: 0, borders: [0, 1], modified: true}"
	if str := node.String(); str != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, str)
	}
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

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}
