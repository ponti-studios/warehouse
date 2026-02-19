package biz

import (
	"fmt"
	"os"
)

func HandleBizCommand(command string, args []string) int {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	newArgs := []string{"hominem"}
	newArgs = append(newArgs, command)
	if len(args) > 2 {
		newArgs = append(newArgs, args[2:]...)
	}
	os.Args = newArgs

	switch command {
	case "plan":
		return HandlePlanCommand()
	case "help", "--help", "-h":
		PrintBizUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Unknown biz command: %s\n\n", command)
		PrintBizUsage()
		return 1
	}
}

func PrintBizUsage() {
	fmt.Println("Business modeling commands:")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  hominem biz <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  plan        Calculate business projections from YAML model")
	fmt.Println("  help        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  hominem biz plan                              # Use subscription.yaml in Documents/business/")
	fmt.Println("  hominem biz plan mymodel.yaml                  # Use custom model file")
	fmt.Println("  hominem biz plan --months 12                  # Override projection length (default: 3)")
	fmt.Println("  hominem biz plan --no-chart                    # Skip ASCII charts")
	fmt.Println("  hominem biz plan --format json                 # Output as JSON")
	fmt.Println("  hominem biz plan --output projections.csv      # Export to CSV")
	fmt.Println()
	fmt.Println("Model files location: ~/Documents/business/")
	fmt.Println()
	fmt.Println("YAML Model Format:")
	fmt.Println("  name: 'My Business'")
	fmt.Println("  description: 'Description'")
	fmt.Println("  time:")
	fmt.Println("    months: 24                    # projection period (default: 3)")
	fmt.Println("  model:")
	fmt.Println("    type: subscription")
	fmt.Println("    users:")
	fmt.Println("      initial: 0                  # starting users")
	fmt.Println("      monthly_new: 20             # new users per month")
	fmt.Println("      churn_rate: 0.05            # 5% monthly churn")
	fmt.Println("      growth_rate: 0.20           # 20% monthly growth")
	fmt.Println("    pricing:")
	fmt.Println("      monthly: 5.00              # monthly price")
	fmt.Println("      annual: 20.00              # annual price")
	fmt.Println("    costs:")
	fmt.Println("      per_user: 5.00             # COGS per user")
	fmt.Println("      fixed_monthly: 130000       # fixed costs (salaries, rent)")
	fmt.Println("      sales_marketing: 5000       # S&M per month")
	fmt.Println("    cash_on_hand: 500000         # OPTIONAL: starting cash (default: $500k)")
	fmt.Println()
	fmt.Println("Key Metrics Calculated:")
	fmt.Println("  MRR, ARR, Revenue, Costs, Profit")
	fmt.Println("  Cash Balance, Runway, Break-even Month")
	fmt.Println("  CAC, LTV, LTV:CAC Ratio, Payback Period")
	fmt.Println("  MRR Growth %, Gross Margin")
}
