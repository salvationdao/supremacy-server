package gamedb

import (
	"database/sql"
	"errors"
)

var StdConn *sql.DB

func New(stdConn *sql.DB) error {
	if StdConn != nil {
		return errors.New("db already initialised")
	}
	StdConn = stdConn
	return nil
}
