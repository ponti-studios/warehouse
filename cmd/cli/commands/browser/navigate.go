package browser

import (
	"fmt"
	"os"

	"gogogo/internal/browser"
)

func runNavigate(url string, newTab bool) error {
	b := browser.New()
	defer b.Close()

	ctx, cancel, err := b.Connect()
	if err != nil {
		return fmt.Errorf("connecting to Chrome: %w", err)
	}
	defer cancel()

	fmt.Printf("Navigating to %s... ", url)
	if err := b.Navigate(ctx, url, newTab); err != nil {
		return fmt.Errorf("navigating: %w", err)
	}
	fmt.Println("✓ Done")
	return nil
}

func runEval(code string) error {
	b := browser.New()
	defer b.Close()

	ctx, cancel, err := b.Connect()
	if err != nil {
		return fmt.Errorf("connecting to Chrome: %w", err)
	}
	defer cancel()

	result, err := b.EvalAsync(ctx, code)
	if err != nil {
		return fmt.Errorf("evaluating: %w", err)
	}

	fmt.Println(result)
	return nil
}

func runScreenshot() error {
	b := browser.New()
	defer b.Close()

	ctx, cancel, err := b.Connect()
	if err != nil {
		return fmt.Errorf("connecting to Chrome: %w", err)
	}
	defer cancel()

	filepath, err := b.Screenshot(ctx)
	if err != nil {
		return fmt.Errorf("taking screenshot: %w", err)
	}

	fmt.Println(filepath)
	return nil
}

func runCookies() error {
	b := browser.New()
	defer b.Close()

	ctx, cancel, err := b.Connect()
	if err != nil {
		return fmt.Errorf("connecting to Chrome: %w", err)
	}
	defer cancel()

	json, err := b.GetCookiesJSON(ctx)
	if err != nil {
		return fmt.Errorf("getting cookies: %w", err)
	}

	fmt.Println(json)
	os.Exit(0)
	return nil
}
