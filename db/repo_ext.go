package db

import (
	"context"
)

type VfsFileParams struct {
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`
}

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

func (s *VfsFileSearch) WithQuery(query *string) *VfsFileSearch {
	if query != nil && *query != "" {
		f := Filter{Field: Columns.VfsFile.Title, Value: *query, SearchType: SearchTypeILike}
		s.WithApply(f.Applier)
	}

	return s
}

func (vr VfsRepo) NextFileID() (int, error) {
	var max int
	_, err := vr.db.Query(&max, `select nextval('"vfsFiles_fileId_seq"')`)
	return max, err
}
