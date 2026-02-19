package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

var (
	dateRegex     = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	currencyCodes = map[string]bool{
		"USD": true, "EUR": true, "GBP": true, "JPY": true, "CAD": true,
		"AUD": true, "CHF": true, "CNY": true, "INR": true, "MXN": true,
		"BRL": true, "KRW": true, "SGD": true, "HKD": true, "NOK": true,
		"SEK": true, "DKK": true, "NZD": true, "ZAR": true, "RUB": true,
		"TRY": true, "PLN": true, "THB": true, "IDR": true, "MYR": true,
		"PHP": true, "CZK": true, "ILS": true, "CLP": true, "AED": true,
		"COP": true, "SAR": true, "TWD": true, "RON": true, "BGN": true,
	}
	validAccountTypes = map[string]bool{
		"CHECKING":    true,
		"SAVINGS":     true,
		"CREDIT":      true,
		"INVESTMENTS": true,
		"CASH":        true,
		"LOAN":        true,
	}
	validDomains = map[string]bool{
		"finance":  true,
		"health":   true,
		"tracking": true,
	}
)

func ValidateRequired(field, value string) *ValidationError {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{Field: field, Message: "is required"}
	}
	return nil
}

func ValidateDate(field, value string) *ValidationError {
	if value == "" {
		return nil
	}
	if !dateRegex.MatchString(value) {
		return &ValidationError{Field: field, Message: "must be in YYYY-MM-DD format"}
	}

	// Check if date is actually valid (e.g., not 2024-99-99)
	_, err := time.Parse("2006-01-02", value)
	if err != nil {
		return &ValidationError{Field: field, Message: "must be a valid date"}
	}
	return nil
}

func ValidateCurrency(field, value string) *ValidationError {
	if value == "" {
		return nil
	}
	value = strings.ToUpper(value)
	if !currencyCodes[value] {
		return &ValidationError{Field: field, Message: "must be a valid ISO 4217 currency code"}
	}
	return nil
}

func ValidateAccountType(field, value string) *ValidationError {
	if value == "" {
		return nil
	}
	value = strings.ToUpper(value)
	if !validAccountTypes[value] {
		return &ValidationError{
			Field:   field,
			Message: "must be one of: CHECKING, SAVINGS, CREDIT, INVESTMENTS, CASH, LOAN",
		}
	}
	return nil
}

func ValidateDomain(field, value string) *ValidationError {
	if value == "" {
		return nil
	}
	if !validDomains[value] {
		return &ValidationError{
			Field:   field,
			Message: "must be one of: finance, health, tracking",
		}
	}
	return nil
}

func ValidateNotEmpty(field, value string) *ValidationError {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{Field: field, Message: "cannot be empty"}
	}
	return nil
}
