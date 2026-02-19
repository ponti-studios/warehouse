package tracking

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gogogo/internal/domain/shared"
	"gogogo/internal/domain/tracking"
	"github.com/jedib0t/go-pretty/v6/table"
)

// TrackingSearchOptions configures how tracking entries are searched.
type TrackingSearchOptions struct {
	Query    string
	Filter   tracking.TrackingFilter
	Types    []string
	FromDate *shared.Timestamp
	ToDate   *shared.Timestamp
}

// NewTrackingSearchOptions returns a search configuration with sane defaults.
func NewTrackingSearchOptions() TrackingSearchOptions {
	filter := tracking.DefaultTrackingFilter()
	filter.Limit = 60
	return TrackingSearchOptions{
		Filter: filter,
	}
}

// WithQuery sets the search query.
func (opts TrackingSearchOptions) WithQuery(q string) TrackingSearchOptions {
	opts.Query = q
	return opts
}

// WithDateRange restricts the date range of the search.
func (opts TrackingSearchOptions) WithDateRange(start, end time.Time) TrackingSearchOptions {
	startTs := shared.Timestamp(start)
	endTs := shared.Timestamp(end)
	opts.FromDate = &startTs
	opts.ToDate = &endTs
	opts.Filter.StartDate = opts.FromDate
	opts.Filter.EndDate = opts.ToDate
	return opts
}

// WithType filters the search to specific entry types.
func (opts TrackingSearchOptions) WithType(entryType string) TrackingSearchOptions {
	opts.Types = []string{entryType}
	opts.Filter.Type = entryType
	return opts
}

// WithLimit overrides the default limit.
func (opts TrackingSearchOptions) WithLimit(limit int) TrackingSearchOptions {
	opts.Filter.Limit = limit
	return opts
}

// SearchTrackingEntries executes a search request against the repository.
func SearchTrackingEntries(ctx context.Context, repo tracking.TrackingEntryRepository, opts TrackingSearchOptions) ([]tracking.TrackingEntry, error) {
	if opts.Query == "" {
		return nil, fmt.Errorf("tracking search requires a non-empty query")
	}
	return repo.Search(ctx, opts.Query, opts.Filter)
}

// RenderTrackingSearchResults displays the entries in a human-friendly table.
func RenderTrackingSearchResults(entries []tracking.TrackingEntry, query string) {
	if len(entries) == 0 {
		fmt.Println("No tracking entries matched your search.")
		return
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.AppendHeader(table.Row{"#", "ID", "Type", "Title", "Date", "Snippet", "Points"})

	for idx, entry := range entries {
		snippet := highlightSnippet(query, entry.Content)
		dateStr := entry.Date.Time().Format(time.RFC3339)
		t.AppendRow(table.Row{
			idx + 1,
			entry.ID.String(),
			entry.Type,
			entry.Title,
			dateStr,
			snippet,
			len(entry.DataPoints),
		})
	}

	fmt.Println(t.Render())
}

// highlightSnippet returns a small excerpt with the query terms emphasized.
func highlightSnippet(query, content string) string {
	if content == "" {
		return ""
	}

	normalized := strings.ToLower(content)
	keywords := strings.Fields(strings.ToLower(query))
	preview := content
	for _, keyword := range keywords {
		if keyword == "" {
			continue
		}
		if idx := strings.Index(normalized, keyword); idx >= 0 {
			start := idx - 30
			if start < 0 {
				start = 0
			}
			end := idx + len(keyword) + 30
			if end > len(content) {
				end = len(content)
			}
			preview = strings.TrimSpace(content[start:end])
			break
		}
	}

	for _, keyword := range keywords {
		if keyword == "" {
			continue
		}
		preview = strings.ReplaceAll(preview, keyword, fmt.Sprintf("[%s]", keyword))
	}

	if len(preview) > 180 {
		return preview[:180] + "..."
	}
	return preview
}
