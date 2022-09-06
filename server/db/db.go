package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"regexp"
	"strings"
)

// SnakeCaseRegexp looks for snakecase words
var SnakeCaseRegexp = regexp.MustCompile(`(^|[_-])([a-z])`)

type Conn interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
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
