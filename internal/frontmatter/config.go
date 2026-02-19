package frontmatter

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

var (
	// Candidate project-level config locations (searched from target up to fs root).
	projectConfigFiles = []string{
		".hominem/settings.yaml",
		".hominem/settings.yml",
		".hominem/settings.json",
		".hominem.yaml",
		".hominem.yml",
		".hominem.json",
	}

	// Candidate global config filenames located under $XDG_CONFIG_HOME/hominem or ~/.config/hominem.
	globalConfigFiles = []string{
		"settings.yaml",
		"settings.yml",
		"settings.json",
	}

	allowedFrontFormats = map[string]struct{}{
		"yaml": {},
		"json": {},
	}

	allowedLogLevels = map[string]struct{}{
		"debug": {},
		"info":  {},
		"warn":  {},
		"error": {},
	}

	errConfigNotFound = errors.New("frontmatter config not found")
)

//go:embed settings.schema.json
var settingsSchemaJSON []byte

var compiledConfigSchema *gojsonschema.Schema

func init() {
	// compile the embedded JSON schema at package init so we can validate loaded configs
	if len(settingsSchemaJSON) > 0 {
		loader := gojsonschema.NewBytesLoader(settingsSchemaJSON)
		schema, err := gojsonschema.NewSchema(loader)
		if err == nil {
			compiledConfigSchema = schema
		} else {
			// If schema compilation fails, we choose to panic because mispackaged schema is a developer error.
			panic(fmt.Sprintf("failed to compile frontmatter settings.schema.json: %v", err))
		}
	}
}

// LoadConfigWithOptions resolves and returns the effective configuration for the
// provided filesystem path while optionally accepting an explicit config file
// and an exclusive mode.
//
// Precedence when explicitPath != "" and exclusive == false (highest wins):
//
//	defaults <- global <- project <- explicit
//
// When exclusive == true and explicitPath != "", only the explicit file is
// loaded (merged with built-in defaults) and project/global discovery is skipped.
//
// When explicitPath == "", behavior is equivalent to LoadConfigForPath.
func LoadConfigWithOptions(target string, explicitPath string, exclusive bool) (Config, error) {
	// Start with built-in defaults.
	cfg := DefaultConfig()

	// If an explicit path is provided, load and apply it first as an overlay
	// on top of defaults. When exclusive is requested, we will skip discovery.
	if explicitPath != "" {
		explicitCfg, err := LoadExplicitConfig(explicitPath)
		if err != nil {
			return Config{}, fmt.Errorf("load explicit config %s: %w", explicitPath, err)
		}
		overlayConfig(&cfg, explicitCfg)

		if exclusive {
			// Lightweight validation of merged (defaults + explicit) config.
			if err := validateConfig(cfg); err != nil {
				return cfg, fmt.Errorf("config validation: %w", err)
			}

			// Validate against JSON Schema with the explicit file as the only source.
			if compiledConfigSchema != nil {
				if err := validateConfigJSONSchema(cfg, []string{explicitPath}); err != nil {
					return cfg, fmt.Errorf("schema validation: %w", err)
				}
			}
			return cfg, nil
		}
	}

	// Apply global config if present (overlay).
	if global, err := loadGlobalConfig(); err == nil {
		overlayConfig(&cfg, global)
	} else if !errors.Is(err, errConfigNotFound) {
		// Unexpected IO/parsing error
		return cfg, fmt.Errorf("load global config: %w", err)
	}

	// Apply nearest project config if present (overlay).
	if project, err := loadProjectConfig(target); err == nil {
		overlayConfig(&cfg, project)
	} else if target != "" && !errors.Is(err, errConfigNotFound) {
		return cfg, fmt.Errorf("load project config: %w", err)
	}

	// Lightweight validation of the merged result.
	if err := validateConfig(cfg); err != nil {
		return cfg, fmt.Errorf("config validation: %w", err)
	}

	// Collect present candidate sources for helpful diagnostics.
	presentSources := FindPresentConfigSources(target)

	// Runtime JSON-Schema validation (if schema compiled).
	if compiledConfigSchema != nil {
		if err := validateConfigJSONSchema(cfg, presentSources); err != nil {
			return cfg, fmt.Errorf("schema validation: %w", err)
		}
	}

	return cfg, nil
}

// LoadConfigForPath is the backwards-compatible convenience wrapper that loads
// the effective configuration for the provided filesystem path using the
// default non-explicit discovery rules.
func LoadConfigForPath(target string) (Config, error) {
	return LoadConfigWithOptions(target, "", false)
}

// loadGlobalConfig tries to read global config files under XDG_CONFIG_HOME/hominem
// or ~/.config/hominem. Returns errConfigNotFound when none are found.
func loadGlobalConfig() (Config, error) {
	// Prefer to consult candidate paths via helper so unit tests or other code can
	// introspect which files the loader will try.
	for _, p := range GlobalConfigPaths() {
		cfg, err := readConfigFileIfExists(p)
		if err != nil {
			// Propagate non-not-exist errors
			if !errors.Is(err, os.ErrNotExist) {
				return Config{}, err
			}
			continue
		}
		return cfg, nil
	}
	return Config{}, errConfigNotFound
}

// LoadExplicitConfig loads, validates, and returns a configuration from an explicit
// single file path provided by the caller. This skips the global/project lookup
// and is useful when the CLI `--config` flag is supplied or when a user wants to
// test a specific configuration file in isolation. The function performs the
// per-file lightweight validation and also runs the runtime JSON-Schema
// validation (when available) against the typed config.
func LoadExplicitConfig(path string) (Config, error) {
	// Read and decode the file into a typed Config (per-file checks are applied).
	cfg, err := LoadConfigFromFile(path)
	if err != nil {
		return Config{}, err
	}

	// Lightweight structural validation of the resulting typed config.
	if err := validateConfig(cfg); err != nil {
		return Config{}, fmt.Errorf("config validation: %w", err)
	}

	// Runtime JSON-Schema validation (if schema compiled).
	if compiledConfigSchema != nil {
		// Provide the single explicit path as the source so the validator can
		// attribute errors back to this file.
		if err := validateConfigJSONSchema(cfg, []string{path}); err != nil {
			return Config{}, fmt.Errorf("schema validation: %w", err)
		}
	}

	return cfg, nil
}

// GlobalConfigPaths returns the ordered list of candidate global config file
// paths that the loader will attempt to read. This helper centralizes path
// construction so unit tests and callers can inspect the exact candidates.
func GlobalConfigPaths() []string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			configDir = filepath.Join(home, ".config")
		} else {
			// Fallback to current directory if home directory is unavailable.
			configDir = "."
		}
	}
	base := filepath.Join(configDir, "hominem")
	paths := make([]string, 0, len(globalConfigFiles))
	for _, name := range globalConfigFiles {
		paths = append(paths, filepath.Join(base, name))
	}
	return paths
}

// ProjectConfigCandidates returns the list of candidate project-level config
// file paths to check for the provided target. The order is top-down from the
// nearest ancestor (or file's directory) up to the filesystem root. This helper
// is deterministic and designed to be used by the loader and unit tests.
func ProjectConfigCandidates(target string) []string {
	if target == "" {
		target = "."
	}
	abs, err := filepath.Abs(target)
	if err != nil {
		// If Abs fails, fall back to the original target string.
		abs = target
	}

	current := abs
	// If target is a file, start from its directory
	if info, err := os.Stat(current); err == nil && !info.IsDir() {
		current = filepath.Dir(current)
	}

	var candidates []string
	for {
		for _, candidate := range projectConfigFiles {
			candidates = append(candidates, filepath.Join(current, candidate))
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return candidates
}

// loadProjectConfig searches upward from target for project-level config files.
// Returns errConfigNotFound when none are found.
func loadProjectConfig(target string) (Config, error) {
	// Use the helper to compute candidate paths in search order; this lets tests
	// and other callers inspect the candidate list and simplifies mocking.
	candidates := ProjectConfigCandidates(target)
	for _, p := range candidates {
		cfg, err := readConfigFileIfExists(p)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return Config{}, err
			}
			continue
		}
		return cfg, nil
	}
	return Config{}, errConfigNotFound
}

// FindPresentConfigSources returns the list of config file paths (global then
// project candidates) that currently exist on disk. It avoids duplicates and
// preserves lookup order so callers can attribute validation failures to the
// most likely source files.
func FindPresentConfigSources(target string) []string {
	var present []string
	seen := map[string]struct{}{}

	for _, p := range GlobalConfigPaths() {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			if _, ok := seen[p]; !ok {
				present = append(present, p)
				seen[p] = struct{}{}
			}
		}
	}

	for _, p := range ProjectConfigCandidates(target) {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			if _, ok := seen[p]; !ok {
				present = append(present, p)
				seen[p] = struct{}{}
			}
		}
	}

	return present
}

// LoadConfigFromFile reads a single config file at path, performs lightweight
// validation of the raw content, and returns a typed Config. This centralizes
// per-file decoding + validation logic so callers (and tests) can operate on
// fully-validated typed configs.
func LoadConfigFromFile(path string) (Config, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	// Decode into an intermediate map first so we can validate the actual
	// contents the user provided (presence + simple semantic checks).
	var raw map[string]interface{}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		if err := json.Unmarshal(bytes, &raw); err != nil {
			return Config{}, fmt.Errorf("decode json %s: %w", path, err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(bytes, &raw); err != nil {
			return Config{}, fmt.Errorf("decode yaml %s: %w", path, err)
		}
	default:
		return Config{}, fmt.Errorf("unsupported config format for %s", path)
	}

	// Per-file lightweight validation
	if err := validateConfigMap(raw); err != nil {
		return Config{}, fmt.Errorf("invalid config file %s: %w", path, err)
	}

	// Unmarshal into the typed Config now that the raw validation passed.
	return unmarshalConfig(bytes, path)
}

/*
readConfigFileIfExists reads a config file if it exists and performs a
lightweight per-file validation before unmarshalling into the strongly-typed
`Config` struct. This ensures badly-formed or semantically-invalid config
files (e.g. version: 0 or empty `frontmatter.schemas`) are rejected early,
instead of being silently ignored when overlaying over defaults.
*/
func readConfigFileIfExists(path string) (Config, error) {
	// Delegate to LoadConfigFromFile which performs per-file decoding and
	// lightweight validation before returning a typed Config.
	return LoadConfigFromFile(path)
}

// unmarshalConfig decodes JSON or YAML into Config based on file extension.
func unmarshalConfig(bytes []byte, path string) (Config, error) {
	var cfg Config
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		if err := json.Unmarshal(bytes, &cfg); err != nil {
			return Config{}, fmt.Errorf("decode json %s: %w", path, err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(bytes, &cfg); err != nil {
			return Config{}, fmt.Errorf("decode yaml %s: %w", path, err)
		}
	default:
		return Config{}, fmt.Errorf("unsupported config format for %s", path)
	}
	return cfg, nil
}

// validateConfigMap performs lightweight validation of a raw unmarshalled map
// representing a single config file. Only checks keys that are present in the
// file are enforced here so that partial/project-level overrides are allowed,
// but obviously-broken values (e.g. version: 0 or empty schemas map) are
// rejected.
func validateConfigMap(m map[string]interface{}) error {
	// version: if present it must be numeric and >= 1
	if v, ok := m["version"]; ok {
		switch vv := v.(type) {
		case int:
			if vv < 1 {
				return fmt.Errorf("version must be >= 1")
			}
		case int64:
			if vv < 1 {
				return fmt.Errorf("version must be >= 1")
			}
		case float64:
			// JSON numbers decode to float64
			if vv < 1 {
				return fmt.Errorf("version must be >= 1")
			}
		default:
			return fmt.Errorf("version must be a number")
		}
	}

	// frontmatter: if present it must be an object; if it contains "schemas"
	// and the user provided that key, it must be an object with at least one entry.
	if fmRaw, ok := m["frontmatter"]; ok {
		fm, ok := fmRaw.(map[string]interface{})
		if !ok {
			return fmt.Errorf("frontmatter must be an object")
		}
		if schemasRaw, ok := fm["schemas"]; ok {
			switch schemas := schemasRaw.(type) {
			case map[string]interface{}:
				if len(schemas) == 0 {
					return fmt.Errorf("frontmatter.schemas must contain at least one schema")
				}
			default:
				return fmt.Errorf("frontmatter.schemas must be an object")
			}
		}
	}

	// defaults.format: if present, ensure supported value
	if defaultsRaw, ok := m["frontmatter"]; ok {
		if fm, ok := defaultsRaw.(map[string]interface{}); ok {
			if defaultsVal, ok := fm["defaults"]; ok {
				if defaultsMap, ok := defaultsVal.(map[string]interface{}); ok {
					if fmtVal, ok := defaultsMap["format"]; ok {
						if s, ok := fmtVal.(string); ok {
							ls := strings.ToLower(strings.TrimSpace(s))
							if _, ok := allowedFrontFormats[ls]; !ok {
								return fmt.Errorf("unsupported defaults.format %q", s)
							}
						} else {
							return fmt.Errorf("defaults.format must be a string")
						}
					}
				}
			}
		}
	}

	// logging.level: if present must be a supported string
	if loggingRaw, ok := m["logging"]; ok {
		if loggingMap, ok := loggingRaw.(map[string]interface{}); ok {
			if lvl, ok := loggingMap["level"]; ok {
				if s, ok := lvl.(string); ok {
					ls := strings.ToLower(strings.TrimSpace(s))
					if _, ok := allowedLogLevels[ls]; !ok {
						return fmt.Errorf("unknown logging.level %q", s)
					}
				} else {
					return fmt.Errorf("logging.level must be a string")
				}
			}
		}
	}

	return nil
}

// overlayConfig merges src into dst. Only non-zero/empty values from src will
// override dst. Maps are merged with src taking precedence for keys present.
func overlayConfig(dst *Config, src Config) {
	if src.Version != 0 {
		dst.Version = src.Version
	}

	// Merge schemas (replace or add)
	if len(src.Frontmatter.Schemas) > 0 {
		if dst.Frontmatter.Schemas == nil {
			dst.Frontmatter.Schemas = map[string]SchemaDefinition{}
		}
		for k, v := range src.Frontmatter.Schemas {
			dst.Frontmatter.Schemas[k] = v
		}
	}

	// Merge directory mapping
	if len(src.Frontmatter.DirectoryMapping) > 0 {
		if dst.Frontmatter.DirectoryMapping == nil {
			dst.Frontmatter.DirectoryMapping = map[string]string{}
		}
		for k, v := range src.Frontmatter.DirectoryMapping {
			dst.Frontmatter.DirectoryMapping[k] = v
		}
	}

	// Defaults
	if src.Frontmatter.Defaults.Format != "" {
		dst.Frontmatter.Defaults.Format = src.Frontmatter.Defaults.Format
	}
	if src.Frontmatter.Defaults.GeneratorDefaults.Timestamp != "" {
		dst.Frontmatter.Defaults.GeneratorDefaults.Timestamp = src.Frontmatter.Defaults.GeneratorDefaults.Timestamp
	}
	if src.Frontmatter.Defaults.SlugCollision.Policy != "" {
		dst.Frontmatter.Defaults.SlugCollision.Policy = src.Frontmatter.Defaults.SlugCollision.Policy
	}
	if src.Frontmatter.Defaults.SlugCollision.MaxAttempts != 0 {
		dst.Frontmatter.Defaults.SlugCollision.MaxAttempts = src.Frontmatter.Defaults.SlugCollision.MaxAttempts
	}
	if src.Frontmatter.Defaults.SlugCollision.Scope != "" {
		dst.Frontmatter.Defaults.SlugCollision.Scope = src.Frontmatter.Defaults.SlugCollision.Scope
	}

	// Logging
	if src.Logging.Level != "" {
		dst.Logging.Level = src.Logging.Level
	}
	// If any config explicitly enables JSON logging, honor it.
	if src.Logging.JSON {
		dst.Logging.JSON = true
	}
}

// validateConfig runs lightweight checks to ensure config is sane. This is
// intentionally small (no full JSON-Schema validation here).
func validateConfig(cfg Config) error {
	if cfg.Version < 1 {
		return fmt.Errorf("unsupported config version %d", cfg.Version)
	}
	if len(cfg.Frontmatter.Schemas) == 0 {
		return fmt.Errorf("no frontmatter schemas defined")
	}
	if cfg.Frontmatter.Defaults.Format != "" {
		if _, ok := allowedFrontFormats[strings.ToLower(cfg.Frontmatter.Defaults.Format)]; !ok {
			return fmt.Errorf("unsupported frontmatter format %q", cfg.Frontmatter.Defaults.Format)
		}
	}
	if cfg.Logging.Level != "" {
		if _, ok := allowedLogLevels[strings.ToLower(cfg.Logging.Level)]; !ok {
			return fmt.Errorf("unknown logging level %q", cfg.Logging.Level)
		}
	}
	// Validate individual schema validators minimally (e.g., collision policy)
	for name, s := range cfg.Frontmatter.Schemas {
		for field, vcfg := range s.Validators {
			if vcfg.Collision != nil {
				scope := strings.ToLower(vcfg.Collision.Scope)
				if scope != "" && scope != "directory" && scope != "project" && scope != "global" {
					return fmt.Errorf("schema %q validator for %q: unsupported collision scope %q", name, field, vcfg.Collision.Scope)
				}
			}
		}
	}
	return nil
}

// fieldNameMapping maps JSON Schema field paths to user-friendly names.
var fieldNameMapping = map[string]string{
	"frontmatter.defaults.format":                      "defaults.format",
	"frontmatter.defaults.generatorDefaults.timestamp": "defaults.generator.timestamp",
	"frontmatter.defaults.generatorDefaults.slug":      "defaults.generator.slug",
	"frontmatter.defaults.slugCollision.policy":        "defaults.slugCollision.policy",
	"frontmatter.defaults.slugCollision.scope":         "defaults.slugCollision.scope",
	"frontmatter.defaults.slugCollision.maxAttempts":   "defaults.slugCollision.maxAttempts",
	"frontmatter.directoryMapping":                     "directoryMapping",
	"frontmatter.schemas":                              "schemas",
	"logging.level":                                    "logging.level",
	"logging.json":                                     "logging.json",
}

// commonMistakes provides specific suggestions for frequently misconfigured fields.
var commonMistakes = map[string]string{
	"format":    "valid values are 'yaml' or 'json'",
	"level":     "valid values are 'debug', 'info', 'warn', 'error'",
	"policy":    "valid values are 'fail', 'increment', 'append-uid'",
	"scope":     "valid values are 'directory', 'project', 'global'",
	"timestamp": "valid values are 'utc', 'local', or a strftime format string",
	"slug":      "valid values are 'auto', 'kebab-title', 'kebab-filename', 'uuid-v4', 'sha1', 'ctime', 'mtime', or a custom template",
	"version":   "version must be a positive integer (e.g., 1)",
	"json":      "logging.json must be a boolean (true or false)",
}

// validateConfigJSONSchema validates the merged config against the embedded JSON Schema.
// It accepts a list of file paths that were present during the merge (in lookup order)
// to help provide source attribution guidance when validation fails.
//
// This implementation improves diagnostics by attempting to attribute each schema
// validation error to the most likely source file (from `sources`) using a
// conservative text-presence heuristic. The goal is to point users at the file
// most likely responsible for a misconfiguration.
func validateConfigJSONSchema(cfg Config, sources []string) error {
	// Marshal the config to JSON for validation
	serialized, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config for schema validation: %w", err)
	}
	loader := gojsonschema.NewBytesLoader(serialized)
	result, err := compiledConfigSchema.Validate(loader)
	if err != nil {
		return fmt.Errorf("schema validation error: %w", err)
	}
	if result.Valid() {
		return nil
	}

	// Collect structured, readable error messages: include field/pointer, description,
	// and (when possible) a best-effort source attribution indicating which config
	// file likely contributed the offending property.
	var details []string

	// Pre-read sources into memory (ignore read errors; treat unreadable files as absent).
	sourceContents := map[string]string{}
	for _, s := range sources {
		if s == "" {
			continue
		}
		if _, ok := sourceContents[s]; ok {
			continue
		}
		if b, err := os.ReadFile(s); err == nil {
			sourceContents[s] = string(b)
		} else {
			// mark unreadable files with empty content; attribution will skip them
			sourceContents[s] = ""
		}
	}

	for _, e := range result.Errors() {
		// Determine a reasonable field/pointer label for the error.
		field := e.Field()
		if field == "" {
			// Some ResultError implementations provide Context or JSON pointer.
			field = e.Context().String()
		}
		if field == "" {
			// As a final fallback use the full error string.
			field = e.String()
		}

		// Prefer description but fall back to the full string if empty.
		desc := e.Description()
		if desc == "" {
			desc = e.String()
		}

		// Heuristic attribution:
		// - Build a small set of candidate tokens to search for in each source file.
		// - Prefer longer, dotted tokens first (e.g. "frontmatter.defaults.format"),
		//   then the last path segment (e.g. "format"), and raw field names.
		var tokens []string
		tokens = append(tokens, field)
		if strings.Contains(field, ".") {
			parts := strings.Split(field, ".")
			last := parts[len(parts)-1]
			if last != "" && last != field {
				tokens = append(tokens, last)
			}
		}

		// Look through candidate sources in precedence order; the first hit is the
		// most likely contributor.
		likelySource := ""
		for _, s := range sources {
			content, ok := sourceContents[s]
			if !ok || content == "" {
				continue
			}
			lower := strings.ToLower(content)
			found := false
			for _, t := range tokens {
				t = strings.ToLower(strings.TrimSpace(t))
				if t == "" {
					continue
				}
				// Rough YAML/JSON key checks: "key:", '"key"' or "'key'".
				if strings.Contains(lower, t+":") || strings.Contains(lower, fmt.Sprintf("\"%s\"", t)) || strings.Contains(lower, fmt.Sprintf("'%s'", t)) {
					found = true
					break
				}
			}
			if found {
				likelySource = s
				break
			}
		}

		// Format the detail line including attribution when available.
		// Apply field name mapping for user-friendly display.
		displayField := field
		if mapped, ok := fieldNameMapping[field]; ok {
			displayField = mapped
		}

		// Add suggestion for common mistakes based on the last path segment.
		var suggestion string
		if strings.Contains(field, ".") {
			parts := strings.Split(field, ".")
			last := parts[len(parts)-1]
			if s, ok := commonMistakes[last]; ok {
				suggestion = fmt.Sprintf(" (%s)", s)
			}
		}

		if likelySource != "" {
			details = append(details, fmt.Sprintf("%s: %s%s (likely in %s)", displayField, desc, suggestion, likelySource))
		} else {
			details = append(details, fmt.Sprintf("%s: %s%s", displayField, desc, suggestion))
		}
	}

	// Build a helpful suggestion: show which config files were considered and
	// recommend checking them in precedence order. If no candidate files were
	// present, suggest checking the built-in defaults and any environment-level
	// config you might be using.
	var suggestion string
	if len(sources) > 0 {
		// Only include sources that actually exist on disk (preserve input order).
		var present []string
		for _, s := range sources {
			if s == "" {
				continue
			}
			if _, err := os.Stat(s); err == nil {
				present = append(present, s)
			}
		}
		if len(present) > 0 {
			suggestion = fmt.Sprintf("Check these config files (precedence order) for the invalid properties: %s", strings.Join(present, ", "))
		} else {
			suggestion = "Candidate config files were provided but none were present on disk. Verify paths and environment (XDG_CONFIG_HOME), and check built-in defaults."
		}
	} else {
		suggestion = "No project/global config files were found. Verify your environment (XDG_CONFIG_HOME) or the project-level .hominem/* files, and check the built-in defaults."
	}

	// Provide an overall message with both the per-field problems and the suggestion.
	return fmt.Errorf("config does not conform to settings.schema.json: %s. %s", strings.Join(details, "; "), suggestion)
}
