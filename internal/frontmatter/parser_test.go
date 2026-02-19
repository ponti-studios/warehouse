package frontmatter

import (
	"strings"
	"testing"
)

func TestParseYAMLFrontmatter_ValidYAML(t *testing.T) {
	content := "---\ntitle: Hello\nstatus: draft\n---\nBody text here."
	result, err := ParseYAMLFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasFM {
		t.Fatal("expected HasFM to be true")
	}
	if result.Frontmatter["title"] != "Hello" {
		t.Fatalf("expected title 'Hello', got %v", result.Frontmatter["title"])
	}
	if result.Frontmatter["status"] != "draft" {
		t.Fatalf("expected status 'draft', got %v", result.Frontmatter["status"])
	}
	if result.Body != "Body text here." {
		t.Fatalf("unexpected body: %q", result.Body)
	}
	if result.Delimiter != DelimiterYAML {
		t.Fatalf("expected delimiter %q, got %q", DelimiterYAML, result.Delimiter)
	}
}

func TestParseYAMLFrontmatter_NoFrontmatter(t *testing.T) {
	content := "Just a normal markdown file.\nNo frontmatter here."
	result, err := ParseYAMLFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasFM {
		t.Fatal("expected HasFM to be false")
	}
	if result.Body != content {
		t.Fatalf("expected body to be original content, got %q", result.Body)
	}
}

func TestParseYAMLFrontmatter_EmptyFrontmatter(t *testing.T) {
	content := "---\n\n---\nBody after empty frontmatter."
	result, err := ParseYAMLFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasFM {
		t.Fatal("expected HasFM to be true for empty frontmatter block")
	}
	if len(result.Frontmatter) != 0 {
		t.Fatalf("expected empty frontmatter map, got %v", result.Frontmatter)
	}
	if result.Body != "Body after empty frontmatter." {
		t.Fatalf("unexpected body: %q", result.Body)
	}
}

func TestParseYAMLFrontmatter_BOM(t *testing.T) {
	content := "\ufeff---\ntitle: BOM Test\n---\nBody."
	result, err := ParseYAMLFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasFM {
		t.Fatal("expected HasFM to be true after BOM stripping")
	}
	if result.Frontmatter["title"] != "BOM Test" {
		t.Fatalf("expected title 'BOM Test', got %v", result.Frontmatter["title"])
	}
}

func TestParseYAMLFrontmatter_WindowsNewlines(t *testing.T) {
	content := "---\r\ntitle: Windows\r\n---\r\nBody.\r\n"
	result, err := ParseYAMLFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasFM {
		t.Fatal("expected HasFM to be true for CRLF content")
	}
	if result.Frontmatter["title"] != "Windows" {
		t.Fatalf("expected title 'Windows', got %v", result.Frontmatter["title"])
	}
}

func TestParseYAMLFrontmatter_InvalidYAML(t *testing.T) {
	content := "---\n: invalid: yaml: [broken\n---\nBody."
	_, err := ParseYAMLFrontmatter(content)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestParseYAMLFrontmatter_JSONInsideDelimiters(t *testing.T) {
	content := "---\n{\"title\": \"JSON Note\", \"status\": \"draft\"}\n---\nBody."
	result, err := ParseYAMLFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error parsing JSON inside --- delimiters: %v", err)
	}
	if !result.HasFM {
		t.Fatal("expected HasFM to be true for JSON inside --- delimiters")
	}
	if result.Frontmatter["title"] != "JSON Note" {
		t.Fatalf("expected title 'JSON Note', got %v", result.Frontmatter["title"])
	}
	if result.Frontmatter["status"] != "draft" {
		t.Fatalf("expected status 'draft', got %v", result.Frontmatter["status"])
	}
}

func TestParseYAMLFrontmatter_UnclosedDelimiter(t *testing.T) {
	content := "---\ntitle: Unclosed\nBody without closing delimiter."
	result, err := ParseYAMLFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasFM {
		t.Fatal("expected HasFM to be false for unclosed delimiter")
	}
}

func TestBuildYAMLFrontmatter_RoundTrip(t *testing.T) {
	fm := map[string]interface{}{
		"title":  "Round Trip",
		"status": "draft",
	}
	body := "Some body content.\n"

	built, err := BuildYAMLFrontmatter(fm, body)
	if err != nil {
		t.Fatalf("unexpected error building: %v", err)
	}

	parsed, err := ParseYAMLFrontmatter(built)
	if err != nil {
		t.Fatalf("unexpected error parsing round-trip: %v", err)
	}
	if !parsed.HasFM {
		t.Fatal("expected HasFM after round-trip")
	}
	if parsed.Frontmatter["title"] != "Round Trip" {
		t.Fatalf("round-trip title mismatch: got %v", parsed.Frontmatter["title"])
	}
	if parsed.Frontmatter["status"] != "draft" {
		t.Fatalf("round-trip status mismatch: got %v", parsed.Frontmatter["status"])
	}
	if !strings.Contains(parsed.Body, "Some body content.") {
		t.Fatalf("round-trip body mismatch: got %q", parsed.Body)
	}
}

func TestBuildYAMLFrontmatter_NilMap(t *testing.T) {
	built, err := BuildYAMLFrontmatter(nil, "body")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(built, "---\n") {
		t.Fatalf("expected output to start with ---, got %q", built[:20])
	}
}

func TestBuildJSONFrontmatter_ValidOutput(t *testing.T) {
	fm := map[string]interface{}{
		"title":  "JSON Test",
		"status": "published",
	}

	built, err := BuildJSONFrontmatter(fm, "Body.\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(built, "---\n") {
		t.Fatalf("expected --- prefix, got %q", built[:10])
	}
	if !strings.Contains(built, `"title"`) {
		t.Fatalf("expected JSON key in output, got %q", built)
	}
	if !strings.Contains(built, "Body.\n") {
		t.Fatalf("expected body in output")
	}
}

func TestBuildJSONFrontmatter_RoundTrip(t *testing.T) {
	fm := map[string]interface{}{
		"title":  "JSON Round Trip",
		"status": "draft",
	}

	built, err := BuildJSONFrontmatter(fm, "Body content.\n")
	if err != nil {
		t.Fatalf("unexpected error building: %v", err)
	}

	parsed, err := ParseYAMLFrontmatter(built)
	if err != nil {
		t.Fatalf("unexpected error parsing JSON frontmatter: %v", err)
	}
	if !parsed.HasFM {
		t.Fatal("expected HasFM after JSON round-trip")
	}
	if parsed.Frontmatter["title"] != "JSON Round Trip" {
		t.Fatalf("JSON round-trip title mismatch: got %v", parsed.Frontmatter["title"])
	}
}

func TestBuildJSONFrontmatter_NilMap(t *testing.T) {
	built, err := BuildJSONFrontmatter(nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(built, "---\n") {
		t.Fatalf("expected --- prefix")
	}
}

func TestBuildYAMLFrontmatterWithOrder_Basic(t *testing.T) {
	fm := map[string]interface{}{
		"title":  "Test Title",
		"status": "draft",
	}
	body := "Body content.\n"

	built, err := BuildYAMLFrontmatterWithOrder(fm, body, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(built, "title: Test Title") {
		t.Fatalf("expected title in output: %s", built)
	}
	if !strings.Contains(built, "status: draft") {
		t.Fatalf("expected status in output: %s", built)
	}
	if !strings.Contains(built, "Body content.") {
		t.Fatalf("expected body in output: %s", built)
	}
}

func TestBuildYAMLFrontmatterWithOrder_WithOriginal(t *testing.T) {
	originalFM := `title: Original Title
status: published
tags:
  - one
  - two
`
	fm := map[string]interface{}{
		"title":  "Updated Title",
		"status": "draft",
		"tags":   []interface{}{"one", "two"},
	}
	body := "Updated body.\n"

	built, err := BuildYAMLFrontmatterWithOrder(fm, body, originalFM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(built, "Updated Title") {
		t.Fatalf("expected updated title in output: %s", built)
	}
	if !strings.Contains(built, "draft") {
		t.Fatalf("expected draft status in output: %s", built)
	}
	if !strings.Contains(built, "Updated body.") {
		t.Fatalf("expected updated body in output: %s", built)
	}
}
