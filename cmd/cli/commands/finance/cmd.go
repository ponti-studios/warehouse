package finance

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	report "gogogo/cmd/cli/commands/report"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "finance",
		Short: "Finance and budget utilities",
	}

	cmd.AddCommand(budgetCmd())
	cmd.AddCommand(reportCmd())

	return cmd
}

func budgetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "budget",
		Short: "Budget calculator (interactive TUI)",
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
			reportCmd := report.ReportCommand{
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
	cmd.Flags().StringVar(&format, "format", "table", "Output format: table, json")
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
