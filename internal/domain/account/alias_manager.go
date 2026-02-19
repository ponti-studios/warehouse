package account

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// AliasManager handles account name normalization and alias resolution
type AliasManager struct {
	aliasRepo AliasRepository
	cache     map[string]string
}

// NewAliasManager creates a new AliasManager
func NewAliasManager(aliasRepo AliasRepository) *AliasManager {
	return &AliasManager{
		aliasRepo: aliasRepo,
		cache:     make(map[string]string),
	}
}

// Normalize converts an account alias to its canonical name
// Returns the canonical name, a boolean indicating if this was a new/unknown alias,
// and any error that occurred
func (am *AliasManager) Normalize(ctx context.Context, name string) (string, bool, error) {
	name = strings.TrimSpace(name)

	// Handle comma-separated values (take first)
	if idx := strings.Index(name, ","); idx != -1 {
		name = strings.TrimSpace(name[:idx])
	}

	if name == "" {
		return "", false, fmt.Errorf("empty account name")
	}

	// Check cache first
	if canonical, ok := am.cache[name]; ok {
		return canonical, false, nil
	}

	// Query database
	canonical, err := am.aliasRepo.FindByAlias(ctx, name)
	if err != nil && err != sql.ErrNoRows {
		return "", false, fmt.Errorf("failed to query alias: %w", err)
	}

	if err == sql.ErrNoRows {
		// Unknown alias - handle it
		return am.handleUnknownAlias(ctx, name)
	}

	// Found in database - update cache and return
	am.cache[name] = canonical
	return canonical, false, nil
}

// handleUnknownAlias processes an unknown alias
func (am *AliasManager) handleUnknownAlias(ctx context.Context, alias string) (string, bool, error) {
	isNew := true
	canonical := alias

	// Check if it's already a canonical name by looking for exact match
	// This would require an account repository - for now, assume it's canonical

	// Add to database with low confidence for review
	mapping := &AliasMapping{
		Alias:           alias,
		CanonicalName:   canonical,
		ConfidenceScore: 0.5,
	}

	if err := am.aliasRepo.Create(ctx, mapping); err != nil {
		return "", false, fmt.Errorf("failed to create alias: %w", err)
	}

	// Update cache
	am.cache[alias] = canonical

	return canonical, isNew, nil
}

// ValidateAlias validates and creates a new alias mapping
func (am *AliasManager) ValidateAlias(ctx context.Context, alias, canonical string) error {
	if err := ValidateAlias(alias, canonical); err != nil {
		return err
	}

	mapping := &AliasMapping{
		Alias:           alias,
		CanonicalName:   canonical,
		ConfidenceScore: 1.0,
	}

	if err := am.aliasRepo.Create(ctx, mapping); err != nil {
		return fmt.Errorf("failed to create alias: %w", err)
	}

	// Update cache
	am.cache[alias] = canonical

	return nil
}

// GetUnvalidated returns aliases that need manual review
func (am *AliasManager) GetUnvalidated(ctx context.Context, threshold float64) ([]AliasMapping, error) {
	return am.aliasRepo.GetUnvalidated(ctx, threshold)
}

// WarmCache preloads the cache with all known aliases
func (am *AliasManager) WarmCache(ctx context.Context) error {
	mappings, err := am.aliasRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load aliases: %w", err)
	}

	for _, m := range mappings {
		am.cache[m.Alias] = m.CanonicalName
	}

	return nil
}

// ClearCache clears the in-memory cache
func (am *AliasManager) ClearCache() {
	am.cache = make(map[string]string)
}
