module gameserver

go 1.16

replace github.com/ninja-software/hub/v2 => /home/darren/go/src/github.com/ninja-software/hub

require (
	github.com/antonholmquist/jason v1.0.0
	github.com/caddyserver/caddy/v2 v2.4.6
	github.com/caddyserver/xcaddy v0.2.0
	github.com/cosmtrek/air v1.27.8
	github.com/georgysavva/scany v0.2.9
	github.com/getsentry/sentry-go v0.11.0
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-chi/cors v1.2.0
	github.com/gofrs/uuid v4.2.0+incompatible
	github.com/golang-migrate/migrate/v4 v4.15.1
	github.com/jackc/pgx/v4 v4.14.1
	github.com/makiuchi-d/arelo v1.9.2
	github.com/ninja-software/log_helpers v1.0.1-0.20211202070223-aff11d9a6ae6
	github.com/ninja-software/terror/v2 v2.0.7
	github.com/oklog/run v1.1.0
	github.com/ory/dockertest/v3 v3.8.1
	github.com/prometheus/client_golang v1.11.0
	github.com/rs/zerolog v1.26.0
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/net v0.0.0-20211013171255-e13a2654a71e
	nhooyr.io/websocket v1.8.7
	github.com/ninja-software/hub/v2 v2.0.4
)

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/jackc/pgconn v1.10.1
	github.com/microcosm-cc/bluemonday v1.0.16
	github.com/ninja-software/tickle v1.3.0
)
