package apple

import (
	"context"
	"flag"
	"fmt"
	"os"

	"gogogo/internal/application/apple"
	"gogogo/internal/infrastructure/config"
	"gogogo/internal/infrastructure/persistence/sqlite"
)

type ImportCommand struct {
	DBPath    string
	SourceDir string
	DryRun    bool
	Force     bool
}

func (c *ImportCommand) Execute(ctx context.Context) error {
	if c.DBPath == "" {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		c.DBPath = cfg.Database.Path
	}

	if c.SourceDir == "" {
		fmt.Fprintln(os.Stderr, "Error: Source directory is required")
		return fmt.Errorf("source directory is required")
	}

	conn, err := sqlite.NewConnection(c.DBPath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close()

	repo := sqlite.NewAppleRepository(conn.DB())
	service := apple.NewService(repo)

	options := apple.ImportOptions{
		DryRun: c.DryRun,
		Force:  c.Force,
	}

	_, err = service.ImportAll(ctx, c.SourceDir, options)
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	return nil
}

func HandleAppleImport() int {
	fs := flag.NewFlagSet("apple-import", flag.ExitOnError)
	dbPath := fs.String("db", "", "Path to SQLite database")
	sourceDir := fs.String("source", "", "Source directory containing Apple data")
	dryRun := fs.Bool("dry-run", false, "Validate without importing")
	force := fs.Bool("force", false, "Skip duplicate checking")

	fs.Parse(os.Args[2:])

	cmd := ImportCommand{
		DBPath:    *dbPath,
		SourceDir: *sourceDir,
		DryRun:    *dryRun,
		Force:     *force,
	}

	ctx := context.Background()

	if err := cmd.Execute(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	return 0
}
