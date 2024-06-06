package db

import (
	"database/sql"

	"github.com/josuebrunel/sportdropin/pkg/xlog"
	_ "github.com/lib/pq"
)

func New(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		xlog.Error("error while opening database", "err", err)
		return nil, err
	}
	return db, nil
}
