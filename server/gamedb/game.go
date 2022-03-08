package gamedb

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v4/pgxpool"
)

var Conn *pgxpool.Pool
var StdConn *sql.DB

func New(conn *pgxpool.Pool, stdConn *sql.DB) error {
	if Conn != nil {
		return errors.New("db already initialised")
	}
	if StdConn != nil {
		return errors.New("db already initialised")
	}
	StdConn = stdConn
	Conn = conn
	return nil
}
