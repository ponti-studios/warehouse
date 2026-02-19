package mcp

import (
	"context"
	"fmt"
	"log"
	"os"

	"gogogo/cmd/hominem/mcp/tools"
	"gogogo/internal/infrastructure/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Server struct {
	server *mcp.Server
	dbPath string
	tools  *tools.Tools
}

func NewServer() (*Server, error) {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "hominem",
			Version: "1.0.0",
		},
		nil,
	)

	t := tools.NewTools(cfg.Database.Path)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_transactions",
		Description: "Get financial transactions with optional filters for account, category, and date range",
	}, t.GetTransactions)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_accounts",
		Description: "Get all financial accounts with their current balances",
	}, t.GetAccounts)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_categories",
		Description: "Get all transaction categories",
	}, t.GetCategories)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_dashboard",
		Description: "Get a financial summary dashboard with account balances and spending overview",
	}, t.GetDashboard)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_transaction",
		Description: "Create a new financial transaction",
	}, t.CreateTransaction)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_transactions_batch",
		Description: "Create multiple financial transactions in a batch",
	}, t.CreateTransactionsBatch)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_transaction",
		Description: "Update an existing financial transaction",
	}, t.UpdateTransaction)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_transaction",
		Description: "Delete a financial transaction",
	}, t.DeleteTransaction)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_account",
		Description: "Create a new financial account",
	}, t.CreateAccount)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_account",
		Description: "Update an existing financial account",
	}, t.UpdateAccount)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_account",
		Description: "Delete a financial account and all its transactions",
	}, t.DeleteAccount)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_categories_tree",
		Description: "Get categories as a hierarchical tree structure",
	}, t.GetCategoriesTree)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_category",
		Description: "Create a new category",
	}, t.CreateCategory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_category",
		Description: "Update an existing category",
	}, t.UpdateCategory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_category",
		Description: "Delete a category and reassign its transactions to Uncategorized",
	}, t.DeleteCategory)

	return &Server{
		server: server,
		dbPath: cfg.Database.Path,
		tools:  t,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	log.Println("Starting MCP server...")
	if err := s.server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		return fmt.Errorf("MCP server error: %w", err)
	}
	return nil
}

func Run() {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "hominem",
			Version: "1.0.0",
		},
		nil,
	)

	t := tools.NewTools(cfg.Database.Path)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_transactions",
		Description: "Get financial transactions with optional filters for account, category, and date range",
	}, t.GetTransactions)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_accounts",
		Description: "Get all financial accounts with their current balances",
	}, t.GetAccounts)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_categories",
		Description: "Get all transaction categories",
	}, t.GetCategories)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_dashboard",
		Description: "Get a financial summary dashboard with account balances and spending overview",
	}, t.GetDashboard)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_transaction",
		Description: "Create a new financial transaction",
	}, t.CreateTransaction)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_transactions_batch",
		Description: "Create multiple financial transactions in a batch",
	}, t.CreateTransactionsBatch)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_transaction",
		Description: "Update an existing financial transaction",
	}, t.UpdateTransaction)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_transaction",
		Description: "Delete a financial transaction",
	}, t.DeleteTransaction)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_account",
		Description: "Create a new financial account",
	}, t.CreateAccount)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_account",
		Description: "Update an existing financial account",
	}, t.UpdateAccount)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_account",
		Description: "Delete a financial account and all its transactions",
	}, t.DeleteAccount)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_categories_tree",
		Description: "Get categories as a hierarchical tree structure",
	}, t.GetCategoriesTree)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_category",
		Description: "Create a new category",
	}, t.CreateCategory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_category",
		Description: "Update an existing category",
	}, t.UpdateCategory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_category",
		Description: "Delete a category and reassign its transactions to Uncategorized",
	}, t.DeleteCategory)

	log.Println("Starting MCP server...")

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}
