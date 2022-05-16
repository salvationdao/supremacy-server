package db

import (
	"server/gamedb"
)

func IsSchemaDirty(count *int) error {
	q := `SELECT count(*) FROM schema_migrations where dirty is true`
	return gamedb.StdConn.QueryRow(q).Scan(count)
}
