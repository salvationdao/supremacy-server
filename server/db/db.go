package db

import (
	"context"
	"errors"
	"regexp"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"strings"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type SortByDir string

const (
	SortByDirAsc  SortByDir = "asc"
	SortByDirDesc SortByDir = "desc"
)

// SnakeCaseRegexp looks for snakecase words
var SnakeCaseRegexp = regexp.MustCompile(`(^|[_-])([a-z])`)

type Conn interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	SendBatch(context.Context, *pgx.Batch) pgx.BatchResults
	Begin(ctx context.Context) (pgx.Tx, error)
}

func ParseQueryText(queryText string, matchAll bool) string {
	// sanity check
	if queryText == "" {
		return ""
	}

	// trim leading and trailing spaces
	re2 := regexp.MustCompile(`\s+`)
	keywords := strings.TrimSpace(queryText)
	// to lowercase
	keywords = strings.ToLower(keywords)
	// remove excess spaces
	keywords = re2.ReplaceAllString(keywords, " ")
	// no non-alphanumeric
	re := regexp.MustCompile(`[^a-z0-9-. ]`)
	keywords = re.ReplaceAllString(keywords, "")

	// keywords array
	xkeywords := strings.Split(keywords, " ")
	// for sql construction
	var keywords2 []string
	// build sql keywords
	for _, keyword := range xkeywords {
		// skip blank, to prevent error on construct sql search
		if len(keyword) == 0 {
			continue
		}

		// add prefix for partial word search
		keyword = keyword + ":*"
		// add to search string queue
		keywords2 = append(keywords2, keyword)
	}
	// construct sql search
	if !matchAll {
		xsearch := strings.Join(keywords2, " | ")
		return xsearch
	}
	xsearch := strings.Join(keywords2, " & ")
	return xsearch
}

func Exec(ctx context.Context, conn Conn, q string, args ...interface{}) error {
	_, err := conn.Exec(ctx, q)
	return err
}

func UpsertPlayer(p *boiler.Player) error {
	boil.DebugMode = true
	err := p.Upsert(
		gamedb.StdConn,
		true,
		[]string{
			boiler.PlayerColumns.PublicAddress,
		},
		boil.Whitelist(
			boiler.PlayerColumns.ID,
			boiler.PlayerColumns.Username,
			boiler.PlayerColumns.FactionID,
			boiler.PlayerColumns.PublicAddress,
		),
		boil.Infer(),
	)
	boil.DebugMode = false
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

func UserStatGet(ctx context.Context, conn Conn, userID server.UserID) (*server.UserStat, error) {
	user := &server.UserStat{}

	q := `
		SELECT 
			us.id,
			COALESCE(us.view_battle_count,0) AS view_battle_count,
			COALESCE(us.total_ability_triggered,0) AS total_ability_triggered,
			COALESCE(us.kill_count,0) AS kill_count
		FROM user_stats us
		WHERE us.id = $1
	`

	err := pgxscan.Get(ctx, conn, user, q, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, terror.Error(err)
	}

	return user, nil
}
