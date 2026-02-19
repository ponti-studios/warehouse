package tools

import (
	"context"
	"fmt"

	"gogogo/internal/application/services"
	"gogogo/internal/infrastructure/persistence/sqlite"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Tools struct {
	transactionsService *services.FinanceTransactionsService
	accountsService     *services.FinanceAccountsService
	categoriesService   *services.FinanceCategoriesService
	dashboardService    *services.FinanceDashboardService
}

func NewTools(dbPath string) *Tools {
	conn, err := sqlite.NewConnection(dbPath)
	if err != nil {
		return nil
	}

	txRepo := sqlite.NewTransactionRepository(conn.DB())
	accountRepo := sqlite.NewAccountRepository(conn.DB())
	categoryRepo := sqlite.NewCategoryRepository(conn.DB())

	return &Tools{
		transactionsService: services.NewFinanceTransactionsService(txRepo),
		accountsService:     services.NewFinanceAccountsService(accountRepo, txRepo),
		categoriesService:   services.NewFinanceCategoriesService(categoryRepo),
		dashboardService:    services.NewFinanceDashboardService(accountRepo),
	}
}

type GetTransactionsInput struct {
	Account   string `json:"account,omitempty" jsonschema:"Filter by account name"`
	Category  string `json:"category,omitempty" jsonschema:"Filter by category name"`
	StartDate string `json:"startDate,omitempty" jsonschema:"Start date in YYYY-MM-DD format"`
	EndDate   string `json:"endDate,omitempty" jsonschema:"End date in YYYY-MM-DD format"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum number of transactions to return, default 20"`
	Offset    int    `json:"offset,omitempty" jsonschema:"Number of transactions to skip for pagination"`
}

type TransactionOutput struct {
	ID       string  `json:"id" jsonschema:"Unique transaction identifier"`
	Date     string  `json:"date" jsonschema:"Transaction date in YYYY-MM-DD format"`
	Amount   float64 `json:"amount" jsonschema:"Transaction amount (positive for credits, negative for debits)"`
	Account  string  `json:"account" jsonschema:"Account name where transaction occurred"`
	Category string  `json:"category" jsonschema:"Transaction category"`
	Payee    string  `json:"payee" jsonschema:"Payee or merchant name"`
	Notes    string  `json:"notes" jsonschema:"User notes about the transaction"`
}

type GetTransactionsOutput struct {
	Transactions []TransactionOutput `json:"transactions" jsonschema:"List of transactions"`
	TotalCount   int                 `json:"totalCount" jsonschema:"Total number of transactions matching the filter"`
}

func (t *Tools) GetTransactions(ctx context.Context, req *mcp.CallToolRequest, input GetTransactionsInput) (*mcp.CallToolResult, GetTransactionsOutput, error) {
	txs, err := t.transactionsService.GetTransactions(context.Background(), services.GetTransactionsOptions{
		Account:   input.Account,
		Category:  input.Category,
		StartDate: input.StartDate,
		EndDate:   input.EndDate,
		Page:      input.Offset/input.Limit + 1,
		PerPage:   input.Limit,
	})
	if err != nil {
		return nil, GetTransactionsOutput{}, err
	}

	output := make([]TransactionOutput, len(txs))
	for i, tx := range txs {
		output[i] = TransactionOutput{
			ID:       tx.ID,
			Date:     tx.Date,
			Amount:   tx.Amount,
			Account:  tx.Account,
			Category: tx.Category,
			Payee:    tx.Payee,
			Notes:    tx.Notes,
		}
	}

	return nil, GetTransactionsOutput{
		Transactions: output,
		TotalCount:   len(output),
	}, nil
}

type AccountOutput struct {
	ID          string  `json:"id" jsonschema:"Unique account identifier"`
	Name        string  `json:"name" jsonschema:"Account name"`
	Type        string  `json:"type" jsonschema:"Account type (checking, savings, credit, etc.)"`
	Balance     float64 `json:"balance" jsonschema:"Current account balance"`
	Currency    string  `json:"currency" jsonschema:"Currency code (e.g., USD)"`
	IsActive    bool    `json:"isActive" jsonschema:"Whether the account is active"`
	LastUpdated string  `json:"lastUpdated" jsonschema:"Last update date in YYYY-MM-DD format"`
}

type GetAccountsOutput struct {
	Accounts []AccountOutput `json:"accounts" jsonschema:"List of accounts"`
	Count    int             `json:"count" jsonschema:"Total number of accounts"`
}

func (t *Tools) GetAccounts(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, GetAccountsOutput, error) {
	accounts, err := t.accountsService.GetAccounts(context.Background())
	if err != nil {
		return nil, GetAccountsOutput{}, err
	}

	output := make([]AccountOutput, len(accounts))
	for i, acc := range accounts {
		output[i] = AccountOutput{
			ID:          acc.ID,
			Name:        acc.Name,
			Type:        acc.Type,
			Balance:     acc.Balance,
			Currency:    acc.Currency,
			IsActive:    acc.IsActive,
			LastUpdated: acc.LastUpdated,
		}
	}

	return nil, GetAccountsOutput{
		Accounts: output,
		Count:    len(output),
	}, nil
}

type CategoryOutput struct {
	ID               string `json:"id" jsonschema:"Unique category identifier"`
	Name             string `json:"name" jsonschema:"Category name"`
	ParentID         string `json:"parentId" jsonschema:"Parent category ID for subcategories"`
	TransactionCount int    `json:"transactionCount" jsonschema:"Number of transactions in this category"`
}

type GetCategoriesOutput struct {
	Categories []CategoryOutput `json:"categories" jsonschema:"List of categories"`
	Count      int              `json:"count" jsonschema:"Total number of categories"`
}

func (t *Tools) GetCategories(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, GetCategoriesOutput, error) {
	categories, err := t.categoriesService.GetCategories(context.Background(), "")
	if err != nil {
		return nil, GetCategoriesOutput{}, err
	}

	output := make([]CategoryOutput, len(categories))
	for i, cat := range categories {
		parentID := ""
		if cat.ParentID != nil {
			parentID = *cat.ParentID
		}
		output[i] = CategoryOutput{
			ID:               cat.ID,
			Name:             cat.Name,
			ParentID:         parentID,
			TransactionCount: 0,
		}
	}

	return nil, GetCategoriesOutput{
		Categories: output,
		Count:      len(output),
	}, nil
}

type DashboardOutput struct {
	TotalBalance       float64             `json:"totalBalance" jsonschema:"Sum of all account balances"`
	TotalIncome        float64             `json:"totalIncome" jsonschema:"Total income in the current period"`
	TotalExpenses      float64             `json:"totalExpenses" jsonschema:"Total expenses in the current period"`
	NetWorth           float64             `json:"netWorth" jsonschema:"Total assets minus liabilities"`
	Accounts           []AccountOutput     `json:"accounts" jsonschema:"List of accounts with balances"`
	TopCategories      []CategorySpending  `json:"topCategories" jsonschema:"Top spending categories"`
	RecentTransactions []TransactionOutput `json:"recentTransactions" jsonschema:"Most recent transactions"`
}

type CategorySpending struct {
	Category string  `json:"category" jsonschema:"Category name"`
	Amount   float64 `json:"amount" jsonschema:"Total spending in this category"`
}

func (t *Tools) GetDashboard(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, DashboardOutput, error) {
	dashboard, err := t.dashboardService.GetDashboard(context.Background())
	if err != nil {
		return nil, DashboardOutput{}, err
	}

	accounts := make([]AccountOutput, len(dashboard.Accounts))
	for i, acc := range dashboard.Accounts {
		accounts[i] = AccountOutput{
			ID:          acc.ID,
			Name:        acc.Name,
			Type:        acc.Type,
			Balance:     acc.Balance,
			Currency:    acc.Currency,
			IsActive:    acc.IsActive,
			LastUpdated: acc.LastUpdated,
		}
	}

	return nil, DashboardOutput{
		TotalBalance: dashboard.TotalBalance,
		NetWorth:     dashboard.NetWorth,
		Accounts:     accounts,
	}, nil
}

type CreateTransactionInput struct {
	Date     string  `json:"date" jsonschema:"Transaction date in YYYY-MM-DD format (required)"`
	Name     string  `json:"name" jsonschema:"Transaction description or payee name (required)"`
	Amount   float64 `json:"amount" jsonschema:"Transaction amount (required, positive for credits, negative for debits)"`
	Account  string  `json:"account" jsonschema:"Account name where transaction occurred (required)"`
	Category string  `json:"category" jsonschema:"Category name for the transaction"`
	Note     string  `json:"note" jsonschema:"User notes about the transaction"`
}

func (t *Tools) CreateTransaction(ctx context.Context, req *mcp.CallToolRequest, input CreateTransactionInput) (*mcp.CallToolResult, TransactionOutput, error) {
	tx, err := t.transactionsService.CreateTransaction(context.Background(), services.CreateTransactionInput{
		Date:     input.Date,
		Name:     input.Name,
		Amount:   input.Amount,
		Account:  input.Account,
		Category: input.Category,
		Note:     input.Note,
	})
	if err != nil {
		return nil, TransactionOutput{}, err
	}

	return nil, TransactionOutput{
		ID:       tx.ID,
		Date:     tx.Date,
		Amount:   tx.Amount,
		Account:  tx.Account,
		Category: tx.Category,
		Payee:    tx.Payee,
		Notes:    tx.Notes,
	}, nil
}

type CreateTransactionsBatchInput struct {
	Transactions   []CreateTransactionInput `json:"transactions" jsonschema:"Array of transactions to create"`
	SkipDuplicates bool                     `json:"skipDuplicates" jsonschema:"Whether to skip duplicate transactions"`
}

type BatchCreateResultOutput struct {
	Created int      `json:"created" jsonschema:"Number of transactions created"`
	Skipped int      `json:"skipped" jsonschema:"Number of transactions skipped"`
	Errors  []string `json:"errors,omitempty" jsonschema:"List of errors encountered"`
}

func (t *Tools) CreateTransactionsBatch(ctx context.Context, req *mcp.CallToolRequest, input CreateTransactionsBatchInput) (*mcp.CallToolResult, BatchCreateResultOutput, error) {
	txs := make([]services.CreateTransactionInput, len(input.Transactions))
	for i, tx := range input.Transactions {
		txs[i] = services.CreateTransactionInput{
			Date:     tx.Date,
			Name:     tx.Name,
			Amount:   tx.Amount,
			Account:  tx.Account,
			Category: tx.Category,
			Note:     tx.Note,
		}
	}

	result, err := t.transactionsService.CreateTransactionsBatch(context.Background(), services.CreateTransactionsBatchInput{
		Transactions:   txs,
		SkipDuplicates: input.SkipDuplicates,
	})
	if err != nil {
		return nil, BatchCreateResultOutput{}, err
	}

	return nil, BatchCreateResultOutput{
		Created: result.Created,
		Skipped: result.Skipped,
		Errors:  result.Errors,
	}, nil
}

type UpdateTransactionInput struct {
	ID       string  `json:"id" jsonschema:"Transaction ID to update (required)"`
	Date     string  `json:"date" jsonschema:"New transaction date in YYYY-MM-DD format"`
	Name     string  `json:"name" jsonschema:"New transaction description or payee name"`
	Amount   float64 `json:"amount" jsonschema:"New transaction amount"`
	Account  string  `json:"account" jsonschema:"New account name"`
	Category string  `json:"category" jsonschema:"New category name"`
	Note     string  `json:"note" jsonschema:"New user notes"`
}

func (t *Tools) UpdateTransaction(ctx context.Context, req *mcp.CallToolRequest, input UpdateTransactionInput) (*mcp.CallToolResult, TransactionOutput, error) {
	tx, err := t.transactionsService.UpdateTransaction(context.Background(), services.UpdateTransactionInput{
		ID:       input.ID,
		Date:     input.Date,
		Name:     input.Name,
		Amount:   input.Amount,
		Account:  input.Account,
		Category: input.Category,
		Note:     input.Note,
	})
	if err != nil {
		return nil, TransactionOutput{}, err
	}

	return nil, TransactionOutput{
		ID:       tx.ID,
		Date:     tx.Date,
		Amount:   tx.Amount,
		Account:  tx.Account,
		Category: tx.Category,
		Payee:    tx.Payee,
		Notes:    tx.Notes,
	}, nil
}

type DeleteTransactionInput struct {
	ID string `json:"id" jsonschema:"Transaction ID to delete (required)"`
}

type DeleteResultOutput struct {
	Success bool   `json:"success" jsonschema:"Whether the deletion was successful"`
	Message string `json:"message" jsonschema:"Result message"`
}

func (t *Tools) DeleteTransaction(ctx context.Context, req *mcp.CallToolRequest, input DeleteTransactionInput) (*mcp.CallToolResult, DeleteResultOutput, error) {
	err := t.transactionsService.DeleteTransaction(context.Background(), input.ID)
	if err != nil {
		return nil, DeleteResultOutput{}, err
	}

	return nil, DeleteResultOutput{
		Success: true,
		Message: "Transaction deleted successfully",
	}, nil
}

type CreateAccountInput struct {
	Name     string `json:"name" jsonschema:"Account name (required)"`
	Type     string `json:"type" jsonschema:"Account type: CHECKING, SAVINGS, CREDIT, INVESTMENTS, CASH, or LOAN (required)"`
	Currency string `json:"currency" jsonschema:"ISO 4217 currency code (required, e.g., USD, EUR)"`
}

func (t *Tools) CreateAccount(ctx context.Context, req *mcp.CallToolRequest, input CreateAccountInput) (*mcp.CallToolResult, AccountOutput, error) {
	acc, err := t.accountsService.CreateAccount(context.Background(), services.CreateAccountInput{
		Name:     input.Name,
		Type:     input.Type,
		Currency: input.Currency,
	})
	if err != nil {
		return nil, AccountOutput{}, err
	}

	return nil, AccountOutput{
		ID:          acc.ID,
		Name:        acc.Name,
		Type:        acc.Type,
		Balance:     acc.Balance,
		Currency:    acc.Currency,
		IsActive:    acc.IsActive,
		LastUpdated: acc.LastUpdated,
	}, nil
}

type UpdateAccountInput struct {
	ID       string `json:"id" jsonschema:"Account ID to update (required)"`
	Name     string `json:"name" jsonschema:"New account name"`
	Type     string `json:"type" jsonschema:"New account type"`
	IsActive *bool  `json:"isActive" jsonschema:"Whether the account is active"`
}

func (t *Tools) UpdateAccount(ctx context.Context, req *mcp.CallToolRequest, input UpdateAccountInput) (*mcp.CallToolResult, AccountOutput, error) {
	acc, err := t.accountsService.UpdateAccount(context.Background(), services.UpdateAccountInput{
		ID:       input.ID,
		Name:     input.Name,
		Type:     input.Type,
		IsActive: input.IsActive,
	})
	if err != nil {
		return nil, AccountOutput{}, err
	}

	return nil, AccountOutput{
		ID:          acc.ID,
		Name:        acc.Name,
		Type:        acc.Type,
		Balance:     acc.Balance,
		Currency:    acc.Currency,
		IsActive:    acc.IsActive,
		LastUpdated: acc.LastUpdated,
	}, nil
}

type DeleteAccountInput struct {
	ID string `json:"id" jsonschema:"Account ID to delete (required)"`
}

func (t *Tools) DeleteAccount(ctx context.Context, req *mcp.CallToolRequest, input DeleteAccountInput) (*mcp.CallToolResult, DeleteResultOutput, error) {
	count, err := t.accountsService.DeleteAccount(context.Background(), input.ID)
	if err != nil {
		return nil, DeleteResultOutput{}, err
	}

	return nil, DeleteResultOutput{
		Success: true,
		Message: fmt.Sprintf("Deleted account and %d associated transaction(s)", count),
	}, nil
}

type GetCategoriesTreeInput struct {
	Domain string `json:"domain,omitempty" jsonschema:"Filter by domain: finance, health, or tracking"`
}

type CategoryTreeOutput struct {
	ID          string                   `json:"id" jsonschema:"Unique category identifier"`
	Name        string                   `json:"name" jsonschema:"Category name"`
	ParentID    string                   `json:"parentId" jsonschema:"Parent category ID for subcategories"`
	Domain      string                   `json:"domain" jsonschema:"Category domain"`
	Description string                   `json:"description" jsonschema:"Category description"`
	Children    []map[string]interface{} `json:"children" jsonschema:"Array of child category objects"`
}

type GetCategoriesTreeOutput struct {
	Categories []CategoryTreeOutput `json:"categories" jsonschema:"Hierarchical tree of categories with nested children"`
}

func (t *Tools) GetCategoriesTree(ctx context.Context, req *mcp.CallToolRequest, input GetCategoriesTreeInput) (*mcp.CallToolResult, GetCategoriesTreeOutput, error) {
	categories, err := t.categoriesService.GetCategoryTree(context.Background(), input.Domain)
	if err != nil {
		return nil, GetCategoriesTreeOutput{}, err
	}

	output := make([]CategoryTreeOutput, len(categories))
	for i, cat := range categories {
		output[i] = categoryToTreeDTO(cat)
	}

	return nil, GetCategoriesTreeOutput{Categories: output}, nil
}

func categoryToTreeDTO(cat services.FinanceCategoryDTO) CategoryTreeOutput {
	var parentID string
	if cat.ParentID != nil {
		parentID = *cat.ParentID
	}

	children := make([]map[string]interface{}, 0)
	for _, child := range cat.Children {
		childDTO := categoryToTreeDTO(child)
		children = append(children, map[string]interface{}{
			"id":          childDTO.ID,
			"name":        childDTO.Name,
			"parentId":    childDTO.ParentID,
			"domain":      childDTO.Domain,
			"description": childDTO.Description,
			"children":    childDTO.Children,
		})
	}

	return CategoryTreeOutput{
		ID:          cat.ID,
		Name:        cat.Name,
		ParentID:    parentID,
		Domain:      cat.Domain,
		Description: cat.Description,
		Children:    children,
	}
}

type CreateCategoryInput struct {
	Name        string  `json:"name" jsonschema:"Category name (required)"`
	ParentID    *string `json:"parentId" jsonschema:"Parent category ID for subcategories"`
	Domain      string  `json:"domain" jsonschema:"Domain: finance, health, or tracking (required)"`
	Description string  `json:"description" jsonschema:"Category description"`
}

func (t *Tools) CreateCategory(ctx context.Context, req *mcp.CallToolRequest, input CreateCategoryInput) (*mcp.CallToolResult, CategoryOutput, error) {
	cat, err := t.categoriesService.CreateCategory(context.Background(), services.CreateCategoryInput{
		Name:        input.Name,
		ParentID:    input.ParentID,
		Domain:      input.Domain,
		Description: input.Description,
	})
	if err != nil {
		return nil, CategoryOutput{}, err
	}

	parentID := ""
	if cat.ParentID != nil {
		parentID = *cat.ParentID
	}

	return nil, CategoryOutput{
		ID:               cat.ID,
		Name:             cat.Name,
		ParentID:         parentID,
		TransactionCount: 0,
	}, nil
}

type UpdateCategoryInput struct {
	ID          string  `json:"id" jsonschema:"Category ID to update (required)"`
	Name        string  `json:"name" jsonschema:"New category name"`
	ParentID    *string `json:"parentId" jsonschema:"New parent category ID"`
	Description string  `json:"description" jsonschema:"New category description"`
}

func (t *Tools) UpdateCategory(ctx context.Context, req *mcp.CallToolRequest, input UpdateCategoryInput) (*mcp.CallToolResult, CategoryOutput, error) {
	cat, err := t.categoriesService.UpdateCategory(context.Background(), services.UpdateCategoryInput{
		ID:          input.ID,
		Name:        input.Name,
		ParentID:    input.ParentID,
		Description: input.Description,
	})
	if err != nil {
		return nil, CategoryOutput{}, err
	}

	parentID := ""
	if cat.ParentID != nil {
		parentID = *cat.ParentID
	}

	return nil, CategoryOutput{
		ID:               cat.ID,
		Name:             cat.Name,
		ParentID:         parentID,
		TransactionCount: 0,
	}, nil
}

type DeleteCategoryInput struct {
	ID string `json:"id" jsonschema:"Category ID to delete (required)"`
}

func (t *Tools) DeleteCategory(ctx context.Context, req *mcp.CallToolRequest, input DeleteCategoryInput) (*mcp.CallToolResult, DeleteResultOutput, error) {
	count, err := t.categoriesService.DeleteCategory(context.Background(), input.ID)
	if err != nil {
		return nil, DeleteResultOutput{}, err
	}

	return nil, DeleteResultOutput{
		Success: true,
		Message: fmt.Sprintf("Deleted category and reassigned %d transaction(s) to Uncategorized", count),
	}, nil
}
