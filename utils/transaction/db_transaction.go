package transaction

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func StartTransaction(ctx context.Context, db *pgxpool.Pool) (pgx.Tx, error) {
	tx, err := db.Begin(ctx)

	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}

	return tx, nil
}
