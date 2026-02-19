package biz

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Input struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Time        TimeConfig  `yaml:"time"`
	Model       ModelConfig `yaml:"model"`
}

const DefaultCashOnHand = 500000

func GetDefaultMonths() int {
	return 3
}

type TimeConfig struct {
	Months int `yaml:"months"`
}

type ModelConfig struct {
	Type        string            `yaml:"type"`
	Users       UsersConfig       `yaml:"users"`
	Pricing     PricingConfig     `yaml:"pricing,omitempty"`
	Transaction TransactionConfig `yaml:"transaction,omitempty"`
	Costs       CostsConfig       `yaml:"costs"`
	CashOnHand  float64           `yaml:"cash_on_hand,omitempty"`
}

type UsersConfig struct {
	Initial    int     `yaml:"initial"`
	MonthlyNew int     `yaml:"monthly_new"`
	ChurnRate  float64 `yaml:"churn_rate"`
	GrowthRate float64 `yaml:"growth_rate,omitempty"`
}

type PricingConfig struct {
	Monthly float64 `yaml:"monthly"`
	Annual  float64 `yaml:"annual"`
}

type TransactionConfig struct {
	AvgTotal       float64 `yaml:"avg_total"`
	FeeTotal       float64 `yaml:"fee_total"`
	RevenuePerUser float64 `yaml:"revenue_per_user"`
	TxPerUser      int     `yaml:"tx_per_user"`
}

type CostsConfig struct {
	PerUser        float64 `yaml:"per_user"`
	FixedMonthly   float64 `yaml:"fixed_monthly"`
	SalesMarketing float64 `yaml:"sales_marketing"`
}

func LoadInput(path string) (*Input, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var input Input
	if err := yaml.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := ValidateInput(&input); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &input, nil
}

func ValidateInput(input *Input) error {
	if input.Time.Months <= 0 {
		return fmt.Errorf("time.months must be a positive integer")
	}

	if input.Model.Type != "subscription" && input.Model.Type != "marketplace" {
		return fmt.Errorf("model.type must be 'subscription' or 'marketplace'")
	}

	if input.Model.Users.Initial < 0 {
		return fmt.Errorf("model.users.initial must be non-negative")
	}

	if input.Model.Users.MonthlyNew < 0 {
		return fmt.Errorf("model.users.monthly_new must be non-negative")
	}

	if input.Model.Users.ChurnRate < 0 || input.Model.Users.ChurnRate > 1 {
		return fmt.Errorf("model.users.churn_rate must be between 0 and 1")
	}

	if input.Model.Type == "subscription" {
		if input.Model.Users.GrowthRate < 0 || input.Model.Users.GrowthRate > 1 {
			return fmt.Errorf("model.users.growth_rate must be between 0 and 1 for subscription model")
		}
		if input.Model.Pricing.Monthly < 0 {
			return fmt.Errorf("model.pricing.monthly must be non-negative for subscription model")
		}
		if input.Model.Pricing.Annual < 0 {
			return fmt.Errorf("model.pricing.annual must be non-negative for subscription model")
		}
	}

	if input.Model.Type == "marketplace" {
		if input.Model.Transaction.AvgTotal < 0 {
			return fmt.Errorf("model.transaction.avg_total must be non-negative for marketplace model")
		}
		if input.Model.Transaction.FeeTotal < 0 || input.Model.Transaction.FeeTotal > 1 {
			return fmt.Errorf("model.transaction.fee_total must be between 0 and 1 for marketplace model")
		}
		if input.Model.Transaction.RevenuePerUser < 0 {
			return fmt.Errorf("model.transaction.revenue_per_user must be non-negative for marketplace model")
		}
		if input.Model.Transaction.TxPerUser < 0 {
			return fmt.Errorf("model.transaction.tx_per_user must be non-negative for marketplace model")
		}
	}

	if input.Model.Costs.FixedMonthly < 0 {
		return fmt.Errorf("model.costs.fixed_monthly must be non-negative")
	}

	if input.Model.Costs.PerUser < 0 {
		return fmt.Errorf("model.costs.per_user must be non-negative")
	}

	if input.Model.Costs.SalesMarketing < 0 {
		return fmt.Errorf("model.costs.sales_marketing must be non-negative")
	}

	return nil
}

func (i *Input) GetDefaultFile() string {
	return "business-model.yaml"
}

func (i *Input) GetMonths() int {
	if i.Time.Months > 0 {
		return i.Time.Months
	}
	return GetDefaultMonths()
}

func (i *Input) GetCashOnHand() float64 {
	if i.Model.CashOnHand > 0 {
		return i.Model.CashOnHand
	}
	return DefaultCashOnHand
}
