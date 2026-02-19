package frontmatter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAtomicWriteFile_Basic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")

	data := []byte("hello world")
	if err := AtomicWriteFile(path, data, 0644); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(got) != "hello world" {
		t.Fatalf("expected 'hello world', got %q", string(got))
	}
}

func TestAtomicWriteFile_Overwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")

	os.WriteFile(path, []byte("original"), 0644)

	if err := AtomicWriteFile(path, []byte("updated"), 0644); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := os.ReadFile(path)
	if string(got) != "updated" {
		t.Fatalf("expected 'updated', got %q", string(got))
	}
}

func TestAtomicWriteFile_NoTempLeftOnSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")

	AtomicWriteFile(path, []byte("data"), 0644)

	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".frontmatter-") && strings.HasSuffix(e.Name(), ".tmp") {
			t.Fatalf("temp file left behind: %s", e.Name())
		}
	}
}

func TestAtomicWriteFile_NoTempLeftOnBadDir(t *testing.T) {
	err := AtomicWriteFile("/nonexistent/dir/file.md", []byte("data"), 0644)
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}

func TestAtomicWriteFile_PreservesPermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")

	AtomicWriteFile(path, []byte("data"), 0600)

	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0600 {
		t.Fatalf("expected perm 0600, got %o", info.Mode().Perm())
	}
}

func TestBackupFile_CreatesBackup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	os.WriteFile(path, []byte("original content"), 0644)

	backupPath, err := BackupFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if backupPath != path+".bak" {
		t.Fatalf("expected backup path %q, got %q", path+".bak", backupPath)
	}

	got, _ := os.ReadFile(backupPath)
	if string(got) != "original content" {
		t.Fatalf("backup content mismatch: got %q", string(got))
	}
}

func TestBackupFile_SourceMissing(t *testing.T) {
	_, err := BackupFile("/nonexistent/file.md")
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestBackupFileWithConfig_Timestamped(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.md")
	os.WriteFile(path, []byte("content"), 0644)

	cfg := BackupConfig{Enabled: true, MaxKeep: 5, Timestamp: true}
	backupPath, err := BackupFileWithConfig(path, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(backupPath, ".bak.md") {
		t.Fatalf("expected timestamped backup path with .bak.md, got %q", backupPath)
	}

	got, _ := os.ReadFile(backupPath)
	if string(got) != "content" {
		t.Fatalf("backup content mismatch")
	}
}

func TestBackupFileWithConfig_Disabled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.md")
	os.WriteFile(path, []byte("content"), 0644)

	cfg := BackupConfig{Enabled: false}
	backupPath, err := BackupFileWithConfig(path, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if backupPath != "" {
		t.Fatalf("expected empty backup path when disabled, got %q", backupPath)
	}
}

func TestBackupFileWithConfig_Retention(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.md")
	os.WriteFile(path, []byte("content"), 0644)

	os.WriteFile(filepath.Join(dir, "note.20240101T000001Z.bak.md"), []byte("old1"), 0644)
	os.WriteFile(filepath.Join(dir, "note.20240101T000002Z.bak.md"), []byte("old2"), 0644)
	os.WriteFile(filepath.Join(dir, "note.20240101T000003Z.bak.md"), []byte("old3"), 0644)

	cfg := BackupConfig{Enabled: true, MaxKeep: 2, Timestamp: true}
	_, err := BackupFileWithConfig(path, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, _ := os.ReadDir(dir)
	bakCount := 0
	for _, e := range entries {
		if strings.Contains(e.Name(), ".bak") {
			bakCount++
		}
	}
	if bakCount > 2 {
		t.Fatalf("expected at most 2 backups after pruning, got %d", bakCount)
	}
}

func TestWriteFileWithBackup_WritesAndBackups(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	os.WriteFile(path, []byte("original"), 0644)

	err := WriteFileWithBackup(path, []byte("updated"), 0644, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := os.ReadFile(path)
	if string(got) != "updated" {
		t.Fatalf("expected 'updated', got %q", string(got))
	}

	entries, _ := os.ReadDir(dir)
	hasBackup := false
	for _, e := range entries {
		if strings.Contains(e.Name(), ".bak") {
			hasBackup = true
			break
		}
	}
	if !hasBackup {
		t.Fatal("expected backup file to exist")
	}
}

func TestWriteFileWithBackup_NoBackupFlag(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	os.WriteFile(path, []byte("original"), 0644)

	WriteFileWithBackup(path, []byte("updated"), 0644, false)

	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.Contains(e.Name(), ".bak") {
			t.Fatal("expected no backup when backup=false")
		}
	}
}

func TestWriteFileWithBackup_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.md")

	err := WriteFileWithBackup(path, []byte("new content"), 0644, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := os.ReadFile(path)
	if string(got) != "new content" {
		t.Fatalf("expected 'new content', got %q", string(got))
	}
}
