package jsondoc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// ErrTrailingToken reports that a JSON input contains more than one top-level
// value. Callers may use this to retry parsing as JSON Lines.
var ErrTrailingToken = errors.New("unexpected trailing token")

// Parse parses data as a single strict JSON value and returns an ordered AST.
func Parse(data []byte, filename string) (*Document, error) {
	doc := &Document{
		Filename: filename,
		Data:     append([]byte(nil), data...),
	}

	root, err := doc.parseSingleValue(data, nil, "", false, 0)
	if err != nil {
		return nil, err
	}
	doc.Root = root

	return doc, nil
}

// ParseJSONL parses data as JSON Lines and returns an ordered AST whose root is
// an array containing one child per non-empty line.
func ParseJSONL(data []byte, filename string) (*Document, error) {
	doc := &Document{
		Filename: filename,
		Data:     append([]byte(nil), data...),
		JSONL:    true,
	}
	doc.Root = doc.newNode(KindArray, nil, "", false, 0)

	for lineIndex, line := range bytes.Split(data, []byte{'\n'}) {
		line = bytes.TrimSpace(bytes.TrimSuffix(line, []byte{'\r'}))
		if len(line) == 0 {
			continue
		}

		child, err := doc.parseSingleValue(line, doc.Root, "", false, len(doc.Root.Children))
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineIndex+1, err)
		}
		doc.Root.Children = append(doc.Root.Children, child)
	}

	return doc, nil
}

func (d *Document) parseSingleValue(data []byte, parent *Node, key string, hasKey bool, index int) (*Node, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	root, err := d.parseValue(dec, parent, key, hasKey, index)
	if err != nil {
		return nil, invalidJSONError(data, err)
	}

	// Ensure no trailing non-whitespace tokens.
	if tok, err := dec.Token(); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("invalid JSON: %w %v", ErrTrailingToken, tok)
		}
		return nil, invalidJSONError(data, err)
	}

	return root, nil
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

func invalidJSONError(data []byte, err error) error {
	if errors.Is(err, io.EOF) {
		return fmt.Errorf("invalid JSON: unexpected end of input")
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		line, col := lineColumn(data, syntaxErr.Offset)
		return fmt.Errorf("invalid JSON at line %d, column %d: %w", line, col, err)
	}

	return fmt.Errorf("invalid JSON: %w", err)
}

func lineColumn(data []byte, offset int64) (int, int) {
	if offset < 1 {
		return 1, 1
	}

	line := 1
	col := 1
	for i, b := range data {
		if int64(i) >= offset-1 {
			break
		}
		if b == '\n' {
			line++
			col = 1
			continue
		}
		col++
	}
	return line, col
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
