package budget

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
)

// InitCommand creates a new budget through interactive prompts
func InitCommand() int {
	fmt.Println("🏦 Welcome to Hominem Budget")
	fmt.Println("Let's create your personalized budget system.")
	fmt.Println()

	// Check if budget already exists
	if Exists() {
		fmt.Println("⚠️  A budget already exists.")
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Do you want to overwrite it? (y/N): ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled. Your existing budget is safe.")
			return 0
		}
		fmt.Println()
	}

	// Create base config
	config := DefaultConfig()

	// Step 1: Basic Info
	fmt.Println("📋 Step 1: Basic Information")
	fmt.Println(strings.Repeat("─", 40))

	var name string
	if err := huh.NewInput().
		Title("What would you like to name this budget?").
		Placeholder("My Personal Budget").
		Value(&name).
		Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	if name != "" {
		config.Name = name
	}
	fmt.Println()

	// Step 2: Income
	fmt.Println("💰 Step 2: Monthly Income")
	fmt.Println(strings.Repeat("─", 40))

	addMoreIncome := true
	for addMoreIncome {
		var incomeName string
		var incomeAmount float64
		var incomeDay int

		if err := huh.NewInput().
			Title("Income name (e.g., 'Primary Salary', 'Side Business')").
			Placeholder("Primary Salary").
			Value(&incomeName).
			Run(); err != nil {
			break
		}

		if incomeName == "" {
			break
		}

		if err := huh.NewInput().
			Title("Monthly amount (e.g., 5000)").
			Placeholder("5000").
			Value(&incomeName).
			Run(); err != nil {
			break
		}

		// Parse amount
		fmt.Print("  Monthly amount (e.g., 5000): ")
		reader := bufio.NewReader(os.Stdin)
		amountStr, _ := reader.ReadString('\n')
		amountStr = strings.TrimSpace(amountStr)
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			fmt.Println("  Invalid amount, skipping.")
			continue
		}
		incomeAmount = amount

		// Get day of month
		fmt.Print("  Day of month received (1-31): ")
		dayStr, _ := reader.ReadString('\n')
		dayStr = strings.TrimSpace(dayStr)
		day, err := strconv.Atoi(dayStr)
		if err != nil || day < 1 || day > 31 {
			day = 1
		}
		incomeDay = day

		config.CashFlow.Income = append(config.CashFlow.Income, IncomeItem{
			Name:        incomeName,
			Amount:      incomeAmount,
			DayOfMonth:  incomeDay,
			Type:        "fixed",
			Reliability: 0.95,
			Category:    "primary",
		})

		fmt.Printf("  ✓ Added: %s - %s/month\n", incomeName, FormatCurrency(incomeAmount))
		fmt.Println()

		// Ask if they want to add more
		var addMore bool
		if err := huh.NewConfirm().
			Title("Add another income source?").
			Value(&addMore).
			Run(); err != nil {
			addMoreIncome = false
		} else {
			addMoreIncome = addMore
		}
		fmt.Println()
	}

	// Step 3: Essential Expenses
	fmt.Println("🏠 Step 3: Essential Expenses (Must-Have)")
	fmt.Println("These are expenses you can't live without: rent, utilities, insurance")
	fmt.Println(strings.Repeat("─", 40))

	essentialCategories := []string{
		"Housing (rent/mortgage)",
		"Utilities (electric, gas, water)",
		"Insurance (health, car)",
		"Transportation (car payment, gas)",
		"Minimum Food (groceries)",
		"Phone/Internet",
		"Other Essential",
	}

	for _, category := range essentialCategories {
		var hasExpense bool
		if err := huh.NewConfirm().
			Title(fmt.Sprintf("Do you have %s expenses?", category)).
			Value(&hasExpense).
			Run(); err != nil {
			continue
		}

		if !hasExpense {
			continue
		}

		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("  Monthly amount for %s: ", category)
		amountStr, _ := reader.ReadString('\n')
		amountStr = strings.TrimSpace(amountStr)
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			fmt.Println("  Invalid amount, skipping.")
			continue
		}

		// Get day of month
		fmt.Print("  Day of month due (1-31, or 0 if variable): ")
		dayStr, _ := reader.ReadString('\n')
		dayStr = strings.TrimSpace(dayStr)
		day, err := strconv.Atoi(dayStr)
		if err != nil || day < 0 || day > 31 {
			day = 1
		}

		expenseName := category
		if strings.HasSuffix(category, ")") {
			// Extract main category name
			expenseName = strings.Split(category, " (")[0]
		}

		config.CashFlow.Expenses = append(config.CashFlow.Expenses, ExpenseItem{
			Name:        expenseName,
			Amount:      amount,
			DayOfMonth:  day,
			Type:        "fixed",
			Essential:   true,
			Adjustable:  false,
			Flexibility: 0.0,
			Category:    strings.ToLower(strings.Split(expenseName, " ")[0]),
		})

		fmt.Printf("  ✓ Added: %s - %s/month\n", expenseName, FormatCurrency(amount))
		fmt.Println()
	}

	// Step 4: Discretionary Expenses
	fmt.Println("🎯 Step 4: Discretionary Expenses (Nice-to-Have)")
	fmt.Println("These improve your life but can be reduced: dining, entertainment, subscriptions")
	fmt.Println(strings.Repeat("─", 40))

	discretionaryCategories := []struct {
		name        string
		flexibility float64
	}{
		{"Dining/Restaurants", 0.8},
		{"Entertainment", 0.9},
		{"Shopping", 0.7},
		{"Subscriptions", 0.5},
		{"Gym/Fitness", 0.6},
		{"Travel", 0.9},
		{"Hobbies", 0.8},
		{"Personal Care", 0.6},
	}

	reader := bufio.NewReader(os.Stdin)
	for _, category := range discretionaryCategories {
		fmt.Printf("Monthly amount for %s (or 0 to skip): ", category.name)
		amountStr, _ := reader.ReadString('\n')
		amountStr = strings.TrimSpace(amountStr)

		if amountStr == "" || amountStr == "0" {
			continue
		}

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			fmt.Println("  Invalid amount, skipping.")
			continue
		}

		if amount == 0 {
			continue
		}

		config.CashFlow.Expenses = append(config.CashFlow.Expenses, ExpenseItem{
			Name:        category.name,
			Amount:      amount,
			DayOfMonth:  15, // Default to mid-month
			Type:        "variable",
			Essential:   false,
			Adjustable:  true,
			Flexibility: category.flexibility,
			Category:    strings.ToLower(strings.Split(category.name, "/")[0]),
		})

		fmt.Printf("  ✓ Added: %s - %s/month (can reduce %s)\n",
			category.name, FormatCurrency(amount), FormatPercent(category.flexibility*100))
	}

	fmt.Println()

	// Step 5: Goals
	fmt.Println("🎯 Step 5: Financial Goals")
	fmt.Println(strings.Repeat("─", 40))

	fmt.Println("Based on your income and expenses, here are some suggested goals:")

	monthlyIncome := config.CashFlow.CalculateMonthlyIncome()
	monthlyExpenses := config.CashFlow.CalculateMonthlyExpenses()
	surplus := monthlyIncome - monthlyExpenses

	if surplus > 0 {
		fmt.Printf("  Monthly surplus: %s\n", FormatCurrency(surplus))

		// Calculate how long to build emergency fund (3-6 months expenses)
		emergencyFundTarget := monthlyExpenses * 6
		monthsToEmergency := emergencyFundTarget / surplus

		config.Goals[0].Target = emergencyFundTarget
		config.Goals[0].TimelineMonths = int(monthsToEmergency) + 1

		fmt.Printf("  Suggested emergency fund: %s (%.0f months expenses)\n",
			FormatCurrency(emergencyFundTarget), monthsToEmergency)

		var adjustGoal bool
		if err := huh.NewConfirm().
			Title("Would you like to adjust this goal?").
			Value(&adjustGoal).
			Run(); err == nil && adjustGoal {

			fmt.Print("  Target amount: ")
			targetStr, _ := reader.ReadString('\n')
			targetStr = strings.TrimSpace(targetStr)
			if target, err := strconv.ParseFloat(targetStr, 64); err == nil && target > 0 {
				config.Goals[0].Target = target
			}

			fmt.Print("  Timeline (months): ")
			timelineStr, _ := reader.ReadString('\n')
			timelineStr = strings.TrimSpace(timelineStr)
			if timeline, err := strconv.Atoi(timelineStr); err == nil && timeline > 0 {
				config.Goals[0].TimelineMonths = timeline
			}
		}
	} else {
		fmt.Printf("  ⚠️  Monthly deficit: %s\n", FormatCurrency(-surplus))
		fmt.Println("  Your expenses exceed income. Consider reducing discretionary spending.")
	}

	fmt.Println()

	// Step 6: Review and Save
	fmt.Println("📊 Step 6: Review Your Budget")
	fmt.Println(strings.Repeat("─", 40))

	fmt.Printf("Budget Name: %s\n", config.Name)
	fmt.Printf("Monthly Income: %s\n", FormatCurrency(config.CashFlow.CalculateMonthlyIncome()))
	fmt.Printf("Monthly Expenses: %s\n", FormatCurrency(config.CashFlow.CalculateMonthlyExpenses()))
	fmt.Printf("Essential: %s | Discretionary: %s\n",
		FormatCurrency(config.CashFlow.CalculateEssentialExpenses()),
		FormatCurrency(config.CashFlow.CalculateDiscretionaryExpenses()))

	surplus = monthlyIncome - monthlyExpenses
	if surplus >= 0 {
		fmt.Printf("Monthly Surplus: %s ✅\n", FormatCurrency(surplus))
	} else {
		fmt.Printf("Monthly Deficit: %s ❌\n", FormatCurrency(surplus))
	}

	fmt.Printf("\nPrimary Goal: %s (%s in %d months)\n",
		config.Goals[0].Name,
		FormatCurrency(config.Goals[0].Target),
		config.Goals[0].TimelineMonths)

	if surplus > 0 && config.Goals[0].TimelineMonths > 0 {
		requiredMonthly := config.Goals[0].Target / float64(config.Goals[0].TimelineMonths)
		if requiredMonthly <= surplus {
			fmt.Printf("Progress: You need to save %s/month - you're on track! ✅\n",
				FormatCurrency(requiredMonthly))
		} else {
			fmt.Printf("Gap: You need to save %s/month, but only have %s surplus ⚠️\n",
				FormatCurrency(requiredMonthly), FormatCurrency(surplus))
			fmt.Println("       Consider extending timeline or reducing expenses.")
		}
	}

	fmt.Println()

	var confirmSave bool
	if err := huh.NewConfirm().
		Title("Save this budget?").
		Value(&confirmSave).
		Run(); err != nil || !confirmSave {
		fmt.Println("Budget creation cancelled.")
		return 0
	}

	// Save the configuration
	config.CreatedAt = time.Now().Format(time.RFC3339)
	config.UpdatedAt = config.CreatedAt

	if err := SaveConfig(config); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to save budget: %v\n", err)
		return 1
	}

	fmt.Println()
	fmt.Println("✅ Budget created successfully!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  • Run 'hominem budget show' to see your budget")
	fmt.Println("  • Run 'hominem budget calendar' to see cash flow timing")
	fmt.Println("  • Run 'hominem budget scenario' to test changes")
	fmt.Println()

	return 0
}
