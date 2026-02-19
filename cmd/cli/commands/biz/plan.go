package biz

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/jedib0t/go-pretty/v6/table"
)

type PlanCommand struct {
	File    string
	Months  int
	Output  string
	NoChart bool
	Format  string
}

func HandlePlanCommand() int {
	fs := flag.NewFlagSet("biz-plan", flag.ExitOnError)
	file := fs.String("file", "", "Path to business model YAML file")
	months := fs.Int("months", 0, "Override projection length")
	output := fs.String("output", "", "Output file (csv/json)")
	noChart := fs.Bool("no-chart", false, "Skip ASCII chart")
	format := fs.String("format", "table", "Output format: table, json")

	fs.Parse(os.Args[2:])

	inputFile := *file
	if inputFile == "" {
		if _, err := os.Stat("business-model.yaml"); err == nil {
			inputFile = "business-model.yaml"
		} else if _, err := os.Stat("../Documents/business/subscription.yaml"); err == nil {
			inputFile = "../Documents/business/subscription.yaml"
		} else {
			fmt.Fprintln(os.Stderr, "Error: No input file specified and business-model.yaml not found")
			fmt.Fprintln(os.Stderr, "Usage: hominem biz plan [file.yaml] [flags]")
			return 1
		}
	} else {
		if fs.NArg() > 0 && inputFile == "" {
			inputFile = fs.Arg(0)
		}
	}

	input, err := LoadInput(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading input: %v\n", err)
		return 1
	}

	projections := Calculate(input, *months)
	summary := Summarize(projections, input)

	switch *format {
	case "json":
		return outputJSON(input, projections, summary)
	case "csv":
		return outputCSV(input, projections, *output)
	default:
		return outputTable(input, projections, summary, *noChart)
	}
}

var (
	headerStyle   = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(0, 1).BorderForeground(lipgloss.Color("86"))
	summaryStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	labelStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	valueStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	positiveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	negativeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	warningStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
)

func outputTable(input *Input, projections []Projection, summary Summary, noChart bool) int {
	modelName := input.Name
	if modelName == "" {
		modelName = "Business Model"
	}

	monthlyPrice := input.Model.Pricing.Monthly
	annualPrice := input.Model.Pricing.Annual

	headerContent := fmt.Sprintf("%s\n", modelName)
	if input.Description != "" {
		headerContent += fmt.Sprintf("%s\n", input.Description)
	}
	headerContent += fmt.Sprintf("Pricing: $%.2f/mo ($%.2f/yr)\n", monthlyPrice, annualPrice)
	headerContent += fmt.Sprintf("Starting Cash: %s\n", FormatCurrency(input.GetCashOnHand()))
	headerContent += fmt.Sprintf("Churn: %s | Growth: %s",
		FormatPercent(input.Model.Users.ChurnRate*100),
		FormatPercent(input.Model.Users.GrowthRate*100))

	fmt.Println(headerStyle.Render(headerContent))
	fmt.Println()

	if !noChart {
		fmt.Println(RenderChartUsers(projections))
		fmt.Println()
		fmt.Println(RenderChartMRR(projections))
		fmt.Println()
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Mo", "Users", "MRR", "Revenue", "Costs", "Profit", "Cash", "Runway", "MRR%"})

	for _, p := range projections {
		mrrGrowth := ""
		if p.MRRGrowth != 0 {
			mrrGrowth = FormatPercent(p.MRRGrowth)
		}

		runway := FormatRunway(p.Runway)

		t.AppendRow([]interface{}{
			p.Month,
			p.Users,
			FormatCurrency(p.MRR),
			FormatCurrency(p.Revenue),
			FormatCurrency(p.Costs),
			FormatCurrencyColored(p.Profit),
			FormatCurrencyColored(p.CashBalance),
			runway,
			mrrGrowth,
		})
	}

	t.Render()
	fmt.Println()

	fmt.Println(summaryStyle.Render("Key Metrics:"))
	fmt.Printf("  %s ", labelStyle.Render("Total Revenue:"))
	fmt.Printf("%s  ", valueStyle.Render(FormatCurrency(summary.TotalRevenue)))
	fmt.Printf("%s ", labelStyle.Render("Total Costs:"))
	fmt.Printf("%s\n", valueStyle.Render(FormatCurrency(summary.TotalCosts)))
	fmt.Printf("  %s ", labelStyle.Render("Total Profit:"))
	fmt.Printf("%s\n", FormatCurrencyColored(summary.TotalProfit))
	fmt.Printf("  %s ", labelStyle.Render("Cash Balance:"))
	fmt.Printf("%s  ", FormatCurrencyColored(summary.CashBalance))
	fmt.Printf("%s ", labelStyle.Render("Starting Cash:"))
	fmt.Printf("%s\n", valueStyle.Render(FormatCurrency(summary.StartingCash)))
	fmt.Println()

	breakEvenInfo := "Not reached"
	breakEvenStyle := negativeStyle
	if summary.BreakEvenMonth > 0 {
		breakEvenInfo = fmt.Sprintf("Month %d (%d users)", summary.BreakEvenMonth, summary.UsersAtBreakEven)
		breakEvenStyle = positiveStyle
	}
	fmt.Printf("  %s %s\n", labelStyle.Render("Break-even:"), breakEvenStyle.Render(breakEvenInfo))
	fmt.Printf("  %s %.1f months\n", labelStyle.Render("Payback Period:"), summary.FinalPayback)
	fmt.Printf("  %s %s\n", labelStyle.Render("LTV:CAC Ratio:"), formatLTVCAC(summary.FinalLTVCAC))
	fmt.Printf("  %s $%.2f\n", labelStyle.Render("LTV:"), summary.FinalLTV)
	fmt.Printf("  %s %s\n", labelStyle.Render("Gross Margin:"), FormatPercent(summary.FinalGrossMargin))

	return 0
}

func formatLTVCAC(ratio float64) string {
	if math.IsInf(ratio, 0) || math.IsNaN(ratio) || ratio <= 0 {
		return warningStyle.Render("N/A")
	}
	if ratio >= 3 {
		return positiveStyle.Render(fmt.Sprintf("%.1f:1", ratio))
	}
	if ratio >= 1 {
		return valueStyle.Render(fmt.Sprintf("%.1f:1", ratio))
	}
	return negativeStyle.Render(fmt.Sprintf("%.1f:1", ratio))
}

func outputJSON(input *Input, projections []Projection, summary Summary) int {
	type output struct {
		Input       *Input       `json:"input"`
		Projections []Projection `json:"projections"`
		Summary     Summary      `json:"summary"`
	}

	out := output{
		Input:       input,
		Projections: projections,
		Summary:     summary,
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		return 1
	}

	fmt.Println(string(data))
	return 0
}

func outputCSV(input *Input, projections []Projection, outputFile string) int {
	var w *csv.Writer
	if outputFile == "" {
		w = csv.NewWriter(os.Stdout)
	} else {
		f, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			return 1
		}
		defer f.Close()
		w = csv.NewWriter(f)
	}
	defer w.Flush()

	w.Write([]string{"Month", "Users", "MRR", "ARR", "Revenue", "Costs", "Profit", "CashBalance", "Runway", "MRRGrowth", "LTV", "LTVCAC", "Payback", "GrossMargin"})

	for _, p := range projections {
		w.Write([]string{
			fmt.Sprintf("%d", p.Month),
			fmt.Sprintf("%d", p.Users),
			fmt.Sprintf("%.2f", p.MRR),
			fmt.Sprintf("%.2f", p.ARR),
			fmt.Sprintf("%.2f", p.Revenue),
			fmt.Sprintf("%.2f", p.Costs),
			fmt.Sprintf("%.2f", p.Profit),
			fmt.Sprintf("%.2f", p.CashBalance),
			fmt.Sprintf("%.2f", p.Runway),
			fmt.Sprintf("%.2f", p.MRRGrowth),
			fmt.Sprintf("%.2f", p.LTV),
			fmt.Sprintf("%.2f", p.LTVCAC),
			fmt.Sprintf("%.2f", p.PaybackPeriod),
			fmt.Sprintf("%.2f", p.GrossMargin),
		})
	}

	if outputFile != "" {
		fmt.Printf("Exported to %s\n", outputFile)
	}

	return 0
}
