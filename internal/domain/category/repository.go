package category

import "context"

type Repository interface {
	FindByID(ctx context.Context, id int) (*Category, error)
	FindByName(ctx context.Context, name string, domain Domain) (*Category, error)
	FindByDomain(ctx context.Context, domain Domain) ([]Category, error)
	FindChildren(ctx context.Context, parentID int) ([]Category, error)
	FindAll(ctx context.Context) ([]Category, error)
	Create(ctx context.Context, cat *Category) error
	Update(ctx context.Context, cat *Category) error
	Delete(ctx context.Context, id int) error
	GetTree(ctx context.Context, domain Domain) ([]Category, error)
	HasChildren(ctx context.Context, id int) (bool, error)
	CountByCategory(ctx context.Context, categoryName string, domain Domain) (int, error)
	ReassignTransactions(ctx context.Context, fromName, toName string, domain Domain) (int, error)
}
