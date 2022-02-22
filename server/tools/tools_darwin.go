// +build tools
// +build !windows,!plan9

package server

//go:generate go build -o ../../bin/migrate -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate
//go:generate go build -o ../../bin/air github.com/cosmtrek/air

import (
	_ "github.com/cosmtrek/air"
	_ "github.com/golang-migrate/migrate/v4"
)
