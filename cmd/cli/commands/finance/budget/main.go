package budget

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// HandleBudgetCommand routes to the appropriate budget subcommand
func HandleBudgetCommand(command string, args []string) int {
	// Extract just the arguments after the subcommand
	// args comes in as: ["budget", "import-csv", "budget.csv"]
	// We need to pass: ["budget.csv"] or ["--file", "budget.csv"]
	var subArgs []string
	for i, arg := range args {
		if arg == command && i < len(args)-1 {
			subArgs = args[i+1:]
			break
		}
	}

	// If we didn't find the command in args, try using args as-is (for flags)
	if subArgs == nil {
		subArgs = args
	}

	switch command {
	case "init":
		return InitCommand()
	case "show":
		return ShowCommand(subArgs)
	case "calendar":
		return CalendarCommand(subArgs)
	case "scenario":
		return ScenarioCommand(subArgs)
	case "export":
		return ExportCommand(subArgs)
	case "help", "--help", "-h":
		PrintBudgetUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Unknown budget command: %s\n\n", command)
		PrintBudgetUsage()
		return 1
	}
}

// PrintBudgetUsage displays help information
func PrintBudgetUsage() {
	fmt.Println("🏦 Budget Management Commands")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  hominem budget <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init                Create a new budget interactively")
	fmt.Println("  show                Display current budget status")
	fmt.Println("  calendar            Show cash flow calendar")
	fmt.Println("  scenario            Test what-if scenarios")
	fmt.Println("  export              Export budget to various formats")
	fmt.Println("  help                Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  hominem budget init                              # Create new budget")
	fmt.Println("  hominem budget show                              # Show budget summary")
	fmt.Println("  hominem budget show --view categories            # Show by category")
	fmt.Println("  hominem budget show --view goals                 # Show goals progress")
	fmt.Println("  hominem budget calendar                          # Cash flow calendar")
	fmt.Println("  hominem budget scenario --reduce dining 50       # Test spending cut")
	fmt.Println("  hominem budget export --format yaml              # Export to YAML")
	fmt.Println()
	fmt.Println("Configuration location:")
	fmt.Println("  ~/.config/hominem/budget/")
	fmt.Println()
	fmt.Println("Files:")
	fmt.Println("  config.yaml         Goals, rules, and settings")
	fmt.Println("  cash_flow.yaml      Income and expenses")
	fmt.Println("  scenarios/          Saved scenarios")
	fmt.Println()
}

// ExportCommand exports budget to various formats
func ExportCommand(args []string) int {
	fs := flag.NewFlagSet("budget-export", flag.ExitOnError)
	format := fs.String("format", "csv", "Export format: csv, json, yaml")
	output := fs.String("output", "", "Output file (default: stdout)")

	fs.Parse(args)

	// Load budget
	config, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		return 1
	}

	switch *format {
	case "csv":
		return exportCSV(config, *output)
	case "json":
		return exportJSON(config, *output)
	case "yaml":
		return exportYAML(config, *output)
	default:
		fmt.Fprintf(os.Stderr, "❌ Unknown format: %s. Use csv, json, or yaml\n", *format)
		return 1
	}
}

func exportCSV(config *BudgetConfig, output string) int {
	var w *os.File
	var err error

	if output == "" {
		w = os.Stdout
	} else {
		w, err = os.Create(output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to create file: %v\n", err)
			return 1
		}
		defer w.Close()
	}

	// Write income
	fmt.Fprintln(w, "Type,Category,Name,Amount,DayOfMonth")
	for _, income := range config.CashFlow.Income {
		fmt.Fprintf(w, "Income,%s,%s,%.2f,%d\n",
			income.Category, income.Name, income.Amount, income.DayOfMonth)
	}

	// Write expenses
	for _, expense := range config.CashFlow.Expenses {
		essential := "No"
		if expense.Essential {
			essential = "Yes"
		}
		fmt.Fprintf(w, "Expense,%s,%s,%.2f,%d,%s\n",
			expense.Category, expense.Name, expense.Amount, expense.DayOfMonth, essential)
	}

	if output != "" {
		fmt.Printf("✅ Exported to %s\n", output)
	}

	return 0
}

func exportJSON(config *BudgetConfig, output string) int {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to marshal JSON: %v\n", err)
		return 1
	}

	if output == "" {
		fmt.Println(string(data))
	} else {
		if err := os.WriteFile(output, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to write file: %v\n", err)
			return 1
		}
		fmt.Printf("✅ Exported to %s\n", output)
	}

	return 0
}

func exportYAML(config *BudgetConfig, output string) int {
	data, err := yaml.Marshal(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to marshal YAML: %v\n", err)
		return 1
	}

	if output == "" {
		fmt.Println(string(data))
	} else {
		if err := os.WriteFile(output, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to write file: %v\n", err)
			return 1
		}
		fmt.Printf("✅ Exported to %s\n", output)
	}

	return 0
}
