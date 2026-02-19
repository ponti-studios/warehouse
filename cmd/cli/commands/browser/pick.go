package browser

import (
	"fmt"

	"gogogo/internal/browser"
)

func runPick(message string) error {
	b := browser.New()
	defer b.Close()

	ctx, cancel, err := b.Connect()
	if err != nil {
		return fmt.Errorf("connecting to Chrome: %w", err)
	}
	defer cancel()

	elements, err := b.Pick(ctx, message)
	if err != nil {
		return fmt.Errorf("picking: %w", err)
	}

	if elements == nil {
		fmt.Println("No element selected")
		return nil
	}

	for _, el := range elements {
		fmt.Printf("tag: %s\n", el.Tag)
		if el.ID != "" {
			fmt.Printf("id: %s\n", el.ID)
		}
		if el.Class != "" {
			fmt.Printf("class: %s\n", el.Class)
		}
		if el.Text != "" {
			fmt.Printf("text: %s\n", el.Text)
		}
		if el.Parents != "" {
			fmt.Printf("parents: %s\n", el.Parents)
		}
		fmt.Println()
	}

	return nil
}
