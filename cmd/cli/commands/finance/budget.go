package finance

import (
	"context"
)

type FinanceResolver struct{}

func (s *FinanceResolver) CalculateGoal(ctx context.Context, budget BudgetInput, years int, goals []*GoalInput) (float64, float64, error) {
	months := years * 12
	expensesPerMonth := 0.0

	for _, expense := range budget.Expenses {
		if expense.Cadence == ExpenseCadenceMonthly {
			expensesPerMonth += expense.Amount
		}
	}

	/**
	 * This is the total amount of money required to cover the
	 * user's current expenses while they work toward their goals.
	 */
	totalExpenses := expensesPerMonth * float64(months)

	/**
	 * This is the total amount of money required to cover the
	 * user's current expenses while they work toward their goals.
	 */
	totalRequiredForGoals := 0.0
	for _, goal := range goals {
		totalRequiredForGoals += goal.Amount
	}

	/**
	 * This is the amount of pre-tax income the user must earn to
	 * afford their goals.
	 */
	preTaxIncome := (totalRequiredForGoals + totalExpenses) / (1 - budget.TaxRate)

	/**
	 * This is the total annual pre-tax income required to reach the user's goals.
	 */
	annualPreTaxIncome := preTaxIncome / 3

	/**
	 * This is the total monthly pre-tax income required to reach the user's goals.
	 */
	monthlyPreTaxIncome := preTaxIncome / 12

	return annualPreTaxIncome, monthlyPreTaxIncome, nil
}
