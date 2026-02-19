package transaction

import "context"

// Repository defines the interface for transaction data access
type Repository interface {
	// FindByID retrieves a transaction by its ID
	FindByID(ctx context.Context, id int) (*Transaction, error)

	// FindByFilter retrieves paginated transactions matching the filter
	FindByFilter(ctx context.Context, filter Filter) (*PaginatedResult, error)

	// FindAll retrieves all transactions (use with caution - no pagination)
	FindAll(ctx context.Context) ([]Transaction, error)

	// Create inserts a new transaction
	Create(ctx context.Context, tx *Transaction) error

	// CreateBatch inserts multiple transactions in a batch
	CreateBatch(ctx context.Context, transactions []Transaction) error

	// Update updates an existing transaction
	Update(ctx context.Context, tx *Transaction) error

	// Delete removes a transaction by ID
	Delete(ctx context.Context, id int) error

	// Exists checks if a transaction exists (for deduplication)
	Exists(ctx context.Context, tx Transaction) (bool, error)

	// Count returns the total count of transactions matching the filter
	Count(ctx context.Context, filter Filter) (int, error)

	// GetCategories returns all unique categories
	GetCategories(ctx context.Context) ([]string, error)

	// GetAccounts returns all unique accounts
	GetAccounts(ctx context.Context) ([]string, error)

	// GetDateRange returns the min and max dates in the database
	GetDateRange(ctx context.Context) (min, max string, err error)
}
