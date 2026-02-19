package frontmatter

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Generator defines a generator implementation for frontmatter fields.
type Generator interface {
	Generate(ctx GeneratorContext) (string, error)
}

// GeneratorContext provides information for field generation.
type GeneratorContext struct {
	FieldName string
	FilePath  string
	FileName  string
	Title     string
	Now       time.Time
}

// GeneratorRegistry holds named generator implementations.
type GeneratorRegistry struct {
	registry map[string]Generator
}

// NewGeneratorRegistry creates a registry with built-in generators.
func NewGeneratorRegistry() *GeneratorRegistry {
	reg := &GeneratorRegistry{registry: map[string]Generator{}}
	reg.Register("uuid-v4", UUIDV4Generator{})
	reg.Register("utc-now", UTCTimestampGenerator{})
	reg.Register("timestamp-rfc3339", UTCTimestampGenerator{})
	reg.Register("kebab-filename", KebabFilenameGenerator{})
	reg.Register("kebab-title", KebabTitleGenerator{})
	reg.Register("file-ctime", FileCTimeGenerator{})
	reg.Register("file-mtime", FileMTimeGenerator{})
	reg.Register("sha1", SHA1Generator{})
	return reg
}

// Register installs a generator under a name.
func (r *GeneratorRegistry) Register(name string, generator Generator) {
	if r.registry == nil {
		r.registry = map[string]Generator{}
	}
	r.registry[name] = generator
}

// Get retrieves a generator by name.
func (r *GeneratorRegistry) Get(name string) (Generator, bool) {
	g, ok := r.registry[name]
	return g, ok
}

// UUIDV4Generator generates UUID v4 values.
type UUIDV4Generator struct{}

func (g UUIDV4Generator) Generate(_ GeneratorContext) (string, error) {
	return uuid.New().String(), nil
}

// UTCTimestampGenerator generates RFC3339 timestamps in UTC.
type UTCTimestampGenerator struct{}

func (g UTCTimestampGenerator) Generate(ctx GeneratorContext) (string, error) {
	t := ctx.Now
	if t.IsZero() {
		t = time.Now().UTC()
	}
	return t.UTC().Format("2006-01-02T15:04:05Z"), nil
}

// KebabFilenameGenerator generates a kebab-case slug from file name.
type KebabFilenameGenerator struct{}

func (g KebabFilenameGenerator) Generate(ctx GeneratorContext) (string, error) {
	base := ctx.FileName
	if base == "" && ctx.FilePath != "" {
		base = filepath.Base(ctx.FilePath)
	}
	base = strings.TrimSuffix(base, filepath.Ext(base))
	return toKebabCase(base), nil
}

// KebabTitleGenerator generates a kebab-case slug from title.
type KebabTitleGenerator struct{}

func (g KebabTitleGenerator) Generate(ctx GeneratorContext) (string, error) {
	if strings.TrimSpace(ctx.Title) == "" {
		return "", fmt.Errorf("title is required for kebab-title generator")
	}
	return toKebabCase(ctx.Title), nil
}

// FileCTimeGenerator returns a timestamp based on file creation time.
// Placeholder: returns current time until filesystem metadata is wired.
type FileCTimeGenerator struct{}

func (g FileCTimeGenerator) Generate(ctx GeneratorContext) (string, error) {
	t := ctx.Now
	if t.IsZero() {
		t = time.Now().UTC()
	}
	return t.UTC().Format("2006-01-02T15:04:05Z"), nil
}

// FileMTimeGenerator returns a timestamp based on file modification time.
// Placeholder: returns current time until filesystem metadata is wired.
type FileMTimeGenerator struct{}

func (g FileMTimeGenerator) Generate(ctx GeneratorContext) (string, error) {
	t := ctx.Now
	if t.IsZero() {
		t = time.Now().UTC()
	}
	return t.UTC().Format("2006-01-02T15:04:05Z"), nil
}

// SHA1Generator computes a SHA1 digest for the configured source.
// For now, it hashes the file path as a stable placeholder.
type SHA1Generator struct{}

func (g SHA1Generator) Generate(ctx GeneratorContext) (string, error) {
	source := ctx.FilePath
	if source == "" {
		source = ctx.FileName
	}
	if source == "" {
		return "", fmt.Errorf("no source available for sha1 generator")
	}
	sum := sha1.Sum([]byte(source))
	return hex.EncodeToString(sum[:]), nil
}

func toKebabCase(input string) string {
	re := regexp.MustCompile(`([a-z])([A-Z])`)
	input = re.ReplaceAllString(input, "$1-$2")
	input = strings.ToLower(input)
	input = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(input, "-")
	return strings.Trim(input, "-")
}
