package frontmatter

// This file defines configuration types and defaults for the frontmatter system.
// It is intentionally lightweight to avoid introducing dependencies in early phases.

const DefaultConfigVersion = 1

// Config is the root configuration structure for the frontmatter system.
type Config struct {
	Version     int               `json:"version" yaml:"version" toml:"version"`
	Frontmatter FrontmatterConfig `json:"frontmatter" yaml:"frontmatter" toml:"frontmatter"`
	Logging     LoggingConfig     `json:"logging,omitempty" yaml:"logging,omitempty" toml:"logging,omitempty"`
}

// FrontmatterConfig holds schema definitions and mapping rules.
type FrontmatterConfig struct {
	Schemas          map[string]SchemaDefinition `json:"schemas" yaml:"schemas" toml:"schemas"`
	DirectoryMapping map[string]string           `json:"directory_mapping,omitempty" yaml:"directory_mapping,omitempty" toml:"directory_mapping,omitempty"`
	Defaults         FrontmatterDefaults         `json:"defaults,omitempty" yaml:"defaults,omitempty" toml:"defaults,omitempty"`
}

// FrontmatterDefaults controls global defaults for frontmatter behaviors.
type FrontmatterDefaults struct {
	Format            string              `json:"format,omitempty" yaml:"format,omitempty" toml:"format,omitempty"`
	GeneratorDefaults GeneratorDefaults   `json:"generator_defaults,omitempty" yaml:"generator_defaults,omitempty" toml:"generator_defaults,omitempty"`
	SlugCollision     SlugCollisionConfig `json:"slug_collision,omitempty" yaml:"slug_collision,omitempty" toml:"slug_collision,omitempty"`
}

// GeneratorDefaults allows global defaults for generator options.
type GeneratorDefaults struct {
	Timestamp string `json:"timestamp,omitempty" yaml:"timestamp,omitempty" toml:"timestamp,omitempty"`
}

// SlugCollisionConfig configures slug collision resolution.
type SlugCollisionConfig struct {
	Policy      string `json:"policy,omitempty" yaml:"policy,omitempty" toml:"policy,omitempty"` // fail | increment | append-uid
	MaxAttempts int    `json:"maxAttempts,omitempty" yaml:"maxAttempts,omitempty" toml:"maxAttempts,omitempty"`
	Scope       string `json:"scope,omitempty" yaml:"scope,omitempty" toml:"scope,omitempty"` // directory | project | global
}

// SchemaDefinition defines required fields, defaults, generators, and validators.
type SchemaDefinition struct {
	Required   []string                   `json:"required,omitempty" yaml:"required,omitempty" toml:"required,omitempty"`
	Defaults   map[string]string          `json:"defaults,omitempty" yaml:"defaults,omitempty" toml:"defaults,omitempty"`
	Generators map[string]GeneratorConfig `json:"generators,omitempty" yaml:"generators,omitempty" toml:"generators,omitempty"`
	Validators map[string]ValidatorConfig `json:"validators,omitempty" yaml:"validators,omitempty" toml:"validators,omitempty"`
}

// GeneratorConfig specifies how to generate a field value.
type GeneratorConfig struct {
	Name    string                 `json:"name,omitempty" yaml:"name,omitempty" toml:"name,omitempty"`
	Options map[string]interface{} `json:"options,omitempty" yaml:"options,omitempty" toml:"options,omitempty"`
}

// ValidatorConfig specifies validation rules for a field.
type ValidatorConfig struct {
	Allowed   []string               `json:"allowed,omitempty" yaml:"allowed,omitempty" toml:"allowed,omitempty"`
	Pattern   string                 `json:"pattern,omitempty" yaml:"pattern,omitempty" toml:"pattern,omitempty"`
	MinLength *int                   `json:"minLength,omitempty" yaml:"minLength,omitempty" toml:"minLength,omitempty"`
	MaxLength *int                   `json:"maxLength,omitempty" yaml:"maxLength,omitempty" toml:"maxLength,omitempty"`
	Custom    string                 `json:"custom,omitempty" yaml:"custom,omitempty" toml:"custom,omitempty"`
	Collision *SlugCollisionConfig   `json:"collision,omitempty" yaml:"collision,omitempty" toml:"collision,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty" yaml:"options,omitempty" toml:"options,omitempty"`
}

// LoggingConfig controls logging verbosity and output format.
type LoggingConfig struct {
	Level string `json:"level,omitempty" yaml:"level,omitempty" toml:"level,omitempty"`
	JSON  bool   `json:"json,omitempty" yaml:"json,omitempty" toml:"json,omitempty"`
}

// DefaultConfig returns a default configuration that mirrors the design doc.
func DefaultConfig() Config {
	return Config{
		Version: DefaultConfigVersion,
		Frontmatter: FrontmatterConfig{
			Schemas: map[string]SchemaDefinition{
				"personal": {
					Required: []string{"title", "uid", "slug", "created", "updated", "type", "status"},
					Defaults: map[string]string{
						"type":   "reference",
						"status": "draft",
					},
					Generators: map[string]GeneratorConfig{
						"uid":     {Name: "uuid-v4"},
						"slug":    {Name: "kebab-filename", Options: map[string]interface{}{"maxLength": 64}},
						"created": {Name: "file-ctime"},
						"updated": {Name: "utc-now"},
					},
					Validators: map[string]ValidatorConfig{
						"type": {
							Allowed: []string{"identity", "lifestyle", "goals", "relationships", "finance", "reference", "tracking"},
						},
						"status": {
							Allowed: []string{"draft", "published", "private", "archived"},
						},
						"slug": {
							Pattern: "^[a-z0-9][a-z0-9-]*[a-z0-9]$",
							Collision: &SlugCollisionConfig{
								Policy:      "increment",
								MaxAttempts: 10,
								Scope:       "directory",
							},
						},
						"uid": {
							Pattern: "^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$",
						},
					},
				},
				"project": {
					Required: []string{"title", "uid", "slug", "created", "updated", "status"},
					Defaults: map[string]string{
						"status": "backlog",
					},
					Validators: map[string]ValidatorConfig{
						"status": {
							Allowed: []string{"backlog", "in-progress", "review", "done"},
						},
					},
				},
				"prd": {
					Required: []string{"title", "uid", "product", "status"},
					Defaults: map[string]string{
						"status": "draft",
					},
				},
			},
			DirectoryMapping: map[string]string{
				"notebook/personal/**": "personal",
				"notebook/projects/**": "project",
				"notebook/prds/**":     "prd",
			},
			Defaults: FrontmatterDefaults{
				Format: "yaml",
				GeneratorDefaults: GeneratorDefaults{
					Timestamp: "utc-now",
				},
				SlugCollision: SlugCollisionConfig{
					Policy:      "increment",
					MaxAttempts: 10,
					Scope:       "directory",
				},
			},
		},
		Logging: LoggingConfig{
			Level: "info",
			JSON:  false,
		},
	}
}
