package frontmatter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// helper to write a file and fail the test on error
func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir failed for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write failed for %s: %v", path, err)
	}
}

func TestLoadConfigForPath_NoConfigFiles_ReturnsDefaults(t *testing.T) {
	td := t.TempDir()

	cfg, err := LoadConfigForPath(td)
	if err != nil {
		t.Fatalf("expected no error loading defaults; got: %v", err)
	}

	def := DefaultConfig()

	if cfg.Version != def.Version {
		t.Fatalf("expected version %d, got %d", def.Version, cfg.Version)
	}

	if cfg.Frontmatter.Defaults.Format != def.Frontmatter.Defaults.Format {
		t.Fatalf("expected default format %q, got %q", def.Frontmatter.Defaults.Format, cfg.Frontmatter.Defaults.Format)
	}

	// ensure at least one builtin schema is present (DefaultConfig provides these)
	if len(cfg.Frontmatter.Schemas) == 0 {
		t.Fatalf("expected at least one schema in default config")
	}
}

func TestLoadConfigForPath_GlobalOverrideApplied(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)

	globalDir := filepath.Join(xdg, "hominem")
	if err := os.MkdirAll(globalDir, 0o755); err != nil {
		t.Fatalf("failed to mkdir %s: %v", globalDir, err)
	}

	globalCfgPath := filepath.Join(globalDir, "settings.yaml")
	globalYAML := `
version: 1
frontmatter:
  schemas:
    custom:
      required: [title]
  defaults:
    format: json
logging:
  level: debug
`
	writeFile(t, globalCfgPath, globalYAML)

	// call loader with an unrelated working dir
	td := t.TempDir()
	cfg, err := LoadConfigForPath(td)
	if err != nil {
		t.Fatalf("expected no error loading global override: %v", err)
	}

	if cfg.Frontmatter.Defaults.Format != "json" {
		t.Fatalf("expected global override format 'json', got %q", cfg.Frontmatter.Defaults.Format)
	}
	if cfg.Logging.Level != "debug" {
		t.Fatalf("expected logging.level 'debug', got %q", cfg.Logging.Level)
	}
	if _, ok := cfg.Frontmatter.Schemas["custom"]; !ok {
		t.Fatalf("expected global schema 'custom' to be present")
	}
}

func TestLoadConfigForPath_ProjectOverrideTakesPrecedence(t *testing.T) {
	root := t.TempDir()
	// nested path where loader should be invoked
	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	// create project-level config at root (ancestor)
	projectCfgPath := filepath.Join(root, ".hominem", "settings.yaml")
	projectYAML := `
version: 1
frontmatter:
  schemas:
    projectSchema:
      required: [title]
  defaults:
    format: json
logging:
  level: info
`
	writeFile(t, projectCfgPath, projectYAML)

	// Also place a global config that sets format to yaml to ensure project wins.
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	globalCfgPath := filepath.Join(xdg, "hominem", "settings.yaml")
	if err := os.MkdirAll(filepath.Dir(globalCfgPath), 0o755); err != nil {
		t.Fatalf("failed to mkdir global config dir: %v", err)
	}
	writeFile(t, globalCfgPath, `
version: 1
frontmatter:
  defaults:
    format: yaml
`)

	cfg, err := LoadConfigForPath(nested)
	if err != nil {
		t.Fatalf("expected no error loading merged configs: %v", err)
	}

	// project config should override global YAML -> JSON
	if cfg.Frontmatter.Defaults.Format != "json" {
		t.Fatalf("expected project override format 'json', got %q", cfg.Frontmatter.Defaults.Format)
	}
	if _, ok := cfg.Frontmatter.Schemas["projectSchema"]; !ok {
		t.Fatalf("expected project schema 'projectSchema' to be present")
	}
}

func TestLoadConfigForPath_InvalidConfigReported(t *testing.T) {
	root := t.TempDir()
	// write an invalid config (version too low and empty schemas)
	invalidCfgPath := filepath.Join(root, ".hominem", "settings.yaml")
	invalidYAML := `
version: 0
frontmatter:
  schemas: {}
`
	writeFile(t, invalidCfgPath, invalidYAML)

	_, err := LoadConfigForPath(root)
	if err == nil {
		t.Fatalf("expected error when loading invalid config, got nil")
	}

	// error should mention version or schema mismatch (allow either message)
	msg := err.Error()
	if !(strings.Contains(msg, "version") || strings.Contains(msg, "schema") || strings.Contains(msg, "no frontmatter schemas")) {
		t.Fatalf("error message not helpful: %v", msg)
	}
}

func TestLoadConfigWithOptions_ExplicitOverridesProjectAndGlobal(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "nested", "dir")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	// global config
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	globalCfgPath := filepath.Join(xdg, "hominem", "settings.yaml")
	if err := os.MkdirAll(filepath.Dir(globalCfgPath), 0o755); err != nil {
		t.Fatalf("mkdir global config dir: %v", err)
	}
	writeFile(t, globalCfgPath, `
version: 1
frontmatter:
  schemas:
    globalSchema:
      required: [title]
  defaults:
    format: yaml
logging:
  level: debug
`)

	// project config
	projectCfgPath := filepath.Join(root, ".hominem", "settings.yaml")
	writeFile(t, projectCfgPath, `
version: 1
frontmatter:
  schemas:
    projectSchema:
      required: [title]
  defaults:
    format: json
`)

	// explicit config (highest precedence)
	explicitPath := filepath.Join(root, "explicit.yaml")
	writeFile(t, explicitPath, `
version: 1
frontmatter:
  schemas:
    explicitSchema:
      required: [title]
  defaults:
    format: json
`)

	cfg, err := LoadConfigWithOptions(nested, explicitPath, false)
	if err != nil {
		t.Fatalf("expected no error loading explicit + discovery: %v", err)
	}

	if cfg.Frontmatter.Defaults.Format != "json" {
		t.Fatalf("expected explicit format 'json', got %q", cfg.Frontmatter.Defaults.Format)
	}
	if cfg.Logging.Level != "debug" {
		t.Fatalf("expected global logging level 'debug', got %q", cfg.Logging.Level)
	}
	if _, ok := cfg.Frontmatter.Schemas["explicitSchema"]; !ok {
		t.Fatalf("expected explicit schema present")
	}
	if _, ok := cfg.Frontmatter.Schemas["projectSchema"]; !ok {
		t.Fatalf("expected project schema present")
	}
	if _, ok := cfg.Frontmatter.Schemas["globalSchema"]; !ok {
		t.Fatalf("expected global schema present")
	}
}

func TestLoadConfigWithOptions_ExplicitExclusiveSkipsDiscovery(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	globalCfgPath := filepath.Join(xdg, "hominem", "settings.yaml")
	if err := os.MkdirAll(filepath.Dir(globalCfgPath), 0o755); err != nil {
		t.Fatalf("mkdir global config dir: %v", err)
	}
	writeFile(t, globalCfgPath, `
version: 1
frontmatter:
  defaults:
    format: json
logging:
  level: debug
`)

	projectCfgPath := filepath.Join(root, ".hominem", "settings.yaml")
	writeFile(t, projectCfgPath, `
version: 1
frontmatter:
  defaults:
    format: yaml
`)

	explicitPath := filepath.Join(root, "explicit.yaml")
	writeFile(t, explicitPath, `
version: 1
frontmatter:
  schemas:
    explicitSchema:
      required: [title]
  defaults:
    format: json
`)

	cfg, err := LoadConfigWithOptions(nested, explicitPath, true)
	if err != nil {
		t.Fatalf("expected no error loading explicit-only config: %v", err)
	}

	if cfg.Frontmatter.Defaults.Format != "json" {
		t.Fatalf("expected explicit-only format 'json', got %q", cfg.Frontmatter.Defaults.Format)
	}
	// Default logging level comes from DefaultConfig(); explicit-only mode should
	// not inherit global/project logging. Ensure it's whatever the default config
	// specifies (typically "info").
	if cfg.Logging.Level != DefaultConfig().Logging.Level {
		t.Fatalf("expected logging level to remain default %q when exclusive, got %q", DefaultConfig().Logging.Level, cfg.Logging.Level)
	}
}

func TestLoadConfigWithOptions_MissingExplicitFileFails(t *testing.T) {
	root := t.TempDir()
	missing := filepath.Join(root, "missing.yaml")

	if _, err := LoadConfigWithOptions(root, missing, false); err == nil {
		t.Fatalf("expected error for missing explicit config file")
	}
}
