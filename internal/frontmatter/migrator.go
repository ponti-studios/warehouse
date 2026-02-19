package frontmatter

import (
	"fmt"
	"path/filepath"
	"time"
)

// MigrationStrategy controls how frontmatter updates are applied.
type MigrationStrategy string

const (
	StrategyFill       MigrationStrategy = "fill"
	StrategyRepair     MigrationStrategy = "repair"
	StrategyOverwrite  MigrationStrategy = "overwrite"
	StrategyTimestamps MigrationStrategy = "timestamps"
)

// FieldChange records a single field update.
type FieldChange struct {
	Before string
	After  string
	Reason string
}

// MigrationResult summarizes the outcome of a migration.
type MigrationResult struct {
	Strategy         MigrationStrategy
	HasFrontmatter   bool
	HasChanges       bool
	Changes          map[string]FieldChange
	ValidationBefore ValidationResult
	ValidationAfter  ValidationResult
}

// ValidateContent parses frontmatter and validates it against a schema.
func ValidateContent(content string, schema SchemaDefinition) (ValidationResult, error) {
	parsed, err := ParseYAMLFrontmatter(content)
	if err != nil {
		return ValidationResult{}, err
	}
	return ValidateFrontmatter(parsed.Frontmatter, schema), nil
}

// MigrateContent parses, updates, validates, and returns updated content.
func MigrateContent(content, filePath string, schema SchemaDefinition, strategy MigrationStrategy, registry *GeneratorRegistry, now time.Time) (string, MigrationResult, error) {
	if registry == nil {
		registry = NewGeneratorRegistry()
	}

	parsed, err := ParseYAMLFrontmatter(content)
	if err != nil {
		return "", MigrationResult{}, err
	}

	frontmatter := copyFrontmatter(parsed.Frontmatter)
	result := MigrationResult{
		Strategy:       strategy,
		HasFrontmatter: parsed.HasFM,
		Changes:        map[string]FieldChange{},
	}

	result.ValidationBefore = ValidateFrontmatter(frontmatter, schema)

	frontmatter, changed, err := applyStrategy(frontmatter, filePath, schema, strategy, registry, now, result.Changes)
	if err != nil {
		return "", result, err
	}

	result.ValidationAfter = ValidateFrontmatter(frontmatter, schema)
	result.HasChanges = changed

	if !changed {
		return content, result, nil
	}

	updated, err := BuildYAMLFrontmatter(frontmatter, parsed.Body)
	if err != nil {
		return "", result, err
	}

	return updated, result, nil
}

func applyStrategy(frontmatter map[string]interface{}, filePath string, schema SchemaDefinition, strategy MigrationStrategy, registry *GeneratorRegistry, now time.Time, changes map[string]FieldChange) (map[string]interface{}, bool, error) {
	if frontmatter == nil || len(frontmatter) == 0 {
		frontmatter = map[string]interface{}{}
	}

	ctx := GeneratorContext{
		FilePath: filePath,
		FileName: filepath.Base(filePath),
		Title:    getString(frontmatter, "title"),
		Now:      now,
	}

	changed := false
	setField := func(field, value, reason string) {
		before := fmt.Sprint(frontmatter[field])
		if before == value {
			return
		}
		frontmatter[field] = value
		changes[field] = FieldChange{
			Before: before,
			After:  value,
			Reason: reason,
		}
		changed = true
	}

	valueForField := func(field string) (string, bool, error) {
		if genCfg, ok := schema.Generators[field]; ok {
			gen, ok := registry.Get(genCfg.Name)
			if !ok {
				return "", false, fmt.Errorf("unknown generator %q for field %q", genCfg.Name, field)
			}
			ctx.FieldName = field
			val, err := gen.Generate(ctx)
			if err != nil {
				return "", false, err
			}
			return val, true, nil
		}
		if def, ok := schema.Defaults[field]; ok {
			return def, true, nil
		}
		return "", false, nil
	}

	applyMissing := func(field, reason string) error {
		if _, exists := frontmatter[field]; exists {
			return nil
		}
		val, ok, err := valueForField(field)
		if err != nil {
			return err
		}
		if ok {
			setField(field, val, reason)
		}
		return nil
	}

	switch strategy {
	case StrategyOverwrite:
		for field := range schema.Defaults {
			val, ok, err := valueForField(field)
			if err != nil {
				return frontmatter, false, err
			}
			if ok {
				setField(field, val, "overwrite")
			}
		}
		for field := range schema.Generators {
			val, ok, err := valueForField(field)
			if err != nil {
				return frontmatter, false, err
			}
			if ok {
				setField(field, val, "overwrite")
			}
		}
		for _, field := range schema.Required {
			val, ok, err := valueForField(field)
			if err != nil {
				return frontmatter, false, err
			}
			if ok {
				setField(field, val, "overwrite")
			}
		}

	case StrategyTimestamps:
		if err := applyMissing("created", "timestamps"); err != nil {
			return frontmatter, false, err
		}
		val, ok, err := valueForField("updated")
		if err != nil {
			return frontmatter, false, err
		}
		if ok {
			setField("updated", val, "timestamps")
		}

	case StrategyRepair, StrategyFill:
		for _, field := range schema.Required {
			if err := applyMissing(field, "missing required field"); err != nil {
				return frontmatter, false, err
			}
		}
		for field := range schema.Defaults {
			if err := applyMissing(field, "default"); err != nil {
				return frontmatter, false, err
			}
		}

		if strategy == StrategyRepair {
			validation := ValidateFrontmatter(frontmatter, schema)
			for _, err := range validation.Errors {
				val, ok, genErr := valueForField(err.Field)
				if genErr != nil {
					return frontmatter, false, genErr
				}
				if ok {
					setField(err.Field, val, "repair")
				}
			}
		}

		if changed {
			val, ok, err := valueForField("updated")
			if err != nil {
				return frontmatter, false, err
			}
			if ok {
				setField("updated", val, "update timestamp")
			}
		}
	default:
		return frontmatter, false, fmt.Errorf("unknown migration strategy: %s", strategy)
	}

	return frontmatter, changed, nil
}

func copyFrontmatter(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return map[string]interface{}{}
	}
	out := make(map[string]interface{}, len(input))
	for k, v := range input {
		out[k] = v
	}
	return out
}

func getString(values map[string]interface{}, key string) string {
	if v, ok := values[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func InitContent(filePath string, schema SchemaDefinition, registry *GeneratorRegistry, now time.Time, fmFormat string) (string, error) {
	if registry == nil {
		registry = NewGeneratorRegistry()
	}

	fm := map[string]interface{}{}

	ctx := GeneratorContext{
		FilePath: filePath,
		FileName: filepath.Base(filePath),
		Now:      now,
	}

	for field := range schema.Defaults {
		fm[field] = schema.Defaults[field]
	}

	for field, genCfg := range schema.Generators {
		gen, ok := registry.Get(genCfg.Name)
		if !ok {
			continue
		}
		ctx.FieldName = field
		val, err := gen.Generate(ctx)
		if err != nil {
			continue
		}
		fm[field] = val
	}

	for _, field := range schema.Required {
		if _, ok := fm[field]; !ok {
			if def, ok := schema.Defaults[field]; ok {
				fm[field] = def
			}
		}
	}

	if fmFormat == "json" {
		return BuildJSONFrontmatter(fm, "")
	}
	return BuildYAMLFrontmatter(fm, "")
}
