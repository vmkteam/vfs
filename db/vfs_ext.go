package db

import (
	"context"

	"github.com/go-pg/pg/v9"
)

func (vr VfsRepo) FolderBranch(ctx context.Context, folderId int) (list []VfsFolder, err error) {
	_, err = vr.db.QueryContext(ctx, &list, `
WITH RECURSIVE r AS (
   SELECT *, 1 AS level from "vfsFolders"
   WHERE "folderId" = ?
   UNION SELECT ff.*, r.level + 1 AS level
   FROM "vfsFolders" ff
 		JOIN r ON ff."folderId" = r."parentFolderId"
)
SELECT * FROM r ORDER by level DESC`, folderId)

	return
}

func (vfs *VfsFileSearch) WithQuery(query *string) *VfsFileSearch {
	if query != nil && *query != "" {
		vfs.TitleILike = query
	}

	return vfs
}

func (vr VfsRepo) NextFileID() (int, error) {
	var max int
	_, err := vr.db.Query(&max, `select nextval('"vfsFiles_fileId_seq"')`)

	return max, err
}

func (vr VfsRepo) UpdateFilesFolder(ctx context.Context, fileIDs []int64, newFolderId int) (bool, error) {
	q := vr.db.ModelContext(ctx, &VfsFile{FolderID: newFolderId}).
		Column(Columns.VfsFile.FolderID).
		Where("? IN (?)", pg.F(Columns.VfsFile.ID), pg.Ints(fileIDs))

	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

func (vr VfsRepo) DeleteVfsFiles(ctx context.Context, fileIDs []int64) (bool, error) {
	q := vr.db.ModelContext(ctx, &VfsFile{StatusID: StatusDeleted}).
		Column(Columns.VfsFile.StatusID).
		Where("? IN (?)", pg.F(Columns.VfsFile.ID), pg.Ints(fileIDs))

	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}
