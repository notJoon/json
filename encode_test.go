package json

import (
	"testing"
)

func TestMarshal_Primitive(t *testing.T) {
	tests := []struct {
		name string
		node *Node
	}{
		{
			name: "null",
			node: NullNode(""),
		},
		{
			name: "true",
			node: BoolNode("", true),
		},
		{
			name: "false",
			node: BoolNode("", false),
		},
		{
			name: `"string"`,
			node: StringNode("", "string"),
		},
		{
			name: `"one \"encoded\" string"`,
			node: StringNode("", `one "encoded" string`),
		},
		{
			name: "100500",
			node: NumberNode("", 100500),
		},
		{
			name: "100.5",
			node: NumberNode("", 100.5),
		},
		{
			name: `[1,2,3]`,
			node: ArrayNode("", []*Node{
				NumberNode("0", 1),
				NumberNode("2", 2),
				NumberNode("3", 3),
			}),
		},
		{
			name: `{"foo":"bar"}`,
			node: ObjectNode("", map[string]*Node{
				"foo": StringNode("foo", "bar"),
			}),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := Marshal(test.node)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if string(value) != test.name {
				t.Errorf("wrong result: '%s', expected '%s'", value, test.name)
			}
		})
	}
}

func TestMarshal_Errors(t *testing.T) {
	tests := []struct {
		name string
		node func() (node *Node)
	}{
		{
			name: "nil",
			node: func() (node *Node) {
				return
			},
		},
		{
			name: "broken",
			node: func() (node *Node) {
				node = Must(Unmarshal([]byte(`{}`)))
				node.borders[1] = 0
				return
			},
		},
		{
			name: "Numeric",
			node: func() (node *Node) {
				return valueNode(nil, "", Number, false)
			},
		},
		{
			name: "String",
			node: func() (node *Node) {
				return valueNode(nil, "", String, false)
			},
		},
		{
			name: "Bool",
			node: func() (node *Node) {
				return valueNode(nil, "", Boolean, 1)
			},
		},
		{
			name: "Array_1",
			node: func() (node *Node) {
				node = ArrayNode("", nil)
				node.next["1"] = NullNode("1")
				return
			},
		},
		{
			name: "Array_2",
			node: func() (node *Node) {
				return ArrayNode("", []*Node{valueNode(nil, "", Boolean, 1)})
			},
		},
		{
			name: "Object",
			node: func() (node *Node) {
				return ObjectNode("", map[string]*Node{"key": valueNode(nil, "key", Boolean, 1)})
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := Marshal(test.node())
			if err == nil {
				t.Errorf("expected error")
			} else if len(value) != 0 {
				t.Errorf("wrong result")
			}
		})
	}
}
