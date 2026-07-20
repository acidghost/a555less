package tui

import (
	"errors"
	"fmt"
	"io"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/term"

	"github.com/acidghost/a555less/internal/jsondoc"
)

type printTarget byte

const (
	printPrettyValue printTarget = iota
	printString
	printQueryPath
)

var errPrintNotString = errors.New("current value is not a string")

type printFinishedMsg struct {
	err error
}

type terminalPrintCommand struct {
	content string
	stdin   io.Reader
	stdout  io.Writer
	stderr  io.Writer
}

func (c *terminalPrintCommand) SetStdin(stdin io.Reader)   { c.stdin = stdin }
func (c *terminalPrintCommand) SetStdout(stdout io.Writer) { c.stdout = stdout }
func (c *terminalPrintCommand) SetStderr(stderr io.Writer) { c.stderr = stderr }

// Run writes directly to the terminal instead of returning content through the
// Bubble Tea renderer, which would clip it to the visible viewport.
func (c *terminalPrintCommand) Run() (runErr error) {
	if c.stdin == nil || c.stdout == nil {
		return errors.New("print command has no terminal input or output")
	}

	if _, err := fmt.Fprintf(c.stdout, "\x1b[2J\x1b[H%s\n\nPress any key to continue.", c.content); err != nil {
		return fmt.Errorf("printing content: %w", err)
	}

	if input, ok := c.stdin.(interface{ Fd() uintptr }); ok && term.IsTerminal(input.Fd()) {
		state, err := term.MakeRaw(input.Fd())
		if err != nil {
			return fmt.Errorf("entering raw mode: %w", err)
		}
		defer func() {
			if err := term.Restore(input.Fd(), state); err != nil && runErr == nil {
				runErr = fmt.Errorf("restoring terminal: %w", err)
			}
		}()
	}

	var input [32]byte
	if _, err := c.stdin.Read(input[:]); err != nil {
		return fmt.Errorf("waiting for key press: %w", err)
	}
	return nil
}

func printContent(doc *jsondoc.Document, n *jsondoc.Node, target printTarget) (string, error) {
	switch target {
	case printPrettyValue:
		return jsondoc.PrettyValue(n), nil
	case printString:
		if n == nil || n.Kind != jsondoc.KindString {
			return "", errPrintNotString
		}
		return jsondoc.PrintableStringContents(n.String), nil
	case printQueryPath:
		return jsondoc.QueryPath(doc, n), nil
	default:
		return "", errors.New("unknown print target")
	}
}

func printTargetForKey(key string) (printTarget, bool) {
	switch key {
	case "p":
		return printPrettyValue, true
	case "s":
		return printString, true
	case "q":
		return printQueryPath, true
	default:
		return 0, false
	}
}

func printTerminal(content string) tea.Cmd {
	return tea.Exec(&terminalPrintCommand{content: content}, func(err error) tea.Msg {
		return printFinishedMsg{err: err}
	})
}
