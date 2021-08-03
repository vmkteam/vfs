// Code generated by mfd-generator; DO NOT EDIT.

//nolint
//lint:file-ignore U1000 ignore unused code, it's generated
package db

import (
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
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
	query.Where(condition, pg.Ident(table), pg.Ident(field), value)
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

	ID            *int
	FolderID      *int
	Title         *string
	Path          *string
	Params        *VfsFileParams
	IsFavorite    *bool
	MimeType      *string
	FileSize      *int
	FileExists    *bool
	CreatedAt     *time.Time
	StatusID      *int
	IDs           []int
	TitleILike    *string
	PathILike     *string
	MimeTypeILike *string
}

func (vfs *VfsFileSearch) Apply(query *orm.Query) *orm.Query {
	if vfs == nil {
		return query
	}
	if vfs.ID != nil {
		vfs.where(query, Tables.VfsFile.Alias, Columns.VfsFile.ID, vfs.ID)
	}
	if vfs.FolderID != nil {
		vfs.where(query, Tables.VfsFile.Alias, Columns.VfsFile.FolderID, vfs.FolderID)
	}
	if vfs.Title != nil {
		vfs.where(query, Tables.VfsFile.Alias, Columns.VfsFile.Title, vfs.Title)
	}
	if vfs.Path != nil {
		vfs.where(query, Tables.VfsFile.Alias, Columns.VfsFile.Path, vfs.Path)
	}
	if vfs.Params != nil {
		vfs.where(query, Tables.VfsFile.Alias, Columns.VfsFile.Params, vfs.Params)
	}
	if vfs.IsFavorite != nil {
		vfs.where(query, Tables.VfsFile.Alias, Columns.VfsFile.IsFavorite, vfs.IsFavorite)
	}
	if vfs.MimeType != nil {
		vfs.where(query, Tables.VfsFile.Alias, Columns.VfsFile.MimeType, vfs.MimeType)
	}
	if vfs.FileSize != nil {
		vfs.where(query, Tables.VfsFile.Alias, Columns.VfsFile.FileSize, vfs.FileSize)
	}
	if vfs.FileExists != nil {
		vfs.where(query, Tables.VfsFile.Alias, Columns.VfsFile.FileExists, vfs.FileExists)
	}
	if vfs.CreatedAt != nil {
		vfs.where(query, Tables.VfsFile.Alias, Columns.VfsFile.CreatedAt, vfs.CreatedAt)
	}
	if vfs.StatusID != nil {
		vfs.where(query, Tables.VfsFile.Alias, Columns.VfsFile.StatusID, vfs.StatusID)
	}
	if len(vfs.IDs) > 0 {
		Filter{Columns.VfsFile.ID, vfs.IDs, SearchTypeArray, false}.Apply(query)
	}
	if vfs.TitleILike != nil {
		Filter{Columns.VfsFile.Title, *vfs.TitleILike, SearchTypeILike, false}.Apply(query)
	}
	if vfs.PathILike != nil {
		Filter{Columns.VfsFile.Path, *vfs.PathILike, SearchTypeILike, false}.Apply(query)
	}
	if vfs.MimeTypeILike != nil {
		Filter{Columns.VfsFile.MimeType, *vfs.MimeTypeILike, SearchTypeILike, false}.Apply(query)
	}

	vfs.apply(query)

	return query
}

func (vfs *VfsFileSearch) Q() applier {
	return func(query *orm.Query) (*orm.Query, error) {
		if vfs == nil {
			return query, nil
		}
		return vfs.Apply(query), nil
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
	IDs            []int
	TitleILike     *string
}

func (vfs *VfsFolderSearch) Apply(query *orm.Query) *orm.Query {
	if vfs == nil {
		return query
	}
	if vfs.ID != nil {
		vfs.where(query, Tables.VfsFolder.Alias, Columns.VfsFolder.ID, vfs.ID)
	}
	if vfs.ParentFolderID != nil {
		vfs.where(query, Tables.VfsFolder.Alias, Columns.VfsFolder.ParentFolderID, vfs.ParentFolderID)
	}
	if vfs.Title != nil {
		vfs.where(query, Tables.VfsFolder.Alias, Columns.VfsFolder.Title, vfs.Title)
	}
	if vfs.IsFavorite != nil {
		vfs.where(query, Tables.VfsFolder.Alias, Columns.VfsFolder.IsFavorite, vfs.IsFavorite)
	}
	if vfs.CreatedAt != nil {
		vfs.where(query, Tables.VfsFolder.Alias, Columns.VfsFolder.CreatedAt, vfs.CreatedAt)
	}
	if vfs.StatusID != nil {
		vfs.where(query, Tables.VfsFolder.Alias, Columns.VfsFolder.StatusID, vfs.StatusID)
	}
	if len(vfs.IDs) > 0 {
		Filter{Columns.VfsFolder.ID, vfs.IDs, SearchTypeArray, false}.Apply(query)
	}
	if vfs.TitleILike != nil {
		Filter{Columns.VfsFolder.Title, *vfs.TitleILike, SearchTypeILike, false}.Apply(query)
	}

	vfs.apply(query)

	return query
}

func (vfs *VfsFolderSearch) Q() applier {
	return func(query *orm.Query) (*orm.Query, error) {
		if vfs == nil {
			return query, nil
		}
		return vfs.Apply(query), nil
	}
}

type VfsHashSearch struct {
	search

	Hash           *string
	Namespace      *string
	Extension      *string
	FileSize       *int
	Width          *int
	Height         *int
	Blurhash       *string
	CreatedAt      *time.Time
	IndexedAt      *time.Time
	Error          *string
	Hashes         []string
	HashILike      *string
	Namespaces     []string
	NamespaceILike *string
	ExtensionILike *string
	BlurhashILike  *string
	ErrorILike     *string
}

func (vhs *VfsHashSearch) Apply(query *orm.Query) *orm.Query {
	if vhs == nil {
		return query
	}
	if vhs.Hash != nil {
		vhs.where(query, Tables.VfsHash.Alias, Columns.VfsHash.Hash, vhs.Hash)
	}
	if vhs.Namespace != nil {
		vhs.where(query, Tables.VfsHash.Alias, Columns.VfsHash.Namespace, vhs.Namespace)
	}
	if vhs.Extension != nil {
		vhs.where(query, Tables.VfsHash.Alias, Columns.VfsHash.Extension, vhs.Extension)
	}
	if vhs.FileSize != nil {
		vhs.where(query, Tables.VfsHash.Alias, Columns.VfsHash.FileSize, vhs.FileSize)
	}
	if vhs.Width != nil {
		vhs.where(query, Tables.VfsHash.Alias, Columns.VfsHash.Width, vhs.Width)
	}
	if vhs.Height != nil {
		vhs.where(query, Tables.VfsHash.Alias, Columns.VfsHash.Height, vhs.Height)
	}
	if vhs.Blurhash != nil {
		vhs.where(query, Tables.VfsHash.Alias, Columns.VfsHash.Blurhash, vhs.Blurhash)
	}
	if vhs.CreatedAt != nil {
		vhs.where(query, Tables.VfsHash.Alias, Columns.VfsHash.CreatedAt, vhs.CreatedAt)
	}
	if vhs.IndexedAt != nil {
		vhs.where(query, Tables.VfsHash.Alias, Columns.VfsHash.IndexedAt, vhs.IndexedAt)
	}
	if vhs.Error != nil {
		vhs.where(query, Tables.VfsHash.Alias, Columns.VfsHash.Error, vhs.Error)
	}
	if len(vhs.Hashes) > 0 {
		Filter{Columns.VfsHash.Hash, vhs.Hashes, SearchTypeArray, false}.Apply(query)
	}
	if vhs.HashILike != nil {
		Filter{Columns.VfsHash.Hash, *vhs.HashILike, SearchTypeILike, false}.Apply(query)
	}
	if len(vhs.Namespaces) > 0 {
		Filter{Columns.VfsHash.Namespace, vhs.Namespaces, SearchTypeArray, false}.Apply(query)
	}
	if vhs.NamespaceILike != nil {
		Filter{Columns.VfsHash.Namespace, *vhs.NamespaceILike, SearchTypeILike, false}.Apply(query)
	}
	if vhs.ExtensionILike != nil {
		Filter{Columns.VfsHash.Extension, *vhs.ExtensionILike, SearchTypeILike, false}.Apply(query)
	}
	if vhs.BlurhashILike != nil {
		Filter{Columns.VfsHash.Blurhash, *vhs.BlurhashILike, SearchTypeILike, false}.Apply(query)
	}
	if vhs.ErrorILike != nil {
		Filter{Columns.VfsHash.Error, *vhs.ErrorILike, SearchTypeILike, false}.Apply(query)
	}

	vhs.apply(query)

	return query
}

func (vhs *VfsHashSearch) Q() applier {
	return func(query *orm.Query) (*orm.Query, error) {
		if vhs == nil {
			return query, nil
		}
		return vhs.Apply(query), nil
	}
}
