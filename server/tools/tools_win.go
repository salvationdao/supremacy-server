// +build tools
// +build windows

package server

//go:generate go build -o ../../bin/migrate.exe -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate
//go:generate go build -o ../../bin/arelo.exe github.com/makiuchi-d/arelo

import (
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/makiuchi-d/arelo"
)
