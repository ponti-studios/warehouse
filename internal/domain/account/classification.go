package account

import (
	"fmt"
	"strings"
)

// ClassifyByName determines account type based on name heuristics
func ClassifyByName(name string) Type {
	nameLower := strings.ToLower(name)

	switch {
	case strings.Contains(nameLower, "checking"), strings.Contains(nameLower, "current"):
		return TypeChecking
	case strings.Contains(nameLower, "savings"), strings.Contains(nameLower, "money market"):
		return TypeSavings
	case strings.Contains(nameLower, "card"), strings.Contains(nameLower, "platinum"),
		strings.Contains(nameLower, "sapphire"), strings.Contains(nameLower, "gold"),
		strings.Contains(nameLower, "venture"), strings.Contains(nameLower, "quicksilver"),
		strings.Contains(nameLower, "savor"), strings.Contains(nameLower, "rewards"),
		strings.Contains(nameLower, "double cash"):
		return TypeCredit
	case strings.Contains(nameLower, "401k"), strings.Contains(nameLower, "ira"),
		strings.Contains(nameLower, "investment"):
		return TypeInvestments
	case strings.Contains(nameLower, "cash"), strings.Contains(nameLower, "pay"):
		return TypeCash
	default:
		return TypeUnknown
	}
}

// Classification represents the asset/liability classification of an account
type Classification string

const (
	ClassificationAsset     Classification = "asset"
	ClassificationLiability Classification = "liability"
)

// GetClassification returns the classification for an account type
func (t Type) GetClassification() Classification {
	if t.IsAsset() {
		return ClassificationAsset
	}
	return ClassificationLiability
}

// String returns the string representation
func (c Classification) String() string {
	return string(c)
}

// ValidateAlias checks if an alias mapping is valid
func ValidateAlias(alias, canonical string) error {
	if strings.TrimSpace(alias) == "" {
		return fmt.Errorf("alias cannot be empty")
	}
	if strings.TrimSpace(canonical) == "" {
		return fmt.Errorf("canonical name cannot be empty")
	}
	if alias == canonical {
		return fmt.Errorf("alias and canonical name cannot be the same")
	}
	return nil
}
