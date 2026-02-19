package browser

import (
	"fmt"

	"gogogo/internal/browser"
)

func runStart() error {
	b := browser.New()
	defer b.Close()

	fmt.Print("Starting Chrome on :9222... ")
	if err := b.Start(false); err != nil {
		return fmt.Errorf("failed to start Chrome: %w", err)
	}
	fmt.Println("✓ Chrome started")
	return nil
}

func runStartProfile() error {
	b := browser.New()
	defer b.Close()

	fmt.Print("Starting Chrome on :9222 with profile... ")
	if err := b.Start(true); err != nil {
		return fmt.Errorf("failed to start Chrome: %w", err)
	}
	fmt.Println("✓ Chrome started with profile")
	return nil
}
