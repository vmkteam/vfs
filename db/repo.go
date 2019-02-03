package db

import (
	"context"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

type VfsRepo struct {
	db      orm.DB
	filters map[string][]Filter
	sort    map[string][]SortField
	join    map[string][]string
}

// NewVfsFileRepo returns new repository for vfs files and folders.
func NewVfsFileRepo(db orm.DB) VfsRepo {
	return VfsRepo{
		db: db,
		filters: map[string][]Filter{
			Tables.VfsFile.Name:   {StatusFilter},
			Tables.VfsFolder.Name: {StatusFilter},
		},
		sort: map[string][]SortField{
			Tables.VfsFile.Name:   {{Column: Columns.VfsFile.Title, Direction: SortAsc}},
			Tables.VfsFolder.Name: {{Column: Columns.VfsFolder.Title, Direction: SortAsc}},
		},
		join: map[string][]string{
			Tables.VfsFile.Name:   {TableColumns, Columns.VfsFile.Folder},
			Tables.VfsFolder.Name: {TableColumns, Columns.VfsFolder.ParentFolder},
		},
	}
}

// WithTransaction is a function that wraps VfsRepo with pg.Tx transaction.
func (vr VfsRepo) WithTransaction(tx orm.DB) VfsRepo {
	vr.db = tx
	return vr
}

// FullVfsFile returns full joins for vfs files.
func (vr VfsRepo) FullVfsFile() OpFunc {
	return WithColumns(vr.join[Tables.VfsFile.Name]...)
}

// DefaultVfsFileSort returns default sort for vfs files.
func (vr VfsRepo) DefaultVfsFileSort() OpFunc {
	return WithSort(vr.sort[Tables.VfsFile.Name]...)
}

// VfsFileByID is a function that returns VfsFile by ID or nil.
func (vr VfsRepo) VfsFileByID(ctx context.Context, id int, ops ...OpFunc) (*VfsFile, error) {
	return vr.OneVfsFile(ctx, &VfsFileSearch{ID: &id}, ops...)
}

// OneVfsFile is a function that returns one VfsFile by filters. It could return pg.ErrMultiRows.
func (vr VfsRepo) OneVfsFile(ctx context.Context, search *VfsFileSearch, ops ...OpFunc) (*VfsFile, error) {
	c := &VfsFile{}
	err := buildQuery(ctx, vr.db, c, search, vr.filters[Tables.VfsFile.Name], PagerTwo, ops...).Select()

	switch err {
	case pg.ErrMultiRows:
		return nil, err
	case pg.ErrNoRows:
		return nil, nil
	}

	return c, err
}

// VfsFilesByFilters returns vfs files.
func (vr VfsRepo) VfsFilesByFilters(ctx context.Context, search *VfsFileSearch, pager Pager, ops ...OpFunc) (list []VfsFile, err error) {
	err = buildQuery(ctx, vr.db, &list, search, vr.filters[Tables.VfsFile.Name], pager, ops...).Select()
	return
}

// CountVfsFiles returns count of vfs files.
func (vr VfsRepo) CountVfsFiles(ctx context.Context, search *VfsFileSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, vr.db, &VfsFile{}, search, vr.filters[Tables.VfsFile.Name], PagerOne, ops...).Count()
}

// AddVfsFile adds file to DB.
func (vr VfsRepo) AddVfsFile(ctx context.Context, c *VfsFile, ops ...OpFunc) (*VfsFile, error) {
	q := vr.db.ModelContext(ctx, c)
	applyOps(q, ops...)
	_, err := q.Insert()

	return c, err
}

// UpdateVfsFile updates file in DB.
func (vr *VfsRepo) UpdateVfsFile(ctx context.Context, c *VfsFile, ops ...OpFunc) (bool, error) {
	q := vr.db.ModelContext(ctx, c).WherePK()
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteVfsFile set statusId to deleted in DB.
func (vr VfsRepo) DeleteVfsFile(ctx context.Context, id int, ops ...OpFunc) (bool, error) {
	c := &VfsFile{ID: id, StatusID: StatusDeleted}
	return vr.UpdateVfsFile(ctx, c, WithColumns(Columns.VfsFile.StatusID))
}

// FullVfsFolder returns full joins for vfs files.
func (vr VfsRepo) FullVfsFolder() OpFunc {
	return WithColumns(vr.join[Tables.VfsFolder.Name]...)
}

// DefaultVfsFolderSort returns default sort for vfs files.
func (vr VfsRepo) DefaultVfsFolderSort() OpFunc {
	return WithSort(vr.sort[Tables.VfsFolder.Name]...)
}

// VfsFolderByID is a function that returns VfsFolder by ID or nil.
func (vr VfsRepo) VfsFolderByID(ctx context.Context, id int, ops ...OpFunc) (*VfsFolder, error) {
	return vr.OneVfsFolder(ctx, &VfsFolderSearch{ID: &id}, ops...)
}

// OneVfsFolder is a function that returns one VfsFolder by filters. It could return pg.ErrMultiRows.
func (vr VfsRepo) OneVfsFolder(ctx context.Context, search *VfsFolderSearch, ops ...OpFunc) (*VfsFolder, error) {
	c := &VfsFolder{}
	err := buildQuery(ctx, vr.db, c, search, vr.filters[Tables.VfsFolder.Name], PagerTwo, ops...).Select()

	switch err {
	case pg.ErrMultiRows:
		return nil, err
	case pg.ErrNoRows:
		return nil, nil
	}

	return c, err
}

// VfsFoldersByFilters returns vfs files.
func (vr VfsRepo) VfsFoldersByFilters(ctx context.Context, search *VfsFolderSearch, pager Pager, ops ...OpFunc) (list []VfsFolder, err error) {
	err = buildQuery(ctx, vr.db, &list, search, vr.filters[Tables.VfsFolder.Name], pager, ops...).Select()
	return
}

// CountVfsFolders returns count of vfs files.
func (vr VfsRepo) CountVfsFolders(ctx context.Context, search *VfsFolderSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, vr.db, &VfsFolder{}, search, vr.filters[Tables.VfsFolder.Name], PagerOne, ops...).Count()
}

// AddVfsFolder adds file to DB.
func (vr VfsRepo) AddVfsFolder(ctx context.Context, c *VfsFolder, ops ...OpFunc) (*VfsFolder, error) {
	q := vr.db.ModelContext(ctx, c)
	applyOps(q, ops...)
	_, err := q.Insert()

	return c, err
}

// UpdateVfsFolder updates file in DB.
func (vr *VfsRepo) UpdateVfsFolder(ctx context.Context, c *VfsFolder, ops ...OpFunc) (bool, error) {
	q := vr.db.ModelContext(ctx, c).WherePK()
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteVfsFolder set statusId to deleted in DB.
func (vr VfsRepo) DeleteVfsFolder(ctx context.Context, id int, ops ...OpFunc) (bool, error) {
	c := &VfsFolder{ID: id, StatusID: StatusDeleted}
	return vr.UpdateVfsFolder(ctx, c, WithColumns(Columns.VfsFolder.StatusID))
}
