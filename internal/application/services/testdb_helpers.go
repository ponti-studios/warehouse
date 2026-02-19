package services

import (
	"database/sql"

	"gogogo/internal/domain/account"
	"gogogo/internal/domain/category"
	"gogogo/internal/domain/transaction"
	"gogogo/internal/infrastructure/persistence/sqlite"
)

func newTestTransactionRepository(db *sql.DB) transaction.Repository {
	return sqlite.NewTransactionRepository(db)
}

func newTestAccountRepository(db *sql.DB) account.Repository {
	return sqlite.NewAccountRepository(db)
}

func newTestCategoryRepository(db *sql.DB) category.Repository {
	return sqlite.NewCategoryRepository(db)
}
