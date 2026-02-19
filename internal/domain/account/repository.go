package account

import "context"

// Repository defines the interface for account data access
type Repository interface {
	// FindByID retrieves an account by its ID
	FindByID(ctx context.Context, id int) (*Account, error)

	// FindByName retrieves an account by its canonical name
	FindByName(ctx context.Context, name string) (*Account, error)

	// FindAll retrieves all accounts
	FindAll(ctx context.Context) ([]Account, error)

	// FindActive retrieves only active accounts
	FindActive(ctx context.Context) ([]Account, error)

	// Create inserts a new account
	Create(ctx context.Context, account *Account) error

	// Update updates an existing account
	Update(ctx context.Context, account *Account) error

	// Delete removes an account by ID
	Delete(ctx context.Context, id int) error

	// GetBalances returns current balances for all accounts
	GetBalances(ctx context.Context) (map[string]float64, error)

	// UpdateBalance updates the current balance for an account
	UpdateBalance(ctx context.Context, name string, balance float64) error
}

// AliasRepository defines the interface for alias data access
type AliasRepository interface {
	// FindByAlias retrieves a canonical name by its alias
	FindByAlias(ctx context.Context, alias string) (string, error)

	// FindAll retrieves all alias mappings
	FindAll(ctx context.Context) ([]AliasMapping, error)

	// Create inserts a new alias mapping
	Create(ctx context.Context, mapping *AliasMapping) error

	// Update updates an existing alias mapping
	Update(ctx context.Context, mapping *AliasMapping) error

	// Delete removes an alias by ID
	Delete(ctx context.Context, id int) error

	// IncrementValidationCount increments the validation count for an alias
	IncrementValidationCount(ctx context.Context, alias string) error

	// GetUnvalidated returns aliases with low confidence scores
	GetUnvalidated(ctx context.Context, threshold float64) ([]AliasMapping, error)
}
