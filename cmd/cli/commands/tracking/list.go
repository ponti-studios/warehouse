package tracking

import (
	"fmt"
	"time"

	"gogogo/internal/domain/shared"
	"gogogo/internal/domain/tracking"
	"github.com/jedib0t/go-pretty/v6/table"
)

type TrackingListOptions struct {
	Filter tracking.TrackingFilter
}

func NewTrackingListOptions() TrackingListOptions {
	filter := tracking.DefaultTrackingFilter()
	filter.Limit = 40
	return TrackingListOptions{
		Filter: filter,
	}
}

func (opts TrackingListOptions) WithDateRange(start, end time.Time) TrackingListOptions {
	startTs := shared.Timestamp(start)
	endTs := shared.Timestamp(end)
	opts.Filter.StartDate = &startTs
	opts.Filter.EndDate = &endTs
	return opts
}

func (opts TrackingListOptions) WithType(entryType string) TrackingListOptions {
	opts.Filter.Type = entryType
	return opts
}

func (opts TrackingListOptions) WithSourceFile(file string) TrackingListOptions {
	opts.Filter.SourceFile = file
	return opts
}

func renderTrackingEntries(entries []tracking.TrackingEntry) {
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.AppendHeader(table.Row{"#", "ID", "Type", "Title", "Date", "Source", "Points"})

	for idx, entry := range entries {
		date := entry.Date.Time().Format(time.RFC3339)
		pointCount := len(entry.DataPoints)
		t.AppendRow(table.Row{
			idx + 1,
			entry.ID.String(),
			entry.Type,
			entry.Title,
			date,
			entry.SourceFile,
			pointCount,
		})
	}

	fmt.Println(t.Render())
}

func PrintTrackingList(entries []tracking.TrackingEntry) {
	if len(entries) == 0 {
		fmt.Println("No tracking entries found (try widening your filters).")
		return
	}
	renderTrackingEntries(entries)
}
