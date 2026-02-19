package health

import (
	"context"
	"flag"
	"fmt"
	"os"

	"gogogo/internal/application/health"
	"gogogo/internal/infrastructure/config"
	"gogogo/internal/infrastructure/persistence/sqlite"
)

type ImportCommand struct {
	DBPath    string
	SourceDir string
	DryRun    bool
	Force     bool
	Source    string
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

	repo := sqlite.NewHealthRepository(conn.DB())
	service := health.NewService(repo)

	options := health.ImportOptions{
		DryRun: c.DryRun,
		Force:  c.Force,
	}

	switch c.Source {
	case "withings":
		_, err := service.ImportWithingsActivities(ctx, c.SourceDir, options)
		if err != nil {
			return fmt.Errorf("import failed: %w", err)
		}
	case "spo2":
		_, err := service.ImportSpO2(ctx, c.SourceDir, options)
		if err != nil {
			return fmt.Errorf("import failed: %w", err)
		}
	case "mfp":
		_, err := service.ImportMFPWeight(ctx, c.SourceDir, options)
		if err != nil {
			return fmt.Errorf("import failed: %w", err)
		}
	case "weight":
		_, err := service.ImportWeight(ctx, c.SourceDir, options)
		if err != nil {
			return fmt.Errorf("import failed: %w", err)
		}
	case "sleep":
		_, err := service.ImportSleep(ctx, c.SourceDir, options)
		if err != nil {
			return fmt.Errorf("import failed: %w", err)
		}
	case "bp", "blood-pressure":
		_, err := service.ImportBloodPressure(ctx, c.SourceDir, options)
		if err != nil {
			return fmt.Errorf("import failed: %w", err)
		}
	case "hr", "heart-rate":
		_, err := service.ImportHeartRate(ctx, c.SourceDir, options)
		if err != nil {
			return fmt.Errorf("import failed: %w", err)
		}
	case "all":
		_, err := service.ImportWithingsActivities(ctx, c.SourceDir, options)
		if err != nil {
			return fmt.Errorf("import withings failed: %w", err)
		}
		_, err = service.ImportSpO2(ctx, c.SourceDir, options)
		if err != nil {
			return fmt.Errorf("import spo2 failed: %w", err)
		}
		_, err = service.ImportMFPWeight(ctx, c.SourceDir, options)
		if err != nil {
			return fmt.Errorf("import mfp failed: %w", err)
		}
		_, err = service.ImportWeight(ctx, c.SourceDir, options)
		if err != nil {
			return fmt.Errorf("import weight failed: %w", err)
		}
		_, err = service.ImportSleep(ctx, c.SourceDir, options)
		if err != nil {
			return fmt.Errorf("import sleep failed: %w", err)
		}
		_, err = service.ImportBloodPressure(ctx, c.SourceDir, options)
		if err != nil {
			return fmt.Errorf("import blood pressure failed: %w", err)
		}
		_, err = service.ImportHeartRate(ctx, c.SourceDir, options)
		if err != nil {
			return fmt.Errorf("import heart rate failed: %w", err)
		}
		fmt.Println("Health data migration complete.")
	default:
		return fmt.Errorf("unknown source: %s (use: withings, spo2, mfp, weight, sleep, bp, hr, or all)", c.Source)
	}

	return nil
}

func HandleHealthImport() int {
	fs := flag.NewFlagSet("health-import", flag.ExitOnError)
	dbPath := fs.String("db", "", "Path to SQLite database")
	sourceDir := fs.String("source", "", "Source directory containing health data files")
	dryRun := fs.Bool("dry-run", false, "Validate without importing")
	force := fs.Bool("force", false, "Skip duplicate checking")
	sourceType := fs.String("source-type", "all", "Source type: withings, spo2, mfp, weight, sleep, bp, hr, or all")

	fs.Parse(os.Args[2:])

	cmd := ImportCommand{
		DBPath:    *dbPath,
		SourceDir: *sourceDir,
		DryRun:    *dryRun,
		Force:     *force,
		Source:    *sourceType,
	}

	ctx := context.Background()

	if err := cmd.Execute(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	return 0
}
