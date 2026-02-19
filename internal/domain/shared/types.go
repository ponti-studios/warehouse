package shared

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

// Common identifier type
type ID string

// Generate a new ID (using timestamp + random component for simplicity)
func NewID(prefix string) ID {
	timestamp := time.Now().Unix()
	return ID(fmt.Sprintf("%s_%d_%d", prefix, timestamp, rand.Intn(10000)))
}

func (id ID) String() string {
	return string(id)
}

// Common timestamp handling
type Timestamp time.Time

func (t Timestamp) Time() time.Time {
	return time.Time(t)
}

func (t Timestamp) String() string {
	return time.Time(t).Format(time.RFC3339)
}

func (t *Timestamp) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	*t = Timestamp(parsed)
	return nil
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func Now() Timestamp {
	return Timestamp(time.Now().UTC())
}

// Common entity interface
type Entity interface {
	GetID() ID
	GetDomain() string
	GetType() string
	GetTitle() string
	GetCreatedAt() Timestamp
	GetUpdatedAt() Timestamp
	SetUpdatedAt(t Timestamp)
}

// Base entity struct that all domain entities can embed
type BaseEntity struct {
	ID        ID        `json:"id" db:"id"`
	Domain    string    `json:"domain" db:"domain"`
	Type      string    `json:"type" db:"entity_type"`
	Title     string    `json:"title" db:"title"`
	Status    string    `json:"status" db:"status"`
	CreatedAt Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt Timestamp `json:"updated_at" db:"updated_at"`
}

func NewBaseEntity(domain, entityType, title string) BaseEntity {
	now := Now()
	return BaseEntity{
		ID:        NewID(domain),
		Domain:    domain,
		Type:      entityType,
		Title:     title,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (e BaseEntity) GetID() ID                 { return e.ID }
func (e BaseEntity) GetDomain() string         { return e.Domain }
func (e BaseEntity) GetType() string           { return e.Type }
func (e BaseEntity) GetTitle() string          { return e.Title }
func (e BaseEntity) GetCreatedAt() Timestamp   { return e.CreatedAt }
func (e BaseEntity) GetUpdatedAt() Timestamp   { return e.UpdatedAt }
func (e *BaseEntity) SetUpdatedAt(t Timestamp) { e.UpdatedAt = t }

// Search result
type SearchResult struct {
	EntityID ID      `json:"entity_id"`
	Domain   string  `json:"domain"`
	Type     string  `json:"type"`
	Title    string  `json:"title"`
	Snippet  string  `json:"snippet"`
	Score    float64 `json:"score"`
}

// Tag represents a label that can be applied to entities
type Tag struct {
	BaseEntity
	Name        string `json:"name" db:"name"`
	Color       string `json:"color" db:"color"`
	Description string `json:"description" db:"description"`
	UsageCount  int    `json:"usage_count" db:"usage_count"`
}

func NewTag(name, domain, color, description string) Tag {
	base := NewBaseEntity("tags", "tag", name)
	if domain != "" {
		base.Domain = domain
	}
	return Tag{
		BaseEntity:  base,
		Name:        name,
		Color:       color,
		Description: description,
		UsageCount:  0,
	}
}

// Relationship between entities across domains
type Relationship struct {
	ID               ID        `json:"id" db:"id"`
	FromEntityID     ID        `json:"from_entity_id" db:"from_entity_id"`
	ToEntityID       ID        `json:"to_entity_id" db:"to_entity_id"`
	RelationshipType string    `json:"relationship_type" db:"relationship_type"`
	Strength         int       `json:"strength" db:"strength"`
	Bidirectional    bool      `json:"bidirectional" db:"bidirectional"`
	Metadata         string    `json:"metadata" db:"metadata"`
	CreatedAt        Timestamp `json:"created_at" db:"created_at"`
}

// Activity log entry for tracking changes
type ActivityLog struct {
	ID          ID        `json:"id" db:"id"`
	EntityID    *ID       `json:"entity_id" db:"entity_id"`
	Action      string    `json:"action" db:"action"`
	Domain      string    `json:"domain" db:"domain"`
	Description string    `json:"description" db:"description"`
	Metadata    string    `json:"metadata" db:"metadata"`
	CreatedAt   Timestamp `json:"created_at" db:"created_at"`
}

// Common list filter for querying entities
type ListFilter struct {
	Domain    string
	Type      string
	Status    string
	Tags      []string
	Search    string
	StartDate *Timestamp
	EndDate   *Timestamp
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// DefaultListFilter returns a filter with sensible defaults
func DefaultListFilter() ListFilter {
	return ListFilter{
		Limit:     20,
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: "DESC",
		Status:    "active",
	}
}

// Priority levels for items that need prioritization
type Priority int

const (
	PriorityLow Priority = iota + 1
	PriorityMedium
	PriorityHigh
	PriorityUrgent
)

func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityMedium:
		return "medium"
	case PriorityHigh:
		return "high"
	case PriorityUrgent:
		return "urgent"
	default:
		return "unknown"
	}
}

// Common status types
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusArchived = "archived"
	StatusDeleted  = "deleted"
)

// Common relationship types
const (
	RelationshipTypeRelated    = "related"
	RelationshipTypeContains   = "contains"
	RelationshipTypeDependsOn  = "depends_on"
	RelationshipTypeBlocks     = "blocks"
	RelationshipTypeFollowsUp  = "follows_up"
	RelationshipTypeReferences = "references"
)

// Error types
type DomainError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Domain  string `json:"domain"`
}

func (e DomainError) Error() string {
	return fmt.Sprintf("[%s:%s] %s", e.Domain, e.Code, e.Message)
}

func NewDomainError(domain, code, message string) DomainError {
	return DomainError{
		Domain:  domain,
		Code:    code,
		Message: message,
	}
}

// Common validation errors
var (
	ErrEntityNotFound = func(domain, id string) DomainError {
		return NewDomainError(domain, "ENTITY_NOT_FOUND", fmt.Sprintf("Entity with ID %s not found", id))
	}
	ErrInvalidInput = func(domain, field string) DomainError {
		return NewDomainError(domain, "INVALID_INPUT", fmt.Sprintf("Invalid input for field: %s", field))
	}
	ErrDuplicateEntity = func(domain, identifier string) DomainError {
		return NewDomainError(domain, "DUPLICATE_ENTITY", fmt.Sprintf("Entity already exists: %s", identifier))
	}
)
