package account

// Type represents the type of financial account
type Type string

const (
	TypeChecking    Type = "CHECKING"
	TypeSavings     Type = "SAVINGS"
	TypeCredit      Type = "CREDIT"
	TypeInvestments Type = "INVESTMENTS"
	TypeCash        Type = "CASH"
	TypeLoan        Type = "LOAN"
	TypeUnknown     Type = "UNKNOWN"
)

// IsAsset returns true if the account type is an asset (positive balance is good)
func (t Type) IsAsset() bool {
	switch t {
	case TypeChecking, TypeSavings, TypeInvestments, TypeCash:
		return true
	default:
		return false
	}
}

// IsLiability returns true if the account type is a liability (negative balance is good)
func (t Type) IsLiability() bool {
	return !t.IsAsset()
}

// String returns the string representation of the account type
func (t Type) String() string {
	return string(t)
}
