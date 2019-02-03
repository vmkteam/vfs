//nolint
//lint:file-ignore U1000 ignore unused code, it's generated
package db

import (
	"time"
)

var Columns = struct {
	VfsFile struct {
		ID, FolderID, Title, Path, Params, IsFavorite, MimeType, FileSize, FileExists, StatusID, CreatedAt string

		Folder string
	}
	VfsFolder struct {
		ID, ParentFolderID, Title, IsFavorite, CreatedAt, StatusID string

		ParentFolder string
	}
}{
	VfsFile: struct {
		ID, FolderID, Title, Path, Params, IsFavorite, MimeType, FileSize, FileExists, StatusID, CreatedAt string

		Folder string
	}{
		ID:         "fileId",
		FolderID:   "folderId",
		Title:      "title",
		Path:       "path",
		Params:     "params",
		IsFavorite: "isFavorite",
		MimeType:   "mimeType",
		FileSize:   "fileSize",
		FileExists: "fileExists",
		StatusID:   "statusId",
		CreatedAt:  "createdAt",

		Folder: "Folder",
	},
	VfsFolder: struct {
		ID, ParentFolderID, Title, IsFavorite, CreatedAt, StatusID string

		ParentFolder string
	}{
		ID:             "folderId",
		ParentFolderID: "parentFolderId",
		Title:          "title",
		IsFavorite:     "isFavorite",
		CreatedAt:      "createdAt",
		StatusID:       "statusId",

		ParentFolder: "ParentFolder",
	},
}

var Tables = struct {
	VfsFile struct {
		Name, Alias string
	}
	VfsFolder struct {
		Name, Alias string
	}
}{
	VfsFile: struct {
		Name, Alias string
	}{
		Name:  "vfsFiles",
		Alias: "t",
	},
	VfsFolder: struct {
		Name, Alias string
	}{
		Name:  "vfsFolders",
		Alias: "t",
	},
}

type VfsFile struct {
	tableName struct{} `sql:"\"vfsFiles\",alias:t" pg:",discard_unknown_columns"`

	ID         int            `sql:"fileId,pk"`
	FolderID   int            `sql:"folderId,notnull"`
	Title      string         `sql:"title,notnull"`
	Path       string         `sql:"path,notnull"`
	Params     *VfsFileParams `sql:"params"`
	IsFavorite *bool          `sql:"isFavorite"`
	MimeType   string         `sql:"mimeType,notnull"`
	FileSize   *int           `sql:"fileSize"`
	FileExists bool           `sql:"fileExists,notnull"`
	StatusID   int            `sql:"statusId,notnull"`
	CreatedAt  time.Time      `sql:"createdAt,notnull"`

	Folder *VfsFolder `pg:"fk:folderId"`
}

type VfsFolder struct {
	tableName struct{} `sql:"\"vfsFolders\",alias:t" pg:",discard_unknown_columns"`

	ID             int       `sql:"folderId,pk"`
	ParentFolderID *int      `sql:"parentFolderId"`
	Title          string    `sql:"title,notnull"`
	IsFavorite     *bool     `sql:"isFavorite"`
	CreatedAt      time.Time `sql:"createdAt,notnull"`
	StatusID       int       `sql:"statusId,notnull"`

	ParentFolder *VfsFolder `pg:"fk:parentFolderId"`
}
