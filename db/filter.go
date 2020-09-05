package db

import (
	"fmt"
	"strings"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/go-pg/pg/v9/types"
)

const (
	SearchTypeEquals = iota
	SearchTypeNull
	SearchTypeGE
	SearchTypeLE
	SearchTypeGreater
	SearchTypeLess
	SearchTypeLike
	SearchTypeILike
	SearchTypeArray
)

var formatter = orm.Formatter{}

var searchTypes = map[bool]map[int]string{
	// include
	false: {
		SearchTypeEquals:  "= ?",
		SearchTypeNull:    "is null",
		SearchTypeGE:      ">= ?",
		SearchTypeLE:      "<= ?",
		SearchTypeGreater: "> ?",
		SearchTypeLess:    "< ?",
		SearchTypeLike:    "like ?",
		SearchTypeILike:   "ilike ?",
		SearchTypeArray:   "in (?)",
	},
	// exclude
	true: {
		SearchTypeEquals:  "!= ?",
		SearchTypeNull:    "is not null",
		SearchTypeGE:      "< ?",
		SearchTypeLE:      "> ?",
		SearchTypeGreater: "<= ?",
		SearchTypeLess:    ">= ?",
		SearchTypeLike:    "not (like ?)",
		SearchTypeILike:   "not (ilike ?)",
		SearchTypeArray:   "not in (?)",
	},
}

const TablePrefix = "t"
const TableColumns = "t.*"

type Filter struct {
	Field      string      `json:"field"`             //search field
	Value      interface{} `json:"value,omitempty"`   //search value
	SearchType int         `json:"type,omitempty"`    //search type. see db/filter.go
	Exclude    bool        `json:"exclude,omitempty"` //is this filter should exclude
}

// String prints filter as sql string
func (f Filter) String() string {
	fld, val := f.prepare()
	return string(formatter.FormatQuery([]byte{}, "? ?", fld, val))
}

// Apply applies filter to go-pg orm
func (f Filter) Apply(query *orm.Query) *orm.Query {
	fld, val := f.prepare()
	return query.Where("? ?", fld, val)
}

func (f Filter) prepare() (field, value types.ValueAppender) {
	// preparing value
	switch f.SearchType {
	case SearchTypeArray:
		f.Value = pg.In(f.Value)
	case SearchTypeILike, SearchTypeLike:
		f.Value = `%` + f.Value.(string) + `%`
	}

	if !strings.Contains(f.Field, ".") {
		f.Field = fmt.Sprintf("%s.%s", TablePrefix, f.Field)
	}

	// preparing search type
	st, ok := searchTypes[f.Exclude][f.SearchType]
	if !ok {
		st = searchTypes[f.Exclude][SearchTypeEquals]
	}

	return pg.Ident(f.Field), pg.SafeQuery(st, f.Value)
}
