package budget

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// BudgetConfig represents the complete budget configuration
type BudgetConfig struct {
	Version   string   `yaml:"version"`
	Name      string   `yaml:"name"`
	CreatedAt string   `yaml:"created_at"`
	UpdatedAt string   `yaml:"updated_at"`
	Goals     []Goal   `yaml:"goals"`
	CashFlow  CashFlow `yaml:"cash_flow"`
	Rules     []Rule   `yaml:"rules"`
	Settings  Settings `yaml:"settings"`
}

// Goal represents a financial goal
type Goal struct {
	Name           string  `yaml:"name"`
	Target         float64 `yaml:"target"`
	Current        float64 `yaml:"current"`
	Priority       int     `yaml:"priority"`
	TimelineMonths int     `yaml:"timeline_months"`
	Type           string  `yaml:"type"` // safety_net, purchase, experience, freedom, giving
	Recurring      bool    `yaml:"recurring"`
	AutoAllocate   bool    `yaml:"auto_allocate"`
	Percentage     float64 `yaml:"percentage,omitempty"` // For auto-allocation
}

// CashFlow represents income and expenses
type CashFlow struct {
	Income   []IncomeItem  `yaml:"income"`
	Expenses []ExpenseItem `yaml:"expenses"`
}

// IncomeItem represents a source of income
type IncomeItem struct {
	Name        string  `yaml:"name"`
	Amount      float64 `yaml:"amount"`
	DayOfMonth  int     `yaml:"day_of_month"`
	Month       int     `yaml:"month,omitempty"` // For annual/quarterly income
	Type        string  `yaml:"type"`            // fixed, variable
	Reliability float64 `yaml:"reliability"`     // 0.0-1.0 confidence level
	Category    string  `yaml:"category"`
}

// ExpenseItem represents an expense
type ExpenseItem struct {
	Name        string  `yaml:"name"`
	Amount      float64 `yaml:"amount"`
	DayOfMonth  int     `yaml:"day_of_month"`
	Type        string  `yaml:"type"`        // fixed, variable
	Essential   bool    `yaml:"essential"`   // Can you live without it?
	Adjustable  bool    `yaml:"adjustable"`  // Can you change the amount?
	Flexibility float64 `yaml:"flexibility"` // 0.0-1.0 how much can you reduce it
	Category    string  `yaml:"category"`
}

// Rule represents an automated budget rule
type Rule struct {
	Name      string                 `yaml:"name"`
	Trigger   string                 `yaml:"trigger"`
	Condition string                 `yaml:"condition,omitempty"`
	Action    string                 `yaml:"action"`
	Params    map[string]interface{} `yaml:"params"`
	Enabled   bool                   `yaml:"enabled"`
}

// Settings contains budget configuration
type Settings struct {
	Currency        string  `yaml:"currency"`
	DefaultView     string  `yaml:"default_view"`
	AlertThreshold  float64 `yaml:"alert_threshold"` // % of budget to alert
	SafetyBuffer    float64 `yaml:"safety_buffer"`   // Minimum cash to maintain
	AutoSaveEnabled bool    `yaml:"auto_save_enabled"`
}

// DefaultConfig creates a new budget with sensible defaults
func DefaultConfig() *BudgetConfig {
	return &BudgetConfig{
		Version: "1.0",
		Name:    "My Budget",
		Goals: []Goal{
			{
				Name:           "Emergency Fund",
				Target:         15000,
				Current:        0,
				Priority:       1,
				TimelineMonths: 12,
				Type:           "safety_net",
				AutoAllocate:   true,
				Percentage:     10,
			},
		},
		CashFlow: CashFlow{
			Income:   []IncomeItem{},
			Expenses: []ExpenseItem{},
		},
		Rules: []Rule{
			{
				Name:    "Month-end sweep to savings",
				Trigger: "month_end",
				Action:  "sweep_to_savings",
				Enabled: true,
				Params: map[string]interface{}{
					"keep_minimum": 1000,
				},
			},
		},
		Settings: Settings{
			Currency:        "USD",
			DefaultView:     "summary",
			AlertThreshold:  90,
			SafetyBuffer:    3000,
			AutoSaveEnabled: true,
		},
	}
}

// GetConfigDir returns the budget configuration directory
func GetConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".config", "hominem", "budget")
}

// GetConfigPath returns the path to the main config file
func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), "config.yaml")
}

// GetCashFlowPath returns the path to the cash flow file
func GetCashFlowPath() string {
	return filepath.Join(GetConfigDir(), "cash_flow.yaml")
}

// LoadConfig loads the budget configuration
func LoadConfig() (*BudgetConfig, error) {
	configPath := GetConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no budget found. Run 'hominem budget init' to create one")
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config BudgetConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the budget configuration
func SaveConfig(config *BudgetConfig) error {
	configDir := GetConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := GetConfigPath()
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// SaveCashFlow saves just the cash flow portion
func SaveCashFlow(cashFlow *CashFlow) error {
	configDir := GetConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	cashFlowPath := GetCashFlowPath()
	data, err := yaml.Marshal(cashFlow)
	if err != nil {
		return fmt.Errorf("failed to marshal cash flow: %w", err)
	}

	if err := os.WriteFile(cashFlowPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cash flow: %w", err)
	}

	return nil
}

// LoadCashFlow loads just the cash flow portion
func LoadCashFlow() (*CashFlow, error) {
	cashFlowPath := GetCashFlowPath()

	data, err := os.ReadFile(cashFlowPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &CashFlow{
				Income:   []IncomeItem{},
				Expenses: []ExpenseItem{},
			}, nil
		}
		return nil, fmt.Errorf("failed to read cash flow: %w", err)
	}

	var cashFlow CashFlow
	if err := yaml.Unmarshal(data, &cashFlow); err != nil {
		return nil, fmt.Errorf("failed to parse cash flow: %w", err)
	}

	return &cashFlow, nil
}

// Exists checks if a budget configuration exists
func Exists() bool {
	_, err := os.Stat(GetConfigPath())
	return !os.IsNotExist(err)
}

// CalculateMonthlyIncome calculates total monthly income
func (cf *CashFlow) CalculateMonthlyIncome() float64 {
	total := 0.0
	for _, income := range cf.Income {
		if income.Month == 0 {
			// Monthly income
			total += income.Amount
		} else {
			// Annual/quarterly income - normalize to monthly
			total += income.Amount / 12
		}
	}
	return total
}

// CalculateMonthlyExpenses calculates total monthly expenses
func (cf *CashFlow) CalculateMonthlyExpenses() float64 {
	total := 0.0
	for _, expense := range cf.Expenses {
		total += expense.Amount
	}
	return total
}

// CalculateEssentialExpenses calculates essential monthly expenses
func (cf *CashFlow) CalculateEssentialExpenses() float64 {
	total := 0.0
	for _, expense := range cf.Expenses {
		if expense.Essential {
			total += expense.Amount
		}
	}
	return total
}

// CalculateDiscretionaryExpenses calculates non-essential monthly expenses
func (cf *CashFlow) CalculateDiscretionaryExpenses() float64 {
	total := 0.0
	for _, expense := range cf.Expenses {
		if !expense.Essential {
			total += expense.Amount
		}
	}
	return total
}

// GetCategories returns unique expense categories
func (cf *CashFlow) GetCategories() []string {
	categoryMap := make(map[string]bool)
	for _, expense := range cf.Expenses {
		categoryMap[expense.Category] = true
	}

	categories := make([]string, 0, len(categoryMap))
	for cat := range categoryMap {
		categories = append(categories, cat)
	}
	return categories
}

// GetExpensesByCategory returns expenses grouped by category
func (cf *CashFlow) GetExpensesByCategory() map[string][]ExpenseItem {
	grouped := make(map[string][]ExpenseItem)
	for _, expense := range cf.Expenses {
		grouped[expense.Category] = append(grouped[expense.Category], expense)
	}
	return grouped
}

// GetTotalByCategory returns total spending by category
func (cf *CashFlow) GetTotalByCategory() map[string]float64 {
	totals := make(map[string]float64)
	for _, expense := range cf.Expenses {
		totals[expense.Category] += expense.Amount
	}
	return totals
}

// CalculateRunway calculates how many months you can survive with current cash
func CalculateRunway(cash float64, monthlyExpenses float64) float64 {
	if monthlyExpenses <= 0 {
		return 999 // Infinite runway
	}
	return cash / monthlyExpenses
}

// FormatCurrency formats a currency value
func FormatCurrency(amount float64) string {
	if amount < 0 {
		return fmt.Sprintf("-$%.2f", -amount)
	}
	return fmt.Sprintf("$%.2f", amount)
}

// FormatPercent formats a percentage value
func FormatPercent(value float64) string {
	return fmt.Sprintf("%.1f%%", value)
}
