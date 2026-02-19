package frontmatter

import (
	"testing"
)

func testConfig() Config {
	return Config{
		Version: 1,
		Frontmatter: FrontmatterConfig{
			Schemas: map[string]SchemaDefinition{
				"personal": {Required: []string{"title", "uid", "slug"}},
				"project":  {Required: []string{"title", "status"}},
				"prd":      {Required: []string{"title", "product"}},
			},
			DirectoryMapping: map[string]string{
				"notebook/personal/**": "personal",
				"notebook/projects/**": "project",
				"notebook/prds/**":     "prd",
			},
		},
	}
}

func TestResolveSchema_ExplicitName(t *testing.T) {
	cfg := testConfig()
	schema, name, err := ResolveSchema(cfg, "project", "some/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "project" {
		t.Fatalf("expected name 'project', got %q", name)
	}
	if len(schema.Required) == 0 {
		t.Fatal("expected non-empty schema")
	}
}

func TestResolveSchema_ExplicitName_Unknown(t *testing.T) {
	cfg := testConfig()
	_, _, err := ResolveSchema(cfg, "nonexistent", "some/path")
	if err == nil {
		t.Fatal("expected error for unknown explicit schema")
	}
}

func TestResolveSchema_DirectoryMapping_Personal(t *testing.T) {
	cfg := testConfig()
	_, name, err := ResolveSchema(cfg, "", "notebook/personal/notes/foo.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "personal" {
		t.Fatalf("expected 'personal' from directory mapping, got %q", name)
	}
}

func TestResolveSchema_DirectoryMapping_Project(t *testing.T) {
	cfg := testConfig()
	_, name, err := ResolveSchema(cfg, "", "notebook/projects/my-app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "project" {
		t.Fatalf("expected 'project' from directory mapping, got %q", name)
	}
}

func TestResolveSchema_DirectoryMapping_PRD(t *testing.T) {
	cfg := testConfig()
	_, name, err := ResolveSchema(cfg, "", "notebook/prds/feature-x.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "prd" {
		t.Fatalf("expected 'prd' from directory mapping, got %q", name)
	}
}

func TestResolveSchema_LongestPrefixWins(t *testing.T) {
	cfg := Config{
		Version: 1,
		Frontmatter: FrontmatterConfig{
			Schemas: map[string]SchemaDefinition{
				"general": {Required: []string{"title"}},
				"deep":    {Required: []string{"title", "depth"}},
			},
			DirectoryMapping: map[string]string{
				"docs/**":          "general",
				"docs/deep/sub/**": "deep",
			},
		},
	}

	_, name, err := ResolveSchema(cfg, "", "docs/deep/sub/nested/file.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "deep" {
		t.Fatalf("expected longest prefix match 'deep', got %q", name)
	}
}

func TestResolveSchema_FallbackToPersonal(t *testing.T) {
	cfg := testConfig()
	_, name, err := ResolveSchema(cfg, "", "unrelated/path/file.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "personal" {
		t.Fatalf("expected fallback to 'personal', got %q", name)
	}
}

func TestResolveSchema_FallbackToFirst(t *testing.T) {
	cfg := Config{
		Version: 1,
		Frontmatter: FrontmatterConfig{
			Schemas: map[string]SchemaDefinition{
				"only-schema": {Required: []string{"title"}},
			},
		},
	}

	_, name, err := ResolveSchema(cfg, "", "any/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "only-schema" {
		t.Fatalf("expected fallback to first schema 'only-schema', got %q", name)
	}
}

func TestResolveSchema_NoSchemas(t *testing.T) {
	cfg := Config{
		Version: 1,
		Frontmatter: FrontmatterConfig{
			Schemas: map[string]SchemaDefinition{},
		},
	}

	_, _, err := ResolveSchema(cfg, "", "any/path")
	if err == nil {
		t.Fatal("expected error for empty schemas")
	}
}
