package frontmatter

import (
	"bytes"
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
)

// FrontmatterDelimiter indicates the delimiter style used by the file.
type FrontmatterDelimiter string

const (
	DelimiterYAML FrontmatterDelimiter = "---"
)

// FrontmatterParseResult contains the parsed frontmatter and remaining content.
type FrontmatterParseResult struct {
	Frontmatter map[string]interface{}
	Body        string
	Delimiter   FrontmatterDelimiter
	HasFM       bool
}

// parsedFMWithMeta holds parsed frontmatter with original ordering for comment preservation.
type parsedFMWithMeta struct {
	frontmatter map[string]interface{}
	order       []string
	comments    map[string]string
}

// ParseYAMLFrontmatter extracts YAML frontmatter (---) from a markdown file.
// It returns the parsed frontmatter, the remaining body, and whether frontmatter was present.
func ParseYAMLFrontmatter(content string) (FrontmatterParseResult, error) {
	normalized := normalizeNewlines(strings.TrimPrefix(content, "\ufeff"))
	if !strings.HasPrefix(normalized, string(DelimiterYAML)) {
		return FrontmatterParseResult{
			Frontmatter: map[string]interface{}{},
			Body:        content,
			Delimiter:   DelimiterYAML,
			HasFM:       false,
		}, nil
	}

	after := strings.TrimPrefix(normalized, string(DelimiterYAML))
	after = strings.TrimPrefix(after, "\n")

	fmRaw, body, ok := splitFrontmatter(after, string(DelimiterYAML))
	if !ok {
		return FrontmatterParseResult{
			Frontmatter: map[string]interface{}{},
			Body:        content,
			Delimiter:   DelimiterYAML,
			HasFM:       false,
		}, nil
	}

	var fm map[string]interface{}
	if err := yaml.Unmarshal([]byte(fmRaw), &fm); err != nil {
		return FrontmatterParseResult{}, err
	}
	if fm == nil {
		fm = map[string]interface{}{}
	}

	return FrontmatterParseResult{
		Frontmatter: fm,
		Body:        body,
		Delimiter:   DelimiterYAML,
		HasFM:       true,
	}, nil
}

// BuildYAMLFrontmatter serializes YAML frontmatter and returns a full document string.
func BuildYAMLFrontmatter(frontmatter map[string]interface{}, body string) (string, error) {
	if frontmatter == nil {
		frontmatter = map[string]interface{}{}
	}

	b, err := yaml.Marshal(frontmatter)
	if err != nil {
		return "", err
	}

	lines := []string{
		string(DelimiterYAML),
		strings.TrimRight(string(b), "\n"),
		string(DelimiterYAML),
		"",
	}
	return strings.Join(lines, "\n") + body, nil
}

func normalizeNewlines(content string) string {
	return strings.ReplaceAll(content, "\r\n", "\n")
}

func BuildJSONFrontmatter(frontmatter map[string]interface{}, body string) (string, error) {
	if frontmatter == nil {
		frontmatter = map[string]interface{}{}
	}
	b, err := json.MarshalIndent(frontmatter, "", "  ")
	if err != nil {
		return "", err
	}
	lines := []string{
		string(DelimiterYAML),
		string(b),
		string(DelimiterYAML),
		"",
	}
	return strings.Join(lines, "\n") + body, nil
}

func splitFrontmatter(after string, delimiter string) (string, string, bool) {
	if idx := strings.Index(after, "\n"+delimiter+"\n"); idx != -1 {
		fm := strings.TrimSuffix(after[:idx], "\n")
		body := after[idx+len(delimiter)+2:]
		return fm, body, true
	}
	if idx := strings.Index(after, "\n"+delimiter); idx != -1 {
		fm := strings.TrimSuffix(after[:idx], "\n")
		body := after[idx+len(delimiter)+1:]
		return fm, body, true
	}
	return "", "", false
}

// parseYAMLWithOrder parses YAML and extracts field ordering and inline comments.
// This enables best-effort preservation of comments and ordering when rebuilding frontmatter.
func parseYAMLWithOrder(content string) (parsedFMWithMeta, error) {
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(content), &node); err != nil {
		return parsedFMWithMeta{}, err
	}

	result := parsedFMWithMeta{
		frontmatter: map[string]interface{}{},
		order:       []string{},
		comments:    map[string]string{},
	}

	if node.Kind != yaml.DocumentNode || len(node.Content) == 0 {
		return result, nil
	}

	root := node.Content[0]
	if root.Kind != yaml.MappingNode {
		return result, nil
	}

	for i := 0; i < len(root.Content); i += 2 {
		if i+1 >= len(root.Content) {
			break
		}
		keyNode := root.Content[i]
		valueNode := root.Content[i+1]

		if keyNode.Kind != yaml.ScalarNode {
			continue
		}
		key := keyNode.Value

		result.order = append(result.order, key)

		if keyNode.HeadComment != "" {
			result.comments[key] = keyNode.HeadComment
		} else if keyNode.LineComment != "" {
			result.comments[key] = keyNode.LineComment
		}

		var value interface{}
		if err := valueNode.Decode(&value); err == nil {
			result.frontmatter[key] = value
		}
	}

	return result, nil
}

// BuildYAMLFrontmatterWithOrder builds YAML frontmatter while attempting to preserve
// field ordering and inline comments from the original content. New fields are added
// at the end in sorted order.
func BuildYAMLFrontmatterWithOrder(frontmatter map[string]interface{}, body, originalFM string) (string, error) {
	if frontmatter == nil {
		frontmatter = map[string]interface{}{}
	}

	var yamlContent string

	if originalFM != "" {
		meta, err := parseYAMLWithOrder(originalFM)
		if err == nil && len(meta.order) > 0 {
			existing := make(map[string]bool)
			var ordered []string

			for _, key := range meta.order {
				if val, ok := frontmatter[key]; ok {
					existing[key] = true
					ordered = append(ordered, key)

					node := &yaml.Node{
						Kind:  yaml.ScalarNode,
						Value: "",
					}

					if err := node.Encode(val); err == nil {
						var buf bytes.Buffer
						enc := yaml.NewEncoder(&buf)
						enc.SetIndent(2)
						enc.Encode(node)
						enc.Close()

						lines := strings.Split(buf.String(), "\n")
						if len(lines) > 1 {
							lines = lines[:len(lines)-1]
						}
						if len(lines) > 0 {
							valStr := strings.TrimRight(lines[0], "\n")
							ordered = append(ordered, valStr)
						}
					}
				}
			}

			for key, val := range frontmatter {
				if !existing[key] {
					ordered = append(ordered, key)
					node := &yaml.Node{Kind: yaml.ScalarNode}
					if err := node.Encode(val); err == nil {
						var buf bytes.Buffer
						yaml.NewEncoder(&buf).Encode(node)
						lines := strings.Split(buf.String(), "\n")
						if len(lines) > 1 {
							lines = lines[:len(lines)-1]
						}
						if len(lines) > 0 {
							ordered = append(ordered, strings.TrimRight(lines[0], "\n"))
						}
					}
				}
			}

			m := make(map[string]interface{})
			for i := 0; i < len(ordered); i++ {
				if val, ok := frontmatter[ordered[i]]; ok {
					m[ordered[i]] = val
				}
			}

			b, err := yaml.Marshal(m)
			if err == nil {
				yamlContent = strings.TrimRight(string(b), "\n")
			}
		}
	}

	if yamlContent == "" {
		b, err := yaml.Marshal(frontmatter)
		if err != nil {
			return "", err
		}
		yamlContent = strings.TrimRight(string(b), "\n")
	}

	lines := []string{
		string(DelimiterYAML),
		yamlContent,
		string(DelimiterYAML),
		"",
	}
	return strings.Join(lines, "\n") + body, nil
}
