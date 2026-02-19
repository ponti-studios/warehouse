package frontmatter

import (
	"testing"
	"time"
)

func TestMigrateContent_NoFrontmatter(t *testing.T) {
	content := "Just a note without frontmatter"
	schema := DefaultConfig().Frontmatter.Schemas["personal"]

	updated, result, err := MigrateContent(content, "/tmp/note.md", schema, StrategyFill, NewGeneratorRegistry(), time.Now())
	if err != nil {
		t.Fatalf("MigrateContent error: %v", err)
	}

	t.Logf("HasChanges: %v", result.HasChanges)
	t.Logf("Frontmatter in result: %v", result.ValidationAfter)
	t.Logf("Updated content:\n%s", updated)

	if !result.HasChanges {
		t.Error("expected HasChanges to be true")
	}

	if updated == content {
		t.Error("expected content to be updated")
	}
}

func TestMigrateContent_WithFrontmatter_Empty(t *testing.T) {
	content := "---\n\n---\nJust a note with empty frontmatter"
	schema := DefaultConfig().Frontmatter.Schemas["personal"]

	updated, result, err := MigrateContent(content, "/tmp/note.md", schema, StrategyFill, NewGeneratorRegistry(), time.Now())
	if err != nil {
		t.Fatalf("MigrateContent error: %v", err)
	}

	t.Logf("HasChanges: %v", result.HasChanges)
	t.Logf("Updated content:\n%s", updated)

	if !result.HasChanges {
		t.Error("expected HasChanges to be true")
	}
}
