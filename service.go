package vfs

import (
	"context"
	"net/http"
	"vfs/db"

	"github.com/semrush/zenrpc"
)

var (
	ErrInternal    = httpAsRpcError(http.StatusInternalServerError)
	ErrNotFound    = httpAsRpcError(http.StatusNotFound)
	ErrInvalidSort = zenrpc.NewStringError(http.StatusBadRequest, "invalid sort field")
)

func httpAsRpcError(code int) *zenrpc.Error {
	return zenrpc.NewStringError(code, http.StatusText(code))
}

func InternalError(err error) *zenrpc.Error {
	return zenrpc.NewError(http.StatusInternalServerError, err)
}

//go:generate zenrpc
type Service struct {
	zenrpc.Service
	vfs  VFS
	repo db.VfsRepo
}

func NewService(repo db.VfsRepo, vfs VFS) Service {
	return Service{repo: repo, vfs: vfs}
}

func (s Service) folderByID(ctx context.Context, id int) (*db.VfsFolder, error) {
	dbc, err := s.repo.VfsFolderByID(ctx, id, s.repo.FullVfsFolder())
	if err != nil {
		return nil, InternalError(err)
	} else if dbc == nil {
		return nil, ErrNotFound
	}
	return dbc, nil
}

// Get Folder with Sub Folders.
//zenrpc:rootFolderId=1
//zenrpc:404 Folder not found
func (s Service) GetFolder(ctx context.Context, rootFolderId int) (*Folder, error) {
	dbf, err := s.folderByID(ctx, rootFolderId)
	if err != nil {
		return nil, err
	}

	childFolders, err := s.repo.VfsFoldersByFilters(ctx, &db.VfsFolderSearch{ParentFolderID: &dbf.ID}, db.PagerNoLimit)
	if err != nil {
		return nil, InternalError(err)
	}

	return NewFullFolder(dbf, childFolders), nil
}

// Get Folder Branch
func (s Service) GetFolderBranch(ctx context.Context, folderId int) ([]Folder, error) {
	dbf, err := s.folderByID(ctx, folderId)
	if err != nil {
		return nil, err
	}

	list, err := s.repo.FolderBranch(ctx, dbf.ID)
	if err != nil {
		return nil, InternalError(err)
	}

	folders := make([]Folder, 0, len(list))
	for i := 0; i < len(list); i++ {
		folders = append(folders, *NewFolder(&list[i]))
	}
	return folders, nil
}

// Get Files
//zenrpc:folderId root folder id
//zenrpc:query file name
//zenrpc:sortField="createdAt" createdAt, title or fileSize
//zenrpc:isDescending=true asc = false, desc = true
//zenrpc:page=0 current page
//zenrpc:pageSize=100 current pageSize
func (s Service) GetFiles(ctx context.Context, folderId int, query *string, sortField string, isDescending bool, page, pageSize int) ([]File, error) {
	dbf, err := s.folderByID(ctx, folderId)
	if err != nil {
		return nil, err
	}

	if sortField != "createdAt" && sortField != "title" && sortField != "fileSize" {
		return nil, ErrInvalidSort
	}

	// set sort
	sort := db.SortField{Column: sortField, Direction: db.SortAsc}
	if isDescending {
		sort.Direction = db.SortDesc
	}

	search := (&db.VfsFileSearch{FolderID: &dbf.ID}).WithQuery(query)
	list, err := s.repo.VfsFilesByFilters(ctx, search, db.Pager{Page: page, PageSize: pageSize}, db.WithSort(sort))
	if err != nil {
		return nil, InternalError(err)
	}

	files := make([]File, 0, len(list))
	for i := 0; i < len(list); i++ {
		files = append(files, *NewFile(&list[i], s.vfs.WebPath(""))) // TODO ns?
	}
	return files, nil
}

// Count Files
//zenrpc:folderId root folder id
//zenrpc:query file name
func (s Service) CountFiles(ctx context.Context, folderId int, query *string) (int, error) {
	search := (&db.VfsFileSearch{FolderID: &folderId}).WithQuery(query)
	count, err := s.repo.CountVfsFiles(ctx, search)
	if err != nil {
		return 0, InternalError(err)
	}

	return count, nil
}

// Move Files
func (s Service) MoveFiles(ctx context.Context, fileIds []int, destinationFolderId int) (bool, error) {
	return false, nil
}

// Delete Files
func (s Service) DeleteFiles(ctx context.Context, fileIds []int) (bool, error) {
	return false, nil
}

// Rename File on Server
func (s Service) SetFilePhysicalName(ctx context.Context, fileIds []int) (bool, error) {
	return false, nil
}

// Search Folder by File Id
func (s Service) SearchFolderByFileId(ctx context.Context, fileId int) (*File, error) {
	return nil, nil
}

// Search Folder by Filename
func (s Service) SearchFolderByFile(ctx context.Context, filename string) (*File, error) {
	return nil, nil
}

// Get Favorites
func (s Service) GetFavorites(ctx context.Context) ([]Folder, error) {
	b := true
	list, err := s.repo.VfsFoldersByFilters(ctx, &db.VfsFolderSearch{IsFavorite: &b}, db.PagerNoLimit)
	if err != nil {
		return nil, InternalError(err)
	}

	folders := make([]Folder, 0, len(list))
	for i := 0; i < len(list); i++ {
		folders = append(folders, *NewFolder(&list[i]))
	}
	return folders, nil
}

// Manage Favorite Folders
func (s Service) ManageFavorites(ctx context.Context, folderId int, isInFavorites bool) (bool, error) {
	return false, nil
}

// Create Folder
func (s Service) CreateFolder(ctx context.Context, rootFolderId int, name string) (bool, error) {
	return false, nil
}

// Delete Folder
func (s Service) DeleteFolder(ctx context.Context, folderId int) (bool, error) {
	return false, nil
}

// Move Folder
func (s Service) MoveFolder(ctx context.Context, folderId, destinationFolderId int) (bool, error) {
	return false, nil
}

// Move Folder
func (s Service) RenameFolder(ctx context.Context, folderId int, name string) (bool, error) {
	return false, nil
}

func (s Service) HelpUpload() HelpUploadResponse {
	return HelpUploadResponse{
		Temp: HelpUploadItem{
			URL:    "/vfs/upload/hash",
			Params: []string{s.vfs.cfg.UploadFormName},
		},
		Queue: HelpUploadItem{
			URL:    "/vfs/upload/file",
			Params: []string{s.vfs.cfg.UploadFormName, "folderId"},
		},
	}
}
