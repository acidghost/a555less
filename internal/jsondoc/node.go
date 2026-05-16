package jsondoc

import "encoding/json"

// Kind identifies the JSON value kind stored in a Node.
type Kind int

const (
	KindNull Kind = iota
	KindBool
	KindNumber
	KindString
	KindObject
	KindArray
)

// Node is an ordered JSON AST node.
type Node struct {
	ID int

	Kind Kind

	Parent *Node
	Key    string
	HasKey bool
	Index  int // child index within object or array

	Children []*Node

	// Primitive values.
	String string
	Number json.Number
	Bool   bool

	// UI state.
	Collapsed bool
}

// IsContainer reports whether n is an object or array.
func (n *Node) IsContainer() bool {
	return n != nil && (n.Kind == KindObject || n.Kind == KindArray)
}

// Document is the parsed JSON document passed to the TUI.
type Document struct {
	Root     *Node
	Filename string
	Data     []byte
	JSONL    bool
	nextID   int
}

// Size returns the input size in bytes.
func (d *Document) Size() int {
	if d == nil {
		return 0
	}
	return len(d.Data)
}

func (d *Document) newNode(kind Kind, parent *Node, key string, hasKey bool, index int) *Node {
	n := &Node{
		ID:     d.nextID,
		Kind:   kind,
		Parent: parent,
		Key:    key,
		HasKey: hasKey,
		Index:  index,
	}
	d.nextID++
	return n
}
