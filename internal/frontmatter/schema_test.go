package frontmatter

import (
	"testing"
)

func TestDefaultConfig_Version(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Version != 1 {
		t.Fatalf("expected version 1, got %d", cfg.Version)
	}
}

func TestDefaultConfig_Schemas(t *testing.T) {
	cfg := DefaultConfig()
	expected := []string{"personal", "project", "prd"}
	for _, name := range expected {
		if _, ok := cfg.Frontmatter.Schemas[name]; !ok {
			t.Fatalf("expected schema %q to be present in defaults", name)
		}
	}
	if len(cfg.Frontmatter.Schemas) != len(expected) {
		t.Fatalf("expected %d schemas, got %d", len(expected), len(cfg.Frontmatter.Schemas))
	}
}

func TestDefaultConfig_PersonalSchema(t *testing.T) {
	cfg := DefaultConfig()
	s := cfg.Frontmatter.Schemas["personal"]

	expectedRequired := []string{"title", "uid", "slug", "created", "updated", "type", "status"}
	if len(s.Required) != len(expectedRequired) {
		t.Fatalf("expected %d required fields, got %d", len(expectedRequired), len(s.Required))
	}
	reqSet := map[string]bool{}
	for _, r := range s.Required {
		reqSet[r] = true
	}
	for _, r := range expectedRequired {
		if !reqSet[r] {
			t.Fatalf("expected required field %q in personal schema", r)
		}
	}

	if s.Defaults["type"] != "reference" {
		t.Fatalf("expected personal default type 'reference', got %q", s.Defaults["type"])
	}
	if s.Defaults["status"] != "draft" {
		t.Fatalf("expected personal default status 'draft', got %q", s.Defaults["status"])
	}

	expectedGenerators := []string{"uid", "slug", "created", "updated"}
	for _, g := range expectedGenerators {
		if _, ok := s.Generators[g]; !ok {
			t.Fatalf("expected generator for field %q in personal schema", g)
		}
	}

	expectedValidators := []string{"type", "status", "slug", "uid"}
	for _, v := range expectedValidators {
		if _, ok := s.Validators[v]; !ok {
			t.Fatalf("expected validator for field %q in personal schema", v)
		}
	}
}

func TestDefaultConfig_ProjectSchema(t *testing.T) {
	cfg := DefaultConfig()
	s := cfg.Frontmatter.Schemas["project"]

	if s.Defaults["status"] != "backlog" {
		t.Fatalf("expected project default status 'backlog', got %q", s.Defaults["status"])
	}

	statusValidator, ok := s.Validators["status"]
	if !ok {
		t.Fatal("expected status validator in project schema")
	}
	allowedSet := map[string]bool{}
	for _, a := range statusValidator.Allowed {
		allowedSet[a] = true
	}
	for _, expected := range []string{"backlog", "in-progress", "review", "done"} {
		if !allowedSet[expected] {
			t.Fatalf("expected %q in project status allowed values", expected)
		}
	}
}

func TestDefaultConfig_PRDSchema(t *testing.T) {
	cfg := DefaultConfig()
	s := cfg.Frontmatter.Schemas["prd"]

	reqSet := map[string]bool{}
	for _, r := range s.Required {
		reqSet[r] = true
	}
	for _, expected := range []string{"title", "uid", "product", "status"} {
		if !reqSet[expected] {
			t.Fatalf("expected required field %q in prd schema", expected)
		}
	}
	if s.Defaults["status"] != "draft" {
		t.Fatalf("expected prd default status 'draft', got %q", s.Defaults["status"])
	}
}

func TestDefaultConfig_DirectoryMapping(t *testing.T) {
	cfg := DefaultConfig()
	dm := cfg.Frontmatter.DirectoryMapping

	expected := map[string]string{
		"notebook/personal/**": "personal",
		"notebook/projects/**": "project",
		"notebook/prds/**":     "prd",
	}

	if len(dm) != len(expected) {
		t.Fatalf("expected %d directory mappings, got %d", len(expected), len(dm))
	}
	for pattern, schema := range expected {
		if dm[pattern] != schema {
			t.Fatalf("expected mapping %q -> %q, got %q", pattern, schema, dm[pattern])
		}
	}
}

func TestDefaultConfig_Defaults(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Frontmatter.Defaults.Format != "yaml" {
		t.Fatalf("expected default format 'yaml', got %q", cfg.Frontmatter.Defaults.Format)
	}
	if cfg.Frontmatter.Defaults.GeneratorDefaults.Timestamp != "utc-now" {
		t.Fatalf("expected default timestamp generator 'utc-now', got %q", cfg.Frontmatter.Defaults.GeneratorDefaults.Timestamp)
	}
	if cfg.Frontmatter.Defaults.SlugCollision.Policy != "increment" {
		t.Fatalf("expected slug collision policy 'increment', got %q", cfg.Frontmatter.Defaults.SlugCollision.Policy)
	}
	if cfg.Frontmatter.Defaults.SlugCollision.MaxAttempts != 10 {
		t.Fatalf("expected slug collision maxAttempts 10, got %d", cfg.Frontmatter.Defaults.SlugCollision.MaxAttempts)
	}
	if cfg.Frontmatter.Defaults.SlugCollision.Scope != "directory" {
		t.Fatalf("expected slug collision scope 'directory', got %q", cfg.Frontmatter.Defaults.SlugCollision.Scope)
	}
}

func TestDefaultConfig_Logging(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Logging.Level != "info" {
		t.Fatalf("expected logging level 'info', got %q", cfg.Logging.Level)
	}
	if cfg.Logging.JSON {
		t.Fatal("expected logging JSON to be false by default")
	}
}
