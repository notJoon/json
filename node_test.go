package json

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"
)

var nilKey *string

type _args struct {
	prev *Node
	buf  *buffer
	typ  ValueType
	key  **string
}

func TestNode_CreateNewNode(t *testing.T) {
	dummyKey := "key"
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
				buf:  NewBuffer(make([]byte, 10)),
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
				buf:  NewBuffer(make([]byte, 10)),
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

	root = NullNode("")
	if _, err = root.GetBool(); err == nil {
		t.Errorf("Error on root.GetBool(): NullNode")
	}

	if _, err := (*Node)(nil).GetBool(); err == nil {
		t.Errorf("(nil).GetBool() should be an error")
	}
}

func TestNode_GetNumeric(t *testing.T) {
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

	root = NullNode("")
	_, err = root.GetObject()
	if err == nil {
		t.Errorf("Error on root.GetArray(): NullNode")
	}

	if _, err := (*Node)(nil).GetObject(); err == nil {
		t.Errorf("(nil).GetObject() should be an error")
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
