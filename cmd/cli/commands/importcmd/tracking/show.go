package tracking

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gogogo/internal/domain/shared"
	domainTracking "gogogo/internal/domain/tracking"
)

type TrackingShowOptions struct {
	EntryID shared.ID
}

func NewTrackingShowOptions() TrackingShowOptions {
	return TrackingShowOptions{}
}

func (opts TrackingShowOptions) WithID(id shared.ID) TrackingShowOptions {
	opts.EntryID = id
	return opts
}

func RenderTrackingEntry(entry domainTracking.TrackingEntry) {
	fmt.Printf("Entry: %s (%s)\n", entry.Title, entry.ID)
	fmt.Printf("Type: %s | Date: %s | Source: %s\n", entry.Type, entry.Date.Time().Format(time.RFC3339), entry.SourceFile)

	if len(entry.Metadata) > 0 {
		fmt.Println("Metadata:")
		for key, value := range entry.Metadata {
			fmt.Printf("  • %s: %v\n", key, value)
		}
	}

	fmt.Println("\nContent:")
	fmt.Println(strings.TrimSpace(entry.Content))

	if len(entry.DataPoints) > 0 {
		fmt.Printf("\nData Points (%d):\n", len(entry.DataPoints))
		for _, point := range entry.DataPoints {
			fmt.Printf("  • %s | %s | %s %s\n",
				point.Date.Time().Format(time.RFC3339),
				point.Type,
				point.Value,
				point.Unit,
			)
		}
	}
}

func ShowTrackingEntry(ctx context.Context, repo domainTracking.TrackingEntryRepository, opts TrackingShowOptions) (*domainTracking.TrackingEntry, error) {
	if opts.EntryID == "" {
		return nil, fmt.Errorf("entry ID is required to show tracking entry")
	}

	entry, err := repo.GetByID(ctx, opts.EntryID)
	if err != nil {
		return nil, fmt.Errorf("failed to load tracking entry: %w", err)
	}

	return entry, nil
}
