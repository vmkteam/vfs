//nolint
//lint:file-ignore U1000 ignore unused code, it's generated
package db

import (
	"time"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

const condition = "?.? = ?"

// base filters
type applier func(query *orm.Query) (*orm.Query, error)

type search struct {
	appliers []applier
}

func (s *search) apply(query *orm.Query) {
	for _, applier := range s.appliers {
		query.Apply(applier)
	}
}

func (s *search) where(query *orm.Query, table, field string, value interface{}) {
	query.Where(condition, pg.F(table), pg.F(field), value)
}

func (s *search) WithApply(a applier) {
	if s.appliers == nil {
		s.appliers = []applier{}
	}
	s.appliers = append(s.appliers, a)
}

func (s *search) With(condition string, params ...interface{}) {
	s.WithApply(func(query *orm.Query) (*orm.Query, error) {
		return query.Where(condition, params...), nil
	})
}

// Searcher is interface for every generated filter
type Searcher interface {
	Apply(query *orm.Query) *orm.Query
	Q() applier

	With(condition string, params ...interface{})
	WithApply(a applier)
}

type VfsFileSearch struct {
	search

	ID         *int
	FolderID   *int
	Title      *string
	Path       *string
	Params     *string
	IsFavorite *bool
	MimeType   *string
	FileSize   *int
	FileExists *bool
	StatusID   *int
	CreatedAt  *time.Time
}

func (s *VfsFileSearch) Apply(query *orm.Query) *orm.Query {
	if s.ID != nil {
		s.where(query, Tables.VfsFile.Alias, Columns.VfsFile.ID, s.ID)
	}
	if s.FolderID != nil {
		s.where(query, Tables.VfsFile.Alias, Columns.VfsFile.FolderID, s.FolderID)
	}
	if s.Title != nil {
		s.where(query, Tables.VfsFile.Alias, Columns.VfsFile.Title, s.Title)
	}
	if s.Path != nil {
		s.where(query, Tables.VfsFile.Alias, Columns.VfsFile.Path, s.Path)
	}
	if s.Params != nil {
		s.where(query, Tables.VfsFile.Alias, Columns.VfsFile.Params, s.Params)
	}
	if s.IsFavorite != nil {
		s.where(query, Tables.VfsFile.Alias, Columns.VfsFile.IsFavorite, s.IsFavorite)
	}
	if s.MimeType != nil {
		s.where(query, Tables.VfsFile.Alias, Columns.VfsFile.MimeType, s.MimeType)
	}
	if s.FileSize != nil {
		s.where(query, Tables.VfsFile.Alias, Columns.VfsFile.FileSize, s.FileSize)
	}
	if s.FileExists != nil {
		s.where(query, Tables.VfsFile.Alias, Columns.VfsFile.FileExists, s.FileExists)
	}
	if s.StatusID != nil {
		s.where(query, Tables.VfsFile.Alias, Columns.VfsFile.StatusID, s.StatusID)
	}
	if s.CreatedAt != nil {
		s.where(query, Tables.VfsFile.Alias, Columns.VfsFile.CreatedAt, s.CreatedAt)
	}

	s.apply(query)

	return query
}

func (s *VfsFileSearch) Q() applier {
	return func(query *orm.Query) (*orm.Query, error) {
		return s.Apply(query), nil
	}
}

type VfsFolderSearch struct {
	search

	ID             *int
	ParentFolderID *int
	Title          *string
	IsFavorite     *bool
	CreatedAt      *time.Time
	StatusID       *int
}

func (s *VfsFolderSearch) Apply(query *orm.Query) *orm.Query {
	if s.ID != nil {
		s.where(query, Tables.VfsFolder.Alias, Columns.VfsFolder.ID, s.ID)
	}
	if s.ParentFolderID != nil {
		s.where(query, Tables.VfsFolder.Alias, Columns.VfsFolder.ParentFolderID, s.ParentFolderID)
	}
	if s.Title != nil {
		s.where(query, Tables.VfsFolder.Alias, Columns.VfsFolder.Title, s.Title)
	}
	if s.IsFavorite != nil {
		s.where(query, Tables.VfsFolder.Alias, Columns.VfsFolder.IsFavorite, s.IsFavorite)
	}
	if s.CreatedAt != nil {
		s.where(query, Tables.VfsFolder.Alias, Columns.VfsFolder.CreatedAt, s.CreatedAt)
	}
	if s.StatusID != nil {
		s.where(query, Tables.VfsFolder.Alias, Columns.VfsFolder.StatusID, s.StatusID)
	}

	s.apply(query)

	return query
}

func (s *VfsFolderSearch) Q() applier {
	return func(query *orm.Query) (*orm.Query, error) {
		return s.Apply(query), nil
	}
}
