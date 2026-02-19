package money

import "fmt"

// Money represents a monetary amount with currency
type Money struct {
	Amount   float64
	Currency string // "USD"
}

// New creates a new Money instance with USD as default currency
func New(amount float64) Money {
	return Money{Amount: amount, Currency: "USD"}
}

// NewWithCurrency creates a new Money instance with the specified currency
func NewWithCurrency(amount float64, currency string) Money {
	return Money{Amount: amount, Currency: currency}
}

// Format returns a formatted string representation of the money
// Negative amounts are shown with a minus sign
func (m Money) Format() string {
	if m.Amount < 0 {
		return fmt.Sprintf("-$%.2f", -m.Amount)
	}
	return fmt.Sprintf("$%.2f", m.Amount)
}

// FormatAbsolute returns the absolute value formatted (no minus sign)
func (m Money) FormatAbsolute() string {
	return fmt.Sprintf("$%.2f", m.Abs().Amount)
}

// Abs returns the absolute value of the money
func (m Money) Abs() Money {
	if m.Amount < 0 {
		return Money{Amount: -m.Amount, Currency: m.Currency}
	}
	return m
}

// IsNegative returns true if the amount is negative
func (m Money) IsNegative() bool {
	return m.Amount < 0
}

// IsPositive returns true if the amount is positive
func (m Money) IsPositive() bool {
	return m.Amount > 0
}

// IsZero returns true if the amount is zero
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// Add returns a new Money with the sum of this and other
func (m Money) Add(other Money) Money {
	return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}
}

// Subtract returns a new Money with the difference of this and other
func (m Money) Subtract(other Money) Money {
	return Money{Amount: m.Amount - other.Amount, Currency: m.Currency}
}

// Multiply returns a new Money multiplied by the given factor
func (m Money) Multiply(factor float64) Money {
	return Money{Amount: m.Amount * factor, Currency: m.Currency}
}

// Divide returns a new Money divided by the given divisor
func (m Money) Divide(divisor float64) Money {
	if divisor == 0 {
		return Money{Amount: 0, Currency: m.Currency}
	}
	return Money{Amount: m.Amount / divisor, Currency: m.Currency}
}

// Negate returns the negation of the money amount
func (m Money) Negate() Money {
	return Money{Amount: -m.Amount, Currency: m.Currency}
}

// Round rounds the amount to the specified number of decimal places
func (m Money) Round(places int) Money {
	factor := 1.0
	for i := 0; i < places; i++ {
		factor *= 10
	}
	return Money{
		Amount:   float64(int(m.Amount*factor+0.5)) / factor,
		Currency: m.Currency,
	}
}

// Sum calculates the sum of multiple Money values
func Sum(monies ...Money) Money {
	if len(monies) == 0 {
		return New(0)
	}
	currency := monies[0].Currency
	total := 0.0
	for _, m := range monies {
		total += m.Amount
	}
	return Money{Amount: total, Currency: currency}
}
