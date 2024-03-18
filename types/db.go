package types

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type Db interface {
	PrepareNamed(query string) (*sqlx.NamedStmt, error)
	Exec(query string, args ...any) (sql.Result, error)
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
}
