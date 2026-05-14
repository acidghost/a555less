package jsondoc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Parse parses data as a single strict JSON value and returns an ordered AST.
func Parse(data []byte, filename string) (*Document, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	doc := &Document{
		Filename: filename,
		Data:     append([]byte(nil), data...),
	}

	root, err := doc.parseValue(dec, nil, "", false, 0)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	doc.Root = root

	// Ensure no trailing non-whitespace tokens.
	if tok, err := dec.Token(); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("invalid JSON: unexpected trailing token %v", tok)
		}
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return doc, nil
}

func (d *Document) parseValue(dec *json.Decoder, parent *Node, key string, hasKey bool, index int) (*Node, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}

	switch tok := tok.(type) {
	case json.Delim:
		switch tok {
		case '{':
			return d.parseObject(dec, parent, key, hasKey, index)
		case '[':
			return d.parseArray(dec, parent, key, hasKey, index)
		default:
			return nil, fmt.Errorf("unexpected delimiter %q", tok)
		}
	case string:
		n := d.newNode(KindString, parent, key, hasKey, index)
		n.String = tok
		return n, nil
	case json.Number:
		n := d.newNode(KindNumber, parent, key, hasKey, index)
		n.Number = tok
		return n, nil
	case bool:
		n := d.newNode(KindBool, parent, key, hasKey, index)
		n.Bool = tok
		return n, nil
	case nil:
		return d.newNode(KindNull, parent, key, hasKey, index), nil
	default:
		return nil, fmt.Errorf("unexpected token %v", tok)
	}
}

func (d *Document) parseObject(dec *json.Decoder, parent *Node, key string, hasKey bool, index int) (*Node, error) {
	n := d.newNode(KindObject, parent, key, hasKey, index)

	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		childKey, ok := keyTok.(string)
		if !ok {
			return nil, fmt.Errorf("expected object key, got %v", keyTok)
		}

		child, err := d.parseValue(dec, n, childKey, true, len(n.Children))
		if err != nil {
			return nil, err
		}
		n.Children = append(n.Children, child)
	}

	if err := expectDelim(dec, '}'); err != nil {
		return nil, err
	}
	return n, nil
}

func (d *Document) parseArray(dec *json.Decoder, parent *Node, key string, hasKey bool, index int) (*Node, error) {
	n := d.newNode(KindArray, parent, key, hasKey, index)

	for dec.More() {
		child, err := d.parseValue(dec, n, "", false, len(n.Children))
		if err != nil {
			return nil, err
		}
		n.Children = append(n.Children, child)
	}

	if err := expectDelim(dec, ']'); err != nil {
		return nil, err
	}
	return n, nil
}

func expectDelim(dec *json.Decoder, want json.Delim) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != want {
		return fmt.Errorf("expected delimiter %q, got %v", want, tok)
	}
	return nil
}
