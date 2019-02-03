package db

import (
	"context"
	"reflect"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

// DB stores db connection
type DB struct {
	*pg.DB
}

// New is a function that returns DB as wrapper on postgres connection.
func New(db *pg.DB) DB {
	d := DB{DB: db}

	return d
}

// Version is a function that returns Postgres version.
func (db *DB) Version() (string, error) {
	var v string
	if _, err := db.QueryOne(pg.Scan(&v), "select version()"); err != nil {
		return "", err
	}

	return v, nil
}

// buildQuery applies all functions to orm query.
func buildQuery(ctx context.Context, db orm.DB, model interface{}, search Searcher, filters []Filter, pager Pager, ops ...OpFunc) *orm.Query {
	q := db.ModelContext(ctx, model)
	for _, filter := range filters {
		filter.Apply(q)
	}

	if !reflect.ValueOf(search).IsNil() {
		search.Apply(q)
	}

	pager.Apply(q)
	applyOps(q, ops...)

	return q
}
