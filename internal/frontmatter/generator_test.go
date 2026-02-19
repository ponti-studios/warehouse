package frontmatter

import (
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestNewGeneratorRegistry_BuiltinsRegistered(t *testing.T) {
	reg := NewGeneratorRegistry()
	builtins := []string{
		"uuid-v4", "utc-now", "timestamp-rfc3339",
		"kebab-filename", "kebab-title",
		"file-ctime", "file-mtime", "sha1",
	}
	for _, name := range builtins {
		if _, ok := reg.Get(name); !ok {
			t.Fatalf("expected built-in generator %q to be registered", name)
		}
	}
}

func TestGeneratorRegistry_RegisterAndGet(t *testing.T) {
	reg := NewGeneratorRegistry()
	custom := UTCTimestampGenerator{}
	reg.Register("custom-gen", custom)

	g, ok := reg.Get("custom-gen")
	if !ok {
		t.Fatal("expected custom generator to be retrievable")
	}
	if g == nil {
		t.Fatal("expected non-nil generator")
	}
}

func TestGeneratorRegistry_GetMissing(t *testing.T) {
	reg := NewGeneratorRegistry()
	_, ok := reg.Get("nonexistent")
	if ok {
		t.Fatal("expected Get to return false for missing generator")
	}
}

func TestGeneratorRegistry_OverwriteExisting(t *testing.T) {
	reg := NewGeneratorRegistry()
	reg.Register("uuid-v4", SHA1Generator{})

	g, ok := reg.Get("uuid-v4")
	if !ok {
		t.Fatal("expected overwritten generator to exist")
	}
	_, isSHA1 := g.(SHA1Generator)
	if !isSHA1 {
		t.Fatal("expected overwritten generator to be SHA1Generator")
	}
}

func TestUUIDV4Generator(t *testing.T) {
	gen := UUIDV4Generator{}
	val, err := gen.Generate(GeneratorContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	uuidRe := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	if !uuidRe.MatchString(val) {
		t.Fatalf("expected UUID format, got %q", val)
	}

	val2, _ := gen.Generate(GeneratorContext{})
	if val == val2 {
		t.Fatal("expected two UUIDs to differ")
	}
}

func TestUTCTimestampGenerator(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 30, 0, 0, time.UTC)
	gen := UTCTimestampGenerator{}
	val, err := gen.Generate(GeneratorContext{Now: now})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "2025-06-15T12:30:00Z" {
		t.Fatalf("expected '2025-06-15T12:30:00Z', got %q", val)
	}
}

func TestUTCTimestampGenerator_ZeroTime(t *testing.T) {
	gen := UTCTimestampGenerator{}
	val, err := gen.Generate(GeneratorContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tsRe := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)
	if !tsRe.MatchString(val) {
		t.Fatalf("expected RFC3339 format, got %q", val)
	}
}

func TestKebabFilenameGenerator(t *testing.T) {
	tests := []struct {
		ctx    GeneratorContext
		expect string
	}{
		{GeneratorContext{FileName: "My File Name.md"}, "my-file-name"},
		{GeneratorContext{FilePath: "/path/to/CamelCase.md"}, "camel-case"},
		{GeneratorContext{FileName: "already-kebab.md"}, "already-kebab"},
		{GeneratorContext{FileName: "with spaces and CAPS.md"}, "with-spaces-and-caps"},
	}

	gen := KebabFilenameGenerator{}
	for _, tt := range tests {
		val, err := gen.Generate(tt.ctx)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", tt.ctx.FileName, err)
		}
		if val != tt.expect {
			t.Fatalf("expected %q, got %q (input: %+v)", tt.expect, val, tt.ctx)
		}
	}
}

func TestKebabTitleGenerator(t *testing.T) {
	gen := KebabTitleGenerator{}

	val, err := gen.Generate(GeneratorContext{Title: "My Great Title"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "my-great-title" {
		t.Fatalf("expected 'my-great-title', got %q", val)
	}

	_, err = gen.Generate(GeneratorContext{Title: ""})
	if err == nil {
		t.Fatal("expected error for empty title")
	}

	_, err = gen.Generate(GeneratorContext{Title: "   "})
	if err == nil {
		t.Fatal("expected error for whitespace-only title")
	}
}

func TestSHA1Generator(t *testing.T) {
	gen := SHA1Generator{}

	val, err := gen.Generate(GeneratorContext{FilePath: "/path/to/file.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(val) != 40 {
		t.Fatalf("expected 40 char hex string, got %q (len %d)", val, len(val))
	}

	val2, _ := gen.Generate(GeneratorContext{FilePath: "/path/to/file.md"})
	if val != val2 {
		t.Fatal("expected deterministic SHA1 for same input")
	}

	val3, _ := gen.Generate(GeneratorContext{FilePath: "/different/path.md"})
	if val == val3 {
		t.Fatal("expected different SHA1 for different input")
	}
}

func TestSHA1Generator_NoSource(t *testing.T) {
	gen := SHA1Generator{}
	_, err := gen.Generate(GeneratorContext{})
	if err == nil {
		t.Fatal("expected error when no source available")
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"CamelCase", "camel-case"},
		{"already-kebab", "already-kebab"},
		{"With Spaces", "with-spaces"},
		{"ALLCAPS", "allcaps"},
		{"mixed--Dashes", "mixed-dashes"},
		{"trailing---", "trailing"},
		{"---leading", "leading"},
	}
	for _, tt := range tests {
		got := toKebabCase(tt.input)
		if got != tt.expect {
			t.Fatalf("toKebabCase(%q): expected %q, got %q", tt.input, tt.expect, got)
		}
	}
}

func TestFileCTimeGenerator(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	gen := FileCTimeGenerator{}
	val, err := gen.Generate(GeneratorContext{Now: now})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(val, "2025-01-01") {
		t.Fatalf("expected timestamp starting with 2025-01-01, got %q", val)
	}
}

func TestFileMTimeGenerator(t *testing.T) {
	now := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	gen := FileMTimeGenerator{}
	val, err := gen.Generate(GeneratorContext{Now: now})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "2025-12-31T23:59:59Z" {
		t.Fatalf("expected '2025-12-31T23:59:59Z', got %q", val)
	}
}
