package finance

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gogogo/cmd/cli/commands/finance/budget"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "finance",
		Short: "Finance and budget utilities",
	}

	cmd.AddCommand(budgetCmd())
	cmd.AddCommand(calculatorCmd())
	cmd.AddCommand(reportCmd())
	cmd.AddCommand(dashboardCmd())

	return cmd
}

func budgetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "budget",
		Short: "Budget management",
	}

	cmd.AddCommand(budgetInitCmd())
	cmd.AddCommand(budgetShowCmd())
	cmd.AddCommand(budgetCalendarCmd())
	cmd.AddCommand(budgetScenarioCmd())
	cmd.AddCommand(budgetExportCmd())

	return cmd
}

func budgetInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create a new budget interactively",
		Run: func(cmd *cobra.Command, args []string) {
			if code := budget.HandleBudgetCommand("init", args); code != 0 {
				os.Exit(code)
			}
		},
	}
}

func budgetShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Display current budget status",
		Run: func(cmd *cobra.Command, args []string) {
			if code := budget.HandleBudgetCommand("show", args); code != 0 {
				os.Exit(code)
			}
		},
	}
}

func budgetCalendarCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "calendar",
		Short: "Show cash flow calendar",
		Run: func(cmd *cobra.Command, args []string) {
			if code := budget.HandleBudgetCommand("calendar", args); code != 0 {
				os.Exit(code)
			}
		},
	}
}

func budgetScenarioCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scenario",
		Short: "Test what-if scenarios",
		Run: func(cmd *cobra.Command, args []string) {
			if code := budget.HandleBudgetCommand("scenario", args); code != 0 {
				os.Exit(code)
			}
		},
	}
}

func budgetExportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Export budget to various formats",
		Run: func(cmd *cobra.Command, args []string) {
			if code := budget.HandleBudgetCommand("export", args); code != 0 {
				os.Exit(code)
			}
		},
	}
}

func calculatorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "calculator",
		Short: "Goal calculator (interactive TUI)",
		Run: func(cmd *cobra.Command, args []string) {
			if err := RunCLI(); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}
}

func reportCmd() *cobra.Command {
	var dbPath string
	var reportType string
	var format string
	var output string
	var page int
	var perPage int
	var startDate string
	var endDate string
	var accounts []string
	var categories []string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate financial reports",
		RunE: func(cmd *cobra.Command, args []string) error {
			reportCmd := ReportCommand{
				DBPath:     dbPath,
				ReportType: reportType,
				Format:     format,
				Output:     output,
				Page:       page,
				PerPage:    perPage,
				StartDate:  startDate,
				EndDate:    endDate,
				Accounts:   accounts,
				Categories: categories,
			}
			return reportCmd.Execute(cmd.Context())
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Path to SQLite database (required)")
	cmd.Flags().StringVar(&reportType, "type", "", "Report type: transactions, accounts, categories (required)")
	cmd.Flags().StringVar(&format, "format", "table", "Output format: table, json, tui")
	cmd.Flags().StringVar(&output, "output", "", "Output file (optional, prints to stdout if not specified)")
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&perPage, "per-page", 50, "Items per page")
	cmd.Flags().StringVar(&startDate, "start-date", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&endDate, "end-date", "", "End date (YYYY-MM-DD)")
	cmd.Flags().StringArrayVar(&accounts, "account", []string{}, "Filter by account")
	cmd.Flags().StringArrayVar(&categories, "category", []string{}, "Filter by category")

	cmd.MarkFlagRequired("db")
	cmd.MarkFlagRequired("type")

	return cmd
}

func dashboardCmd() *cobra.Command {
	var dbPath string
	var format string

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Show financial dashboard summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			dashboardCmd := DashboardCommand{
				DBPath: dbPath,
				Format: format,
			}
			return dashboardCmd.Execute(cmd.Context())
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Path to SQLite database (required)")
	cmd.Flags().StringVar(&format, "format", "table", "Output format: table, json")

	cmd.MarkFlagRequired("db")

	return cmd
}
