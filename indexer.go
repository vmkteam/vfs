package vfs

import (
	"context"
	"image"
	"io"
	"os"

	"github.com/vmkteam/vfs/db"

	"github.com/bbrks/go-blurhash"
)

type HashInfo struct {
	Hash      string
	Extension string
	Width     int
	Height    int
	FileSize  int64
	BlurHash  string
}

type HashIndexer struct {
	dbc db.DB
	vfs VFS
}

func NewHashIndexer(dbc db.DB, vfs VFS) *HashIndexer {
	return &HashIndexer{dbc: dbc, vfs: vfs}
}

// Index indexes file and save results to DB.
func (hi HashIndexer) Index(ctx context.Context, filepath string) error {
	// index files
	// save to db
	return nil
}

func (hi HashIndexer) IndexFile(ns, relFilepath string) (HashInfo, error) {
	var hash HashInfo

	reader, err := os.Open(hi.vfs.Path(ns, relFilepath))
	if err != nil {
		return hash, err
	}
	f, err := reader.Stat()
	if err != nil {
		return hash, err
	}

	// get file size
	hash.FileSize = f.Size()

	// detect image size
	im, _, err := image.DecodeConfig(reader)
	if err != nil {
		return hash, err
	}

	hash.Width = im.Width
	hash.Height = im.Height

	// detect image hash
	_, err = reader.Seek(0, io.SeekStart)
	if err != nil {
		return hash, err
	}

	img, _, err := image.Decode(reader)
	if err != nil {
		return hash, err
	}
	//return hash, nil

	bh, err := blurhash.Encode(4, 3, img)
	if err != nil {
		return hash, err
	}
	hash.BlurHash = bh

	return hash, err
}

// InitQueue reads media folder, detects namespaces & files and loads files into vfsHashes.
func (hi HashIndexer) InitQueue(ctx context.Context) error {
	// detects namespaces
	// get files
	// load to db

	return nil
}

// ProcessQueue gets not indexed data from vfsHashes, index and saves data to db.
func (hi HashIndexer) ProcessQueue(ctx context.Context) error {
	// get data from queue
	// index file
	// save data to db
	return nil
}
