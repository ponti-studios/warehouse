package frontmatter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectSlugCollisions_DirectoryScope(t *testing.T) {
	tmp := t.TempDir()

	subdir := filepath.Join(tmp, "notes")
	os.MkdirAll(subdir, 0755)

	createFile(t, filepath.Join(subdir, "note1.md"), "---\nslug: hello\n---\n")
	createFile(t, filepath.Join(subdir, "note2.md"), "---\nslug: hello\n---\n")
	createFile(t, filepath.Join(tmp, "other.md"), "---\nslug: hello\n---\n")

	results, err := DetectSlugCollisions(tmp, "directory", WalkOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected collisions in directory scope")
	}

	for _, r := range results {
		if len(r.Collisions) == 0 {
			t.Errorf("expected collisions for %s", r.Path)
		}
	}
}

func TestDetectSlugCollisions_ProjectScope(t *testing.T) {
	tmp := t.TempDir()

	subdir1 := filepath.Join(tmp, "a")
	subdir2 := filepath.Join(tmp, "b")
	os.MkdirAll(subdir1, 0755)
	os.MkdirAll(subdir2, 0755)

	createFile(t, filepath.Join(subdir1, "note1.md"), "---\nslug: hello\n---\n")
	createFile(t, filepath.Join(subdir2, "note2.md"), "---\nslug: hello\n---\n")

	results, err := DetectSlugCollisions(tmp, "project", WalkOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 collision results, got %d", len(results))
	}
}

func TestDetectSlugCollisions_NoSlug(t *testing.T) {
	tmp := t.TempDir()
	createFile(t, filepath.Join(tmp, "note.md"), "---\ntitle: No slug\n---\n")

	results, err := DetectSlugCollisions(tmp, "directory", WalkOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected no collisions, got %d", len(results))
	}
}

func TestDetectSlugCollisions_EmptyDir(t *testing.T) {
	tmp := t.TempDir()

	results, err := DetectSlugCollisions(tmp, "directory", WalkOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected no collisions, got %d", len(results))
	}
}

func TestResolveSlugCollision_FailPolicy(t *testing.T) {
	existing := map[string]bool{"hello": true}

	result, err := ResolveSlugCollision("hello", existing, "fail", 5)
	if err == nil {
		t.Fatal("expected error for fail policy")
	}

	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}

	_, ok := err.(*SlugCollisionError)
	if !ok {
		t.Errorf("expected SlugCollisionError, got %T", err)
	}
}

func TestResolveSlugCollision_IncrementPolicy(t *testing.T) {
	existing := map[string]bool{
		"hello":    true,
		"hello-2":  true,
		"hello-3":  true,
		"hello-10": true,
	}

	result, err := ResolveSlugCollision("hello", existing, "increment", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "hello-4" {
		t.Errorf("expected hello-4, got %q", result)
	}
}

func TestResolveSlugCollision_IncrementPolicy_Exhausted(t *testing.T) {
	existing := map[string]bool{
		"hello":   true,
		"hello-2": true,
		"hello-3": true,
	}

	result, err := ResolveSlugCollision("hello", existing, "increment", 2)
	if err == nil {
		t.Fatal("expected error for exhausted increments")
	}

	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}
}

func TestResolveSlugCollision_AppendUIDPolicy(t *testing.T) {
	existing := map[string]bool{"hello": true}

	result, err := ResolveSlugCollision("hello", existing, "append-uid", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == "hello" {
		t.Error("expected slug to be modified")
	}

	if len(result) <= len("hello") {
		t.Errorf("expected uid to be appended, got %q", result)
	}
}

func TestResolveSlugCollision_NoCollision(t *testing.T) {
	existing := map[string]bool{"hello": true}

	result, err := ResolveSlugCollision("world", existing, "fail", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "world" {
		t.Errorf("expected world, got %q", result)
	}
}

func TestResolveSlugCollision_UnknownPolicy(t *testing.T) {
	existing := map[string]bool{"hello": true}

	_, err := ResolveSlugCollision("hello", existing, "invalid", 5)
	if err == nil {
		t.Fatal("expected error for unknown policy")
	}
}

func TestSlugCollisionError_Error(t *testing.T) {
	err := SlugCollisionError{
		Slug:       "hello",
		Path:       "/path/to/file.md",
		Collisions: []string{"/path/other.md", "/path/another.md"},
	}

	expected := `slug "hello" in /path/to/file.md collides with: /path/other.md, /path/another.md`
	if err.Error() != expected {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func createFile(t *testing.T, path, content string) {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
}
