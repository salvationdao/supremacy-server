package gamedb

import (
	"github.com/jackc/pgx/v4/pgxpool"
)

var Conn *pgxpool.Pool

func New(conn *pgxpool.Pool) {
	if Conn != nil {
		panic("db already initialised")
	}
	Conn = conn
}
