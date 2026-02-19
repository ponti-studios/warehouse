package shared

import "context"

// Generic repository interface that all domain repositories can implement
type Repository[T Entity] interface {
	Create(ctx context.Context, entity T) error
	GetByID(ctx context.Context, id ID) (T, error)
	Update(ctx context.Context, entity T) error
	Delete(ctx context.Context, id ID) error
	List(ctx context.Context, filter ListFilter) ([]T, error)
	Count(ctx context.Context, filter ListFilter) (int, error)
}

// Search repository for cross-domain search functionality
type SearchRepository interface {
	Index(ctx context.Context, entity Entity, content string) error
	Search(ctx context.Context, query string, domains []string) ([]SearchResult, error)
	Reindex(ctx context.Context, domain string) error
	RemoveFromIndex(ctx context.Context, entityID ID) error
}

// Tag repository for managing tags across domains
type TagRepository interface {
	Create(ctx context.Context, tag Tag) error
	GetByID(ctx context.Context, id ID) (Tag, error)
	GetByName(ctx context.Context, name, domain string) (Tag, error)
	List(ctx context.Context, domain string) ([]Tag, error)
	Update(ctx context.Context, tag Tag) error
	Delete(ctx context.Context, id ID) error
	IncrementUsage(ctx context.Context, id ID) error
	DecrementUsage(ctx context.Context, id ID) error
	GetMostUsed(ctx context.Context, domain string, limit int) ([]Tag, error)
}

// Relationship repository for managing cross-domain relationships
type RelationshipRepository interface {
	Create(ctx context.Context, rel Relationship) error
	GetByID(ctx context.Context, id ID) (Relationship, error)
	GetForEntity(ctx context.Context, entityID ID) ([]Relationship, error)
	GetBetween(ctx context.Context, fromID, toID ID) ([]Relationship, error)
	Update(ctx context.Context, rel Relationship) error
	Delete(ctx context.Context, id ID) error
	DeleteForEntity(ctx context.Context, entityID ID) error
}

// Activity log repository for tracking changes
type ActivityLogRepository interface {
	Create(ctx context.Context, activity ActivityLog) error
	GetByID(ctx context.Context, id ID) (ActivityLog, error)
	GetForEntity(ctx context.Context, entityID ID, limit int) ([]ActivityLog, error)
	GetForDomain(ctx context.Context, domain string, limit int) ([]ActivityLog, error)
	GetRecent(ctx context.Context, limit int) ([]ActivityLog, error)
	DeleteOldEntries(ctx context.Context, olderThan Timestamp) error
}

// Entity tag mapping repository
type EntityTagRepository interface {
	AddTag(ctx context.Context, entityID, tagID ID) error
	RemoveTag(ctx context.Context, entityID, tagID ID) error
	GetTagsForEntity(ctx context.Context, entityID ID) ([]Tag, error)
	GetEntitiesForTag(ctx context.Context, tagID ID) ([]Entity, error)
	RemoveAllTags(ctx context.Context, entityID ID) error
}

// Universal entity repository for basic entity operations
type EntityRepository interface {
	Create(ctx context.Context, entity BaseEntity) error
	GetByID(ctx context.Context, id ID) (BaseEntity, error)
	Update(ctx context.Context, entity BaseEntity) error
	Delete(ctx context.Context, id ID) error
	List(ctx context.Context, filter ListFilter) ([]BaseEntity, error)
	Count(ctx context.Context, filter ListFilter) (int, error)
	GetByDomain(ctx context.Context, domain string, filter ListFilter) ([]BaseEntity, error)
	GetRecentlyUpdated(ctx context.Context, limit int) ([]BaseEntity, error)
	UpdateMetadata(ctx context.Context, id ID, metadata string) error
}
