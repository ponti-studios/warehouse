package frontmatter

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// WalkOptions controls how file traversal behaves.
type WalkOptions struct {
	IncludeHidden bool
	Extensions    []string
	IncludeGlobs  []string
	ExcludeGlobs  []string
	MaxFiles      int
}

// FileAction represents a validation or migration result for a single file.
type FileAction struct {
	Path       string
	HasChanges bool
	Errors     []error
	Result     MigrationResult
}

// WalkMarkdownFiles collects markdown files under root respecting options.
func WalkMarkdownFiles(root string, options WalkOptions) ([]string, error) {
	if len(options.Extensions) == 0 {
		options.Extensions = []string{".md", ".markdown"}
	}

	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if !options.IncludeHidden && strings.HasPrefix(entry.Name(), ".") {
				if path == root {
					return nil
				}
				return filepath.SkipDir
			}
			return nil
		}

		if !options.IncludeHidden && strings.HasPrefix(entry.Name(), ".") {
			return nil
		}

		if !hasExtension(path, options.Extensions) {
			return nil
		}

		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			rel = path
		}

		if !matchesGlobs(rel, options.IncludeGlobs, true) {
			return nil
		}

		if matchesGlobs(rel, options.ExcludeGlobs, false) {
			return nil
		}

		files = append(files, path)
		if options.MaxFiles > 0 && len(files) >= options.MaxFiles {
			return fs.SkipAll
		}
		return nil
	})

	return files, err
}

// ValidateFiles validates frontmatter for all markdown files under root.
func ValidateFiles(root string, schema SchemaDefinition, options WalkOptions) ([]FileAction, error) {
	files, err := WalkMarkdownFiles(root, options)
	if err != nil {
		return nil, err
	}

	actions := make([]FileAction, 0, len(files))
	for _, path := range files {
		content, readErr := os.ReadFile(path)
		action := FileAction{Path: path}

		if readErr != nil {
			action.Errors = append(action.Errors, fmt.Errorf("read error: %w", readErr))
			actions = append(actions, action)
			continue
		}

		result, err := ValidateContent(string(content), schema)
		if err != nil {
			action.Errors = append(action.Errors, err)
			actions = append(actions, action)
			continue
		}

		action.Result = MigrationResult{
			HasFrontmatter:   true,
			ValidationBefore: result,
			ValidationAfter:  result,
		}

		if result.HasErrors() {
			for _, v := range result.Errors {
				action.Errors = append(action.Errors, v)
			}
		}

		actions = append(actions, action)
	}

	return actions, nil
}

// MigrateFiles migrates frontmatter for markdown files under root.
func MigrateFiles(root string, schema SchemaDefinition, strategy MigrationStrategy, options WalkOptions, writeChanges bool, backup bool, registry *GeneratorRegistry, now time.Time) ([]FileAction, error) {
	files, err := WalkMarkdownFiles(root, options)
	if err != nil {
		return nil, err
	}

	actions := make([]FileAction, 0, len(files))
	for _, path := range files {
		content, readErr := os.ReadFile(path)
		action := FileAction{Path: path}

		if readErr != nil {
			action.Errors = append(action.Errors, fmt.Errorf("read error: %w", readErr))
			actions = append(actions, action)
			continue
		}

		updated, result, err := MigrateContent(string(content), path, schema, strategy, registry, now)
		if err != nil {
			action.Errors = append(action.Errors, err)
			actions = append(actions, action)
			continue
		}

		action.Result = result
		action.HasChanges = result.HasChanges

		if result.ValidationAfter.HasErrors() {
			for _, v := range result.ValidationAfter.Errors {
				action.Errors = append(action.Errors, v)
			}
		}

		if result.HasChanges && writeChanges {
			if writeErr := WriteFileWithBackup(path, []byte(updated), 0644, backup); writeErr != nil {
				action.Errors = append(action.Errors, fmt.Errorf("write error: %w", writeErr))
			}
		}

		actions = append(actions, action)
	}

	return actions, nil
}

func hasExtension(path string, extensions []string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, allowed := range extensions {
		if strings.ToLower(allowed) == ext {
			return true
		}
	}
	return false
}

func matchesGlobs(path string, patterns []string, defaultWhenEmpty bool) bool {
	if len(patterns) == 0 {
		return defaultWhenEmpty
	}
	for _, pattern := range patterns {
		if ok, _ := filepath.Match(pattern, path); ok {
			return true
		}
	}
	return false
}
