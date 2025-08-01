// nolint
package vfs

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/fs"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vmkteam/vfs/db"

	"github.com/bbrks/go-blurhash"
	"github.com/go-pg/pg/v10"
	lru "github.com/hashicorp/golang-lru"
	"github.com/labstack/echo/v4"
	"github.com/vmkteam/embedlog"
	"go.uber.org/atomic"
)

const (
	defaultInterval  = time.Second * 5
	defaultCacheSize = 1024
	httpTimeLayout   = `Mon, 02 Jan 2006 15:04:05 MST`
	DefaultNamespace = "default"
)

type HashIndexer struct {
	embedlog.Logger
	dbc          db.DB
	vfs          VFS
	repo         *db.VfsRepo
	totalWorkers int
	batchSize    uint64
	calcBlurHash bool

	cache    *lru.ARCCache
	t        *time.Ticker
	scanning *atomic.Bool
	indexing *atomic.Bool
}

type HashInfo struct {
	Hash      string
	Extension string
	Width     int
	Height    int
	FileSize  int64
	BlurHash  string
}

type ScanResults struct {
	Scanned  uint64        `json:"scanned"`
	Added    uint64        `json:"added"`
	Duration time.Duration `json:"duration"`
}

type cacheEntry struct {
	data  []byte
	mtime time.Time
}

func NewHashIndexer(sl embedlog.Logger, dbc db.DB, repo *db.VfsRepo, vfs VFS, totalWorkers int, batchSize uint64, calculateBlurHash bool) *HashIndexer {
	cache, _ := lru.NewARC(defaultCacheSize)
	return &HashIndexer{
		Logger:       sl,
		dbc:          dbc,
		repo:         repo,
		vfs:          vfs,
		scanning:     atomic.NewBool(false),
		indexing:     atomic.NewBool(false),
		cache:        cache,
		totalWorkers: totalWorkers,
		batchSize:    batchSize,
		calcBlurHash: calculateBlurHash,
	}
}

func (hi HashIndexer) Start() {
	hi.t = time.NewTicker(defaultInterval)
	for range hi.t.C {
		wg := sync.WaitGroup{}
		for i := 0; i < hi.totalWorkers; i++ {
			wg.Add(1)
			go func() {
				ctx := context.Background()
				rows, err := hi.ProcessQueue(ctx)
				hi.PrintOrErr(ctx, "process queue", err, "rows", rows)
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func (hi HashIndexer) Stop() {
	if hi.t != nil {
		hi.t.Stop()
	}
}

func (hi HashIndexer) IndexFile(ns, relFilepath string) (HashInfo, error) {
	var hash HashInfo

	reader, err := os.Open(hi.vfs.Path(ns, relFilepath))
	if err != nil {
		return hash, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(reader)
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

	// calculate blurhash
	if hi.calcBlurHash {
		bh, err := blurhash.Encode(4, 3, img)
		if err != nil {
			return hash, err
		}
		hash.BlurHash = bh
	}

	return hash, err
}

// ScanFiles reads media folder, detects namespaces & files and loads files into vfsHashes.
func (hi HashIndexer) ScanFiles(ctx context.Context) (r ScanResults, err error) {
	// forbid running FS scan in parallel
	if hi.scanning.Load() {
		return r, errors.New("already scanning")
	}
	hi.scanning.Store(true)
	defer hi.scanning.Store(false)

	// pipe for CSV -> temp table
	pr, pw := io.Pipe()
	cw := csv.NewWriter(pw)
	cw.Comma = ';'

	// save files hashes and sizes to DB
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			_ = pr.Close()
		}()
		err = hi.dbc.RunInTransaction(ctx, func(tx *pg.Tx) error {
			if err := hi.dbc.CreateTempHashesTable(ctx, tx); err != nil {
				return err
			}
			scanned, err := hi.dbc.CopyHashesFromSTDIN(tx, pr)
			if err != nil {
				return err
			}
			r.Scanned = uint64(scanned)
			updated, duration, err := hi.dbc.UpsertHashesTable(ctx, tx)
			if err != nil {
				return err
			}
			r.Added = uint64(updated)
			r.Duration = duration
			return nil
		})
	}()

	// scan files
	separator := string(filepath.Separator)
	rootDir := strings.TrimSuffix(hi.vfs.cfg.Path, separator) + separator
	if err := filepath.Walk(hi.vfs.cfg.Path, hi.walkFn(rootDir, cw)); err != nil {
		cw.Flush()
		_ = pw.Close()
		return r, err
	}
	cw.Flush()
	_ = pw.Close()
	wg.Wait()

	return r, err
}

type ScanFilesResponse struct {
	ScanResults `json:",omitempty"`
	Error       string `json:"error,omitempty"` // error message
}

func (hi HashIndexer) ScanFilesHandler(c echo.Context) error {
	sr, err := hi.ScanFiles(context.Background())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ScanFilesResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, ScanFilesResponse{ScanResults: sr})
}

// ProcessQueue gets not indexed data from vfsHashes, index and saves data to db.
func (hi HashIndexer) ProcessQueue(ctx context.Context) (int, error) {
	var rows int
	err := hi.dbc.RunInTransaction(ctx, func(tx *pg.Tx) error {
		repo := hi.repo.WithTransaction(tx)
		// get data from queue
		list, err := repo.HashesForUpdate(ctx, hi.batchSize)
		if err != nil {
			return fmt.Errorf("hash for update failed: %w", err)
		}
		if len(list) < 1 {
			return nil
		}

		// index file
		now := time.Now().UTC()
		for i, v := range list {
			ns := v.Namespace
			if ns == DefaultNamespace {
				ns = ""
			}
			list[i].IndexedAt = &now
			info, err := hi.IndexFile(ns, NewFileHash(v.Hash, v.Extension).File())
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					// skip in case file not found (not downloaded yet)
					list[i].IndexedAt = nil
				}
				list[i].Error = err.Error()
				continue
			}
			list[i].Height = info.Height
			list[i].Width = info.Width
			list[i].Blurhash = &info.BlurHash
		}

		// save data to db
		_, err = tx.
			ModelContext(ctx, &list).
			Column(
				db.Columns.VfsHash.Hash,
				db.Columns.VfsHash.Namespace,
				db.Columns.VfsHash.Height,
				db.Columns.VfsHash.Width,
				db.Columns.VfsHash.IndexedAt,
				db.Columns.VfsHash.Blurhash,
				db.Columns.VfsHash.Error,
			).
			Update()

		rows = len(list)
		return err
	})
	return rows, err
}

func (hi HashIndexer) Preview(c echo.Context) error {
	nsp, file := c.Param("ns"), c.Param("file")
	file = strings.TrimSuffix(file, filepath.Ext(file))
	ns := DefaultNamespace
	if nsp != "" && hi.vfs.IsValidNamespace(nsp) {
		ns = nsp
	}
	key := cacheKey(ns, file)
	entry, ok := hi.cache.Get(key)
	if ok {
		return writePreview(entry.(cacheEntry), c)
	}

	hash, err := hi.repo.OneVfsHash(context.Background(), &db.VfsHashSearch{
		Hash:      &file,
		Namespace: &ns,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if hash == nil {
		return c.String(http.StatusNotFound, "hash not found")
	}

	if hash.Width == 0 || hash.Height == 0 || hash.Blurhash == nil || *hash.Blurhash == "" {
		return c.String(http.StatusNotFound, "hash not indexed yet")
	}

	if imsTime, err := time.Parse(httpTimeLayout, c.Request().Header.Get("If-Modified-Since")); err == nil {
		if hash.IndexedAt.Before(imsTime) {
			return c.NoContent(http.StatusNotModified)
		}
	}

	newWidth := 32
	newHeight := int(math.Round(float64(newWidth*hash.Height) / float64(hash.Width)))

	img, err := blurhash.Decode(*hash.Blurhash, newWidth, newHeight, 1)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	buf := bytes.NewBuffer(nil)
	if err := png.Encode(buf, img); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	entry = cacheEntry{data: buf.Bytes(), mtime: hash.IndexedAt.UTC()}
	hi.cache.Add(key, entry)
	return writePreview(entry.(cacheEntry), c)
}

func writePreview(e cacheEntry, c echo.Context) error {
	c.Response().Header().Set("Last-Modified", e.mtime.UTC().Format(httpTimeLayout))
	c.Response().Header().Set("Content-Length", strconv.Itoa(len(e.data)))
	c.Response().Header().Set("Cache-Control", "public, max-age=31536000;")

	return c.Blob(http.StatusOK, "image/png", e.data)
}

func cacheKey(ns, hash string) string {
	return ns + "|" + hash
}

func (hi HashIndexer) walkFn(rootDir string, cw *csv.Writer) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !info.Mode().IsRegular() {
			return nil
		}
		relPath := strings.TrimPrefix(path, rootDir)
		ns := getNs(hi.vfs.cfg.Namespaces, relPath)
		if !isHashFile(ns, relPath) {
			return nil
		}

		ext := filepath.Ext(relPath)
		baseName := strings.TrimSuffix(filepath.Base(relPath), ext)
		if len(baseName) > 40 {
			baseName = baseName[:40]
		}

		if err := cw.Write([]string{
			baseName,
			ns,
			strconv.FormatInt(info.Size(), 10),
			strings.TrimPrefix(ext, "."),
		}); err != nil {
			return err
		}
		return nil
	}
}

func getNs(namespaces []string, path string) string {
	for _, ns := range namespaces {
		if ns != "" && strings.HasPrefix(path, ns) {
			return ns
		}
	}
	return ""
}

// isHashFile checks if file path has a namespace format.
// e.g. "7/0c/70c565ef460af43688b7ee6251028db9.jpg"
func isHashFile(ns string, path string) bool {
	if len(ns) > 0 && len(path) > len(ns) {
		path = path[len(ns)+1:]
	}
	path = strings.TrimSuffix(path, filepath.Ext(path))
	if len(path) != 37 {
		return false
	}
	if path[1] != filepath.Separator || path[4] != filepath.Separator {
		return false
	}
	if path[0] != path[5] || path[2:4] != path[6:8] {
		return false
	}
	for _, c := range path[5:37] {
		if !isHex(c) {
			return false
		}
	}
	return true
}

func isHex(c int32) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')
}
