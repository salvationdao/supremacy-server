//go:build !windows && !plan9
// +build !windows,!plan9

// package server holds platform wide source code

//go:generate ../bin/sqlboiler ../bin/sqlboiler-psql --config ./sqlboiler.toml

package server
