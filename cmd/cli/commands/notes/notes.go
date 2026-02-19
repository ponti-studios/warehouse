package notes

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
)

type NotesResolver struct{}

func (r *NotesResolver) Notes(ctx context.Context) ([]string, error) {
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
	flagSet.Parse(os.Args[2:])

	if *filePath == "" {
		fmt.Println("Usage: notes -file <path/to/markdown/file.md>")
		return fmt.Errorf("missing required -file flag")
	}

	data, err := os.ReadFile(*filePath)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "- ") {
			fmt.Println(strings.TrimPrefix(line, "- "))
		}
	}

	return nil
}
