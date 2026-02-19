package frontmatter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SlugCollisionResult struct {
	Slug       string
	Path       string
	Collisions []string
}

type SlugCollisionError struct {
	Slug       string
	Path       string
	Collisions []string
}

func (e SlugCollisionError) Error() string {
	return fmt.Sprintf("slug %q in %s collides with: %s", e.Slug, e.Path, strings.Join(e.Collisions, ", "))
}

func DetectSlugCollisions(root string, scope string, options WalkOptions) ([]SlugCollisionResult, error) {
	files, err := WalkMarkdownFiles(root, options)
	if err != nil {
		return nil, err
	}

	type slugEntry struct {
		slug string
		path string
		dir  string
	}

	var entries []slugEntry
	for _, path := range files {
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			continue
		}
		parsed, parseErr := ParseYAMLFrontmatter(string(content))
		if parseErr != nil || !parsed.HasFM {
			continue
		}
		slug, ok := parsed.Frontmatter["slug"].(string)
		if !ok || slug == "" {
			continue
		}
		entries = append(entries, slugEntry{
			slug: slug,
			path: path,
			dir:  filepath.Dir(path),
		})
	}

	type bucketKey struct {
		slug  string
		scope string
	}

	buckets := map[bucketKey][]string{}
	for _, e := range entries {
		var key bucketKey
		switch scope {
		case "directory":
			key = bucketKey{slug: e.slug, scope: e.dir}
		case "project":
			key = bucketKey{slug: e.slug, scope: root}
		case "global":
			key = bucketKey{slug: e.slug, scope: ""}
		default:
			key = bucketKey{slug: e.slug, scope: e.dir}
		}
		buckets[key] = append(buckets[key], e.path)
	}

	var results []SlugCollisionResult
	for key, paths := range buckets {
		if len(paths) <= 1 {
			continue
		}
		for _, path := range paths {
			others := make([]string, 0, len(paths)-1)
			for _, p := range paths {
				if p != path {
					others = append(others, p)
				}
			}
			results = append(results, SlugCollisionResult{
				Slug:       key.slug,
				Path:       path,
				Collisions: others,
			})
		}
	}

	return results, nil
}

func ResolveSlugCollision(slug string, existingSlugs map[string]bool, policy string, maxAttempts int) (string, error) {
	if !existingSlugs[slug] {
		return slug, nil
	}

	switch policy {
	case "fail":
		return "", &SlugCollisionError{
			Slug:       slug,
			Collisions: []string{"(existing)"},
		}
	case "increment":
		for i := 2; i <= maxAttempts+1; i++ {
			candidate := fmt.Sprintf("%s-%d", slug, i)
			if !existingSlugs[candidate] {
				return candidate, nil
			}
		}
		return "", fmt.Errorf("slug %q: exhausted %d increment attempts", slug, maxAttempts)
	case "append-uid":
		return fmt.Sprintf("%s-%s", slug, generateShortUID()), nil
	default:
		return "", fmt.Errorf("unknown slug collision policy: %q", policy)
	}
}

func generateShortUID() string {
	b := make([]byte, 4)
	for i := range b {
		b[i] = "abcdefghijklmnopqrstuvwxyz0123456789"[int(b[i])%36]
	}
	return string(b)
}
