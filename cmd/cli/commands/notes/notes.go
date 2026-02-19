package notes

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
)

type NotesResolver struct{}

func (r *NotesResolver) Notes() ([]string, error) {
	data, err := os.ReadFile("path/to/your/markdown/file.md")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	var notes []string
	for _, line := range lines {
		if strings.HasPrefix(line, "- ") {
			notes = append(notes, strings.TrimPrefix(line, "- "))
		}
	}

	return notes, nil
}

func Run() error {
	flagSet := flag.NewFlagSet("notes", flag.ExitOnError)
	filePath := flagSet.String("file", "", "Path to markdown file")
	render := flagSet.Bool("render", false, "Render markdown with glamour")
	flagSet.Parse(os.Args[2:])

	if *filePath == "" {
		fmt.Println("Usage: notes --file <path/to/markdown/file.md> [--render]")
		return fmt.Errorf("missing required --file flag")
	}

	data, err := os.ReadFile(*filePath)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	content := string(data)

	if *render {
		out, err := glamour.Render(content, "auto")
		if err != nil {
			return fmt.Errorf("rendering markdown: %w", err)
		}
		fmt.Print(out)
	} else {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "- ") {
				fmt.Println(strings.TrimPrefix(line, "- "))
			}
		}
	}

	return nil
}
