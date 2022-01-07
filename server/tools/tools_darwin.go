// +build tools
// +build !windows,!plan9

package server

//go:generate go build -o ../../bin/migrate -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate
//go:generate go build -o ../../bin/air github.com/cosmtrek/air
//go:generate go build -o ../../bin/xcaddy github.com/caddyserver/xcaddy/cmd/xcaddy

import (
	_ "github.com/caddyserver/caddy/v2"
	_ "github.com/caddyserver/xcaddy/cmd/xcaddy"
	_ "github.com/cosmtrek/air"
	_ "github.com/golang-migrate/migrate/v4"
)
