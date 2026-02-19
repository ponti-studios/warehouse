package frontmatter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ResolveSchema(cfg Config, schemaName string, target string) (SchemaDefinition, string, error) {
	if schemaName != "" {
		schema, ok := cfg.Frontmatter.Schemas[schemaName]
		if !ok {
			return SchemaDefinition{}, "", fmt.Errorf("unknown schema: %s", schemaName)
		}
		return schema, schemaName, nil
	}

	normTarget := filepath.Clean(target)
	if fi, err := os.Stat(normTarget); err == nil && !fi.IsDir() {
		normTarget = filepath.Dir(normTarget)
	}

	longestMatch := ""
	var chosenSchema string
	for pattern, schemaNameCandidate := range cfg.Frontmatter.DirectoryMapping {
		prefix := pattern
		if strings.HasSuffix(prefix, "/**") {
			prefix = strings.TrimSuffix(prefix, "/**")
		}
		prefix = filepath.Clean(prefix)
		if prefix == "" {
			continue
		}
		if strings.HasPrefix(normTarget, prefix) {
			if len(prefix) > len(longestMatch) {
				longestMatch = prefix
				chosenSchema = schemaNameCandidate
			}
		}
	}

	if chosenSchema != "" {
		if schema, ok := cfg.Frontmatter.Schemas[chosenSchema]; ok {
			return schema, chosenSchema, nil
		}
	}

	if schema, ok := cfg.Frontmatter.Schemas["personal"]; ok {
		return schema, "personal", nil
	}

	for name, schema := range cfg.Frontmatter.Schemas {
		return schema, name, nil
	}

	return SchemaDefinition{}, "", fmt.Errorf("no schemas configured")
}
