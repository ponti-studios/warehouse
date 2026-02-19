package tracking

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gogogo/internal/domain/tracking"
	"gogogo/internal/infrastructure/persistence/sqlite"
)

// ImportCommand handles importing tracking data from markdown files
type ImportCommand struct {
	DBPath    string
	SourceDir string
	DryRun    bool
	Verbose   bool
}

// Execute runs the tracking import command
func (c *ImportCommand) Execute(ctx context.Context) error {
	if c.Verbose {
		fmt.Printf("Starting tracking import from: %s\n", c.SourceDir)
		if c.DryRun {
			fmt.Println("DRY RUN MODE - no data will be written to database")
		}
		fmt.Println()
	}

	// Validate source directory
	if _, err := os.Stat(c.SourceDir); os.IsNotExist(err) {
		return fmt.Errorf("source directory does not exist: %s", c.SourceDir)
	}

	// Connect to database
	conn, err := sqlite.NewConnection(c.DBPath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close()

	// Find all markdown files
	markdownFiles, err := c.findMarkdownFiles()
	if err != nil {
		return fmt.Errorf("failed to find markdown files: %w", err)
	}

	if len(markdownFiles) == 0 {
		fmt.Println("No markdown files found in source directory")
		return nil
	}

	fmt.Printf("Found %d markdown files to process\n", len(markdownFiles))

	// Parse files
	parser := tracking.NewMarkdownParser()
	var allEntries []tracking.TrackingEntry
	var parseErrors []string

	for filePath, content := range markdownFiles {
		if c.Verbose {
			fmt.Printf("Parsing: %s\n", filePath)
		}

		entry, err := parser.ParseFile(filePath, content)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("%s: %v", filePath, err))
			continue
		}

		allEntries = append(allEntries, *entry)
	}

	// Report parsing results
	fmt.Printf("\nParsing Results:\n")
	fmt.Printf("  Successfully parsed: %d entries\n", len(allEntries))
	fmt.Printf("  Parse errors: %d\n", len(parseErrors))

	if len(parseErrors) > 0 && c.Verbose {
		fmt.Printf("\nParse Errors:\n")
		for _, errMsg := range parseErrors {
			fmt.Printf("  - %s\n", errMsg)
		}
	}

	if len(allEntries) == 0 {
		return fmt.Errorf("no entries could be parsed from markdown files")
	}

	// Show summary statistics
	stats := parser.ExtractSummaryStats(allEntries)
	fmt.Printf("\nImport Summary:\n")
	fmt.Printf("  Total entries: %v\n", stats["total_entries"])
	fmt.Printf("  Total data points: %v\n", stats["total_data_points"])

	if types, ok := stats["types"].(map[string]int); ok {
		fmt.Printf("  Entry types:\n")
		for entryType, count := range types {
			fmt.Printf("    %s: %d entries\n", entryType, count)
		}
	}

	if dateRange, ok := stats["date_range"].(map[string]string); ok {
		fmt.Printf("  Date range: %s to %s\n", dateRange["earliest"], dateRange["latest"])
	}

	// If dry run, stop here
	if c.DryRun {
		fmt.Printf("\nDRY RUN COMPLETE - No data was written to database\n")
		return nil
	}

	// Import entries to database
	fmt.Printf("\nImporting entries to database...\n")

	var importErrors []string
	successCount := 0

	for i, entry := range allEntries {
		if c.Verbose {
			fmt.Printf("  Importing %d/%d: %s (%s)\n", i+1, len(allEntries), entry.Title, entry.Type)
		}

		// Here we would use a proper tracking service/repository
		// For now, we'll create basic entities in the universal entities table
		if err := c.importEntry(ctx, conn, entry); err != nil {
			importErrors = append(importErrors, fmt.Sprintf("%s: %v", entry.Title, err))
			continue
		}
		successCount++
	}

	// Report import results
	fmt.Printf("\nImport Results:\n")
	fmt.Printf("  Successfully imported: %d entries\n", successCount)
	fmt.Printf("  Import errors: %d\n", len(importErrors))

	if len(importErrors) > 0 && c.Verbose {
		fmt.Printf("\nImport Errors:\n")
		for _, errMsg := range importErrors {
			fmt.Printf("  - %s\n", errMsg)
		}
	}

	fmt.Printf("\n✅ Import completed successfully!\n")
	fmt.Printf("You can now search your tracking data with: hominem search <query> --domains tracking\n")

	return nil
}

// findMarkdownFiles recursively finds all .md files in the source directory
func (c *ImportCommand) findMarkdownFiles() (map[string]string, error) {
	files := make(map[string]string)

	err := filepath.WalkDir(c.SourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}

		// Skip hidden files and directories
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			if c.Verbose {
				fmt.Printf("Warning: could not read file %s: %v\n", path, err)
			}
			return nil // Continue with other files
		}

		// Use relative path as key
		relPath, err := filepath.Rel(c.SourceDir, path)
		if err != nil {
			relPath = path
		}

		files[relPath] = string(content)
		return nil
	})

	return files, err
}

// importEntry imports a single tracking entry to the database
func (c *ImportCommand) importEntry(ctx context.Context, conn *sqlite.Connection, entry tracking.TrackingEntry) error {
	// Get metadata as JSON string
	metadataJSON, err := entry.GetMetadataJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize metadata: %w", err)
	}

	// Insert into tracking_entries table
	_, err = conn.DB().ExecContext(ctx, `
		INSERT OR REPLACE INTO tracking_entries (
			id, type, title, content, metadata, date, status, source_file, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		entry.ID,
		entry.Type,
		entry.Title,
		entry.Content,
		metadataJSON,
		entry.Date.String(),
		entry.Status,
		entry.SourceFile,
		entry.CreatedAt.String(),
		entry.UpdatedAt.String(),
	)

	if err != nil {
		return fmt.Errorf("failed to insert tracking entry: %w", err)
	}

	// Insert data points
	for _, point := range entry.DataPoints {
		_, err = conn.DB().ExecContext(ctx, `
			INSERT OR REPLACE INTO tracking_data_points (
				id, entry_id, date, type, value, unit, metadata, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`,
			point.ID,
			point.EntryID,
			point.Date.String(),
			point.Type,
			point.Value,
			point.Unit,
			point.Metadata,
			point.CreatedAt.String(),
		)

		if err != nil {
			return fmt.Errorf("failed to insert data point: %w", err)
		}
	}

	return nil
}

// ValidateImport performs validation checks before import
func (c *ImportCommand) ValidateImport() error {
	// Check source directory exists and is readable
	info, err := os.Stat(c.SourceDir)
	if err != nil {
		return fmt.Errorf("cannot access source directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("source path is not a directory: %s", c.SourceDir)
	}

	// Check database is accessible (if not dry run)
	if !c.DryRun {
		if _, err := os.Stat(c.DBPath); err != nil {
			return fmt.Errorf("cannot access database: %w", err)
		}
	}

	return nil
}
