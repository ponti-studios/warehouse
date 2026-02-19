package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   string
		wantErr bool
	}{
		{"valid value", "name", "test", false},
		{"empty value", "name", "", true},
		{"whitespace only", "name", "   ", true},
		{"valid date", "date", "2024-01-15", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequired(tt.field, tt.value)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateDate(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		wantErr bool
	}{
		{"valid date", "2024-01-15", false},
		{"invalid format", "01-15-2024", true},
		{"invalid format 2", "2024/01/15", true},
		{"empty - optional", "", false},
		{"valid leap year", "2024-02-29", false},
		{"valid dec date", "2024-12-31", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDate("date", tt.date)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateCurrency(t *testing.T) {
	tests := []struct {
		name     string
		currency string
		wantErr  bool
	}{
		{"valid USD", "USD", false},
		{"valid EUR", "EUR", false},
		{"valid lowercase", "usd", false},
		{"invalid code", "XXX", true},
		{"empty - optional", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCurrency("currency", tt.currency)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateAccountType(t *testing.T) {
	tests := []struct {
		name    string
		accType string
		wantErr bool
	}{
		{"valid CHECKING", "CHECKING", false},
		{"valid SAVINGS", "SAVINGS", false},
		{"valid CREDIT", "CREDIT", false},
		{"valid lowercase", "checking", false},
		{"invalid type", "INVALID", true},
		{"empty - optional", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAccountType("type", tt.accType)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateDomain(t *testing.T) {
	tests := []struct {
		name    string
		domain  string
		wantErr bool
	}{
		{"valid finance", "finance", false},
		{"valid health", "health", false},
		{"valid tracking", "tracking", false},
		{"invalid domain", "invalid", true},
		{"empty - optional", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDomain("domain", tt.domain)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
