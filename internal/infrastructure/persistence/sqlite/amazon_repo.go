package sqlite

import (
	"context"
	"database/sql"
	"fmt"
)

type AmazonRepository struct {
	db *sql.DB
}

func NewAmazonRepository(db *sql.DB) *AmazonRepository {
	return &AmazonRepository{db: db}
}

func (r *AmazonRepository) LoadExisting(ctx context.Context) (map[string]bool, error) {
	query := "SELECT order_id, asin_isbn FROM amazon_purchases"
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query existing: %w", err)
	}
	defer rows.Close()

	existing := make(map[string]bool)
	for rows.Next() {
		var orderID, asinISBN sql.NullString
		if err := rows.Scan(&orderID, &asinISBN); err != nil {
			continue
		}
		key := orderID.String + "|" + asinISBN.String
		existing[key] = true
	}

	return existing, nil
}

func (r *AmazonRepository) InsertBatch(ctx context.Context, records []interface{}) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO amazon_purchases (
			order_date, order_id, title, category, asin_isbn,
			purchase_price_per_unit, quantity, shipment_date,
			shipping_address_name, shipping_address_street_1, shipping_address_street_2,
			shipping_address_city, shipping_address_state, shipping_address_zip,
			order_status, carrier_name_and_tracking_number,
			item_subtotal, item_subtotal_tax, item_total
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, record := range records {
		row := record.([]interface{})
		_, err := stmt.ExecContext(ctx, row...)
		if err != nil {
			return fmt.Errorf("failed to insert record: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}
