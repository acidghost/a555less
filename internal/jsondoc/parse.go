package jsondoc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Document is the parsed JSON document passed to the TUI.
//
// Phase 1 only validates and stores the input bytes. The ordered AST is added
// in the next phase.
type Document struct {
	Filename string
	Data     []byte
}

// Parse validates data as a single strict JSON value and returns a document.
func Parse(data []byte, filename string) (*Document, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	var value any
	if err := dec.Decode(&value); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	var trailing any
	if err := dec.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("invalid JSON: unexpected trailing value")
		}
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return &Document{
		Filename: filename,
		Data:     append([]byte(nil), data...),
	}, nil
}

// Size returns the input size in bytes.
func (d *Document) Size() int {
	if d == nil {
		return 0
	}
	return len(d.Data)
}
