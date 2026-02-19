package flatten

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	dryRun           bool
	directory        string
	includeParentDir bool
}

func log(message string) {
	fmt.Printf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
}

func moveFile(path string, config Config) error {
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") {
		return nil
	}

	dir := filepath.Dir(path)
	if dir == config.directory {
		return nil
	}

	newBase := base
	if config.includeParentDir {
		parentDir := filepath.Base(filepath.Dir(path))
		if parentDir != config.directory {
			ext := filepath.Ext(base)
			name := strings.TrimSuffix(base, ext)
			newBase = fmt.Sprintf("%s_%s%s", name, parentDir, ext)
		}
	}

	newPath := filepath.Join(config.directory, newBase)
	counter := 1
	for {
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			break
		}
		ext := filepath.Ext(newBase)
		name := strings.TrimSuffix(newBase, ext)
		newPath = filepath.Join(config.directory, fmt.Sprintf("%s_%d%s", name, counter, ext))
		counter++
	}

	if config.dryRun {
		log(fmt.Sprintf("Would move: %s -> %s", path, newPath))
		return nil
	}

	if err := os.Rename(path, newPath); err != nil {
		return fmt.Errorf("error moving file: %w", err)
	}
	log(fmt.Sprintf("Moved: %s -> %s", path, newPath))
	return nil
}

func Run() error {
	config := Config{}
	flagSet := flag.NewFlagSet("flatten", flag.ExitOnError)
	flagSet.BoolVar(&config.dryRun, "d", false, "Dry run mode")
	flagSet.BoolVar(&config.includeParentDir, "p", false, "Include parent directory name in filename")
	flagSet.StringVar(&config.directory, "dir", "", "Directory to flatten")
	flagSet.Parse(os.Args[2:])

	if config.directory == "" {
		fmt.Println("Error: Directory is required")
		fmt.Println("Usage: flatten -dir <directory> [-d] [-p]")
		fmt.Println("  -dir  Directory to flatten")
		fmt.Println("  -d    Dry run mode")
		fmt.Println("  -p    Include parent directory name in filename")
		return fmt.Errorf("missing required directory")
	}

	if info, err := os.Stat(config.directory); err != nil || !info.IsDir() {
		fmt.Printf("Error: '%s' is not a directory or does not exist\n", config.directory)
		return fmt.Errorf("invalid directory")
	}

	log(fmt.Sprintf("Starting file reorganization in directory: %s", config.directory))

	err := filepath.WalkDir(config.directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() != ".DS_Store" {
			if err := moveFile(path, config); err != nil {
				log(fmt.Sprintf("Failed to process: %s - %v", path, err))
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		return err
	}

	log("Finished file reorganization")
	return nil
}
