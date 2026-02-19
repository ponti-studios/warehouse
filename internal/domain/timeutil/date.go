package timeutil

import (
	"fmt"
	"time"
)

// DateFormat is the standard date format used throughout the application
const DateFormat = "2006-01-02"

// Date is a type-safe wrapper around date strings in YYYY-MM-DD format
// It provides efficient sorting/filtering (lexicographic comparison works)
// while maintaining type safety
type Date string

// NewDate creates a Date from a time.Time
func NewDate(t time.Time) Date {
	return Date(t.Format(DateFormat))
}

// ParseDate parses a date string in various formats
// Handles both "2024-01-01" and "2025-04-03T00:00:00.000Z" formats
func ParseDate(s string) (Date, error) {
	if s == "" {
		return "", fmt.Errorf("empty date string")
	}

	layouts := []string{
		DateFormat,
		time.RFC3339,
		"2006-01-02T00:00:00.000Z",
		"2006-01-02 15:04:05",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return NewDate(t), nil
		}
	}

	return "", fmt.Errorf("invalid date format: %s", s)
}

// Time converts the Date back to time.Time
func (d Date) Time() (time.Time, error) {
	if d == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}
	return time.Parse(DateFormat, string(d))
}

// Year extracts the year from the date
func (d Date) Year() int {
	t, err := d.Time()
	if err != nil {
		return 0
	}
	return t.Year()
}

// Month extracts the month from the date
func (d Date) Month() time.Month {
	t, err := d.Time()
	if err != nil {
		return 0
	}
	return t.Month()
}

// Day extracts the day from the date
func (d Date) Day() int {
	t, err := d.Time()
	if err != nil {
		return 0
	}
	return t.Day()
}

// AddDays adds the specified number of days to the date
func (d Date) AddDays(days int) Date {
	t, err := d.Time()
	if err != nil {
		return d
	}
	return NewDate(t.AddDate(0, 0, days))
}

// AddMonths adds the specified number of months to the date
func (d Date) AddMonths(months int) Date {
	t, err := d.Time()
	if err != nil {
		return d
	}
	return NewDate(t.AddDate(0, months, 0))
}

// Before returns true if this date is before the other date
func (d Date) Before(other Date) bool {
	return d < other
}

// After returns true if this date is after the other date
func (d Date) After(other Date) bool {
	return d > other
}

// Equal returns true if this date equals the other date
func (d Date) Equal(other Date) bool {
	return d == other
}

// IsZero returns true if the date is empty/unset
func (d Date) IsZero() bool {
	return d == ""
}

// String returns the string representation of the date
func (d Date) String() string {
	return string(d)
}

// StartOfMonth returns the first day of the month
func (d Date) StartOfMonth() Date {
	t, err := d.Time()
	if err != nil {
		return d
	}
	return NewDate(time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()))
}

// EndOfMonth returns the last day of the month
func (d Date) EndOfMonth() Date {
	t, err := d.Time()
	if err != nil {
		return d
	}
	nextMonth := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
	return NewDate(nextMonth.AddDate(0, 0, -1))
}

// Today returns today's date
func Today() Date {
	return NewDate(time.Now())
}

// ParseDateRange parses start and end dates, returning them as Date objects
func ParseDateRange(startStr, endStr string) (start, end Date, err error) {
	if startStr != "" {
		start, err = ParseDate(startStr)
		if err != nil {
			return "", "", fmt.Errorf("invalid start date: %w", err)
		}
	}

	if endStr != "" {
		end, err = ParseDate(endStr)
		if err != nil {
			return "", "", fmt.Errorf("invalid end date: %w", err)
		}
	}

	if !start.IsZero() && !end.IsZero() && start.After(end) {
		return "", "", fmt.Errorf("start date cannot be after end date")
	}

	return start, end, nil
}
