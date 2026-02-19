package account

// Account represents a financial account entity
type Account struct {
	ID             int
	Name           string
	CanonicalName  string
	Type           Type
	IsActive       bool
	CreditLimit    *float64
	CurrentBalance float64
	Currency       string
}

// DisplayBalance returns the balance flipped for display purposes.
// In the database, assets are stored as negative (outflow), so we flip the sign
// for display to show positive balances as assets.
func (a *Account) DisplayBalance() float64 {
	return -a.CurrentBalance
}

// IsAsset returns true if this account is an asset type
func (a *Account) IsAsset() bool {
	return a.Type.IsAsset()
}

// IsLiability returns true if this account is a liability type
func (a *Account) IsLiability() bool {
	return a.Type.IsLiability()
}

// AliasMapping represents a mapping from an alias to a canonical account name
type AliasMapping struct {
	ID              int
	Alias           string
	CanonicalName   string
	AccountID       *int
	ConfidenceScore float64
	ValidationCount int
	LastSeenAt      string
	CreatedAt       string
}
