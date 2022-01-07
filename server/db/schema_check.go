package db

import (
	"context"

	"github.com/georgysavva/scany/pgxscan"
)

func IsSchemaDirty(ctx context.Context, conn Conn, count *int) error {
	q := `SELECT count(*) FROM schema_migrations where dirty is true`
	return pgxscan.Get(ctx, conn, count, q)
}
