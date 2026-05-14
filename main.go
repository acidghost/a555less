package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/acidghost/a555less/internal/jsondoc"
	"github.com/acidghost/a555less/internal/tui"
)

var (
	buildVersion string
	buildCommit  string
	buildDate    string
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "a555less: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 1 {
		switch args[0] {
		case "-h", "--help":
			printUsage(os.Stdout)
			return nil
		case "--version":
			fmt.Printf("Version: %s\nCommit:  %s\nDate:    %s\n", buildVersion, buildCommit, buildDate)
			return nil
		}
	}

	filename, data, err := readInput(args)
	if err != nil {
		return err
	}

	doc, err := jsondoc.Parse(data, filename)
	if err != nil {
		return fmt.Errorf("parse %s: %w", filename, err)
	}

	program := tea.NewProgram(tui.New(doc))
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}

	return nil
}

func readInput(args []string) (string, []byte, error) {
	switch len(args) {
	case 0:
		if stdinIsTerminal() {
			return "", nil, errors.New("missing input file; pass a JSON file or pipe JSON on stdin")
		}
		data, err := io.ReadAll(os.Stdin)
		return "stdin", data, err
	case 1:
		if args[0] == "-" {
			data, err := io.ReadAll(os.Stdin)
			return "stdin", data, err
		}
		// #nosec G703 -- this CLI intentionally reads the user-specified JSON path.
		data, err := os.ReadFile(args[0])
		return args[0], data, err
	default:
		return "", nil, errors.New("too many arguments")
	}
}

func stdinIsTerminal() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func printUsage(out io.Writer) {
	fmt.Fprintln(out, "Usage: a555less [FILE|-]")
	fmt.Fprintln(out, "Read JSON from FILE, or stdin when FILE is '-' or omitted.")
}
