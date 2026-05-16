package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/acidghost/a555less/internal/jsondoc"
	"github.com/acidghost/a555less/internal/tui"
)

var (
	buildVersion string
	buildCommit  string
	buildDate    string
)

//go:embed banner.txt
var banner string

var (
	errNoFile      = errors.New("missing input file; pass a JSON file or pipe JSON on stdin")
	errTooManyArgs = errors.New("too many arguments")
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
			printUsage()
			return nil
		case "--version":
			fmt.Printf("Version: %s\nCommit:  %s\nDate:    %s\n", buildVersion, buildCommit, buildDate)
			return nil
		}
	}

	filename, data, err := readInput(args)
	if err != nil {
		if errors.Is(err, errNoFile) || errors.Is(err, errTooManyArgs) {
			printUsage()
		}
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
			return "", nil, errNoFile
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
		return "", nil, errTooManyArgs
	}
}

func stdinIsTerminal() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func printUsage() {
	bannerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		BorderForeground(tui.ColorViolet).
		Foreground(tui.ColorPink).
		Padding(3, 3, 1, 3).
		Margin(1, 3, 0, 3)
	progStyle := lipgloss.NewStyle().Foreground(tui.ColorPurple).Bold(true)
	argsStyle := lipgloss.NewStyle().Foreground(tui.ColorPink).Underline(true)

	prog := filepath.Base(os.Args[0])
	usage := fmt.Sprintf("%s\n\n%s %s\n%s",
		bannerStyle.Render(banner),
		progStyle.Render(prog),
		argsStyle.Render("[FILE]"),
		`
    -h, --help     Show this help
        --version  Show version
`,
	)
	fmt.Println(lipgloss.NewStyle().Margin(0, 1).Render(usage))
}
