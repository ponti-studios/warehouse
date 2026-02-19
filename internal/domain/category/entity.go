package category

import (
	"time"
)

type Domain string

const (
	DomainFinance  Domain = "finance"
	DomainHealth   Domain = "health"
	DomainTracking Domain = "tracking"
)

func (d Domain) IsValid() bool {
	switch d {
	case DomainFinance, DomainHealth, DomainTracking:
		return true
	}
	return false
}

func (d Domain) String() string {
	return string(d)
}

type Category struct {
	ID          int
	Name        string
	ParentID    *int
	Domain      Domain
	Description string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Children    []*Category
}

func (c *Category) IsRoot() bool {
	return c.ParentID == nil
}

func (c *Category) HasChildren() bool {
	return len(c.Children) > 0
}

func (c *Category) AddChild(child *Category) {
	c.Children = append(c.Children, child)
}

const UncategorizedName = "Uncategorized"

func UncategorizedCategory(domain Domain) Category {
	return Category{
		Name:        UncategorizedName,
		Domain:      domain,
		Description: "Default category for uncategorized items",
		IsActive:    true,
	}
}
