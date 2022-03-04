package gamedb

import (
	"errors"

	"github.com/jackc/pgx/v4/pgxpool"
)

var Conn *pgxpool.Pool

func New(conn *pgxpool.Pool) error {
	if Conn != nil {
		return errors.New("db already initialised")
	}
	Conn = conn
	return nil
}
