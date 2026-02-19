package frontmatter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type BackupConfig struct {
	Enabled   bool
	MaxKeep   int
	Timestamp bool
}

func DefaultBackupConfig() BackupConfig {
	return BackupConfig{
		Enabled:   true,
		MaxKeep:   3,
		Timestamp: true,
	}
}

func AtomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".frontmatter-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	success := false
	defer func() {
		if !success {
			tmp.Close()
			os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("fsync temp file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("chmod temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename temp to target: %w", err)
	}

	success = true
	return nil
}

func BackupFile(path string) (string, error) {
	return BackupFileWithConfig(path, BackupConfig{Enabled: true, MaxKeep: 1, Timestamp: false})
}

func BackupFileWithConfig(path string, cfg BackupConfig) (string, error) {
	if !cfg.Enabled {
		return "", nil
	}

	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("source file not found: %w", err)
	}

	var backupPath string
	if cfg.Timestamp {
		ts := time.Now().UTC().Format("20060102T150405Z")
		ext := filepath.Ext(path)
		base := strings.TrimSuffix(path, ext)
		backupPath = fmt.Sprintf("%s.%s.bak%s", base, ts, ext)
	} else {
		backupPath = path + ".bak"
	}

	src, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open source for backup: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("create backup file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("copy to backup: %w", err)
	}

	if err := dst.Sync(); err != nil {
		return "", fmt.Errorf("fsync backup: %w", err)
	}

	if cfg.MaxKeep > 0 {
		pruneBackups(path, cfg.MaxKeep)
	}

	return backupPath, nil
}

func pruneBackups(originalPath string, maxKeep int) {
	dir := filepath.Dir(originalPath)
	base := filepath.Base(originalPath)
	ext := filepath.Ext(base)
	nameNoExt := strings.TrimSuffix(base, ext)

	pattern1 := nameNoExt + ".*.bak" + ext
	pattern2 := base + ".bak"

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	var backups []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		matched1, _ := filepath.Match(pattern1, name)
		if matched1 || name == filepath.Base(pattern2) {
			backups = append(backups, filepath.Join(dir, name))
		}
	}

	if len(backups) <= maxKeep {
		return
	}

	sort.Strings(backups)
	toRemove := backups[:len(backups)-maxKeep]
	for _, path := range toRemove {
		os.Remove(path)
	}
}

func WriteFileWithBackup(path string, data []byte, perm os.FileMode, backup bool) error {
	if backup {
		if _, err := os.Stat(path); err == nil {
			bcfg := DefaultBackupConfig()
			if _, err := BackupFileWithConfig(path, bcfg); err != nil {
				return fmt.Errorf("backup before write: %w", err)
			}
		}
	}
	return AtomicWriteFile(path, data, perm)
}
