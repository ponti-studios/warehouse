package xlsx

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

func HandleXLSXCommand(command string, args []string) int {
	if command == "help" || command == "--help" || command == "-h" {
		PrintXLSXUsage()
		return 0
	}

	switch command {
	case "csv":
		return HandleXLSXToCSV()
	default:
		fmt.Fprintf(os.Stderr, "Unknown xlsx command: %s\n\n", command)
		PrintXLSXUsage()
		return 1
	}
}

func PrintXLSXUsage() {
	fmt.Println("XLSX utility commands:")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  hominem xlsx <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  csv         Convert XLSX file(s) to CSV")
	fmt.Println("  help        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  hominem xlsx csv file.xlsx")
	fmt.Println("  hominem xlsx csv file.xlsx --outputDir ./csv/")
	fmt.Println("  hominem xlsx csv --all  # converts all xlsx in current directory")
}

func HandleXLSXToCSV() int {
	fs := flag.NewFlagSet("xlsx-csv", flag.ExitOnError)
	outputDir := fs.String("outputDir", "", "Output directory (default: same as input)")
	all := fs.Bool("all", false, "Convert all xlsx files in current directory")

	fs.Parse(os.Args[3:])

	if !*all && fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: XLSX file required")
		fmt.Fprintln(os.Stderr, "Usage: hominem xlsx csv <file.xlsx> [options]")
		fmt.Fprintln(os.Stderr, "   or: hominem xlsx csv --all [options]")
		return 1
	}

	if *all {
		return ConvertAllXLSX(*outputDir)
	}

	inputPath := fs.Arg(0)
	return ConvertXLSX(inputPath, *outputDir)
}

func ConvertAllXLSX(outputDir string) int {
	entries, err := os.ReadDir(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
		return 1
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(entry.Name()), ".xlsx") {
			if err := convertXLSXFile(entry.Name(), outputDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error converting %s: %v\n", entry.Name(), err)
			} else {
				count++
			}
		}
	}

	fmt.Printf("Converted %d file(s) to CSV\n", count)
	return 0
}

func ConvertXLSX(inputPath string, outputDir string) int {
	if err := convertXLSXFile(inputPath, outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

func convertXLSXFile(inputPath string, outputDir string) error {
	f, err := excelize.OpenFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()

	for _, sheetName := range sheets {
		rows, err := f.GetRows(sheetName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not read sheet %s: %v\n", sheetName, err)
			continue
		}

		var outputPath string
		if len(sheets) == 1 {
			baseName := strings.TrimSuffix(inputPath, filepath.Ext(inputPath))
			outputPath = baseName + ".csv"
		} else {
			baseName := strings.TrimSuffix(inputPath, filepath.Ext(inputPath))
			outputPath = baseName + "_" + sheetName + ".csv"
		}

		if outputDir != "" {
			baseName := filepath.Base(strings.TrimSuffix(inputPath, filepath.Ext(inputPath)))
			if len(sheets) == 1 {
				outputPath = filepath.Join(outputDir, baseName+".csv")
			} else {
				outputPath = filepath.Join(outputDir, baseName+"_"+sheetName+".csv")
			}
		}

		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil && filepath.Dir(outputPath) != "." {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()

		for _, row := range rows {
			for i, cell := range row {
				if i > 0 {
					fmt.Fprint(file, ",")
				}
				fmt.Fprint(file, escapeCSV(cell))
			}
			fmt.Fprintln(file)
		}

		fmt.Printf("Created: %s\n", outputPath)
	}

	return nil
}

func escapeCSV(s string) string {
	s = strings.ReplaceAll(s, "\"", "\"\"")
	if strings.ContainsAny(s, ",\"\n") {
		return "\"" + s + "\""
	}
	return s
}
