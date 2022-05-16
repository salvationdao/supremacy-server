package api

import (
	"fmt"
	"server/db"

	"github.com/ninja-software/terror/v2"
)

var (
	ErrCheckDBQuery = fmt.Errorf("error: executing db query")
	ErrCheckDBDirty = fmt.Errorf("db is dirty")
)

// check checks server is working correctly
func check() error {

	// check db dirty
	count := 0
	err := db.IsSchemaDirty(&count)
	if err != nil {
		return terror.Error(ErrCheckDBQuery)
	}
	if count > 0 {
		return terror.Error(ErrCheckDBDirty)
	}

	return nil
}
