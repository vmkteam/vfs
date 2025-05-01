package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/vmkteam/vfs"
	"github.com/vmkteam/vfs/db"
	"github.com/vmkteam/vfs/internal/app"

	"github.com/go-pg/pg/v10"
	"github.com/namsral/flag"
	"github.com/vmkteam/embedlog"
)

const (
	appName = "vfssrv"
)

var (
	fs               = flag.NewFlagSetWithEnvPrefix(os.Args[0], "VFS", 0)
	flAddr           = fs.String("addr", "0.0.0.0:9999", "listen address")
	flDir            = fs.String("dir", "testdata", "storage path")
	flNamespaces     = fs.String("ns", "items,test", "namespaces, separated by comma")
	flWebPath        = fs.String("webpath", "/media/", "web path to files")
	flPreviewPath    = fs.String("preview-path", "/media/small/", "preview path to image files")
	flExtensions     = fs.String("ext", "jpg,jpeg,png,gif", "allowed file extensions for hash upload, separated by comma")
	flMimeTypes      = fs.String("mime", "image/jpeg,image/png,image/gif", "allowed mime types for hash upload, separated by comma (use * for any)")
	flDBConn         = fs.String("conn", "postgresql://localhost:5432/vfs?sslmode=disable", "database connection dsn")
	flJWTKey         = fs.String("jwt-key", "QuiuNae9OhzoKohcee0h", "JWT key")
	flJWTHeader      = fs.String("jwt-header", "AuthorizationJWT", "JWT header")
	flFileSize       = fs.Int64("maxsize", 32<<20, "max file size in bytes")
	flVerbose        = fs.Bool("verbose", false, "enable debug output")
	flJSONLogs       = fs.Bool("json", false, "enable json output")
	flDev            = fs.Bool("dev", false, "enable dev mode")
	flIndex          = fs.Bool("index", false, "index files on start: width, height, blurhash")
	flIndexBlurhash  = fs.Bool("index-blurhash", true, "calculate blurhash (could be long operation on large files)")
	flIndexWorkers   = fs.Int("index-workers", runtime.NumCPU()/2, "total running indexer workers, default is cores/2")
	flIndexBatchSize = fs.Uint64("index-batch-size", 64, "indexer batch size for files, default is 64")
)

func main() {
	err := fs.Parse(os.Args[1:])
	exitOnError(err)

	// setup logger
	sl, ctx := embedlog.NewLogger(*flVerbose, *flJSONLogs), context.Background()
	if *flDev {
		sl = embedlog.NewDevLogger()
	}
	slog.SetDefault(sl.Log()) // set default logger

	// init config
	cfg := app.Config{
		Server: app.ServerConfig{
			Addr:           *flAddr,
			IsDevel:        *flVerbose,
			JWTHeader:      *flJWTHeader,
			JWTKey:         *flJWTKey,
			Index:          *flIndex,
			IndexBlurhash:  *flIndexBlurhash,
			IndexWorkers:   *flIndexWorkers,
			IndexBatchSize: *flIndexBatchSize,
		},
		VFS: vfs.Config{
			MaxFileSize:    *flFileSize,
			Path:           *flDir,
			WebPath:        *flWebPath,
			PreviewPath:    *flPreviewPath,
			UploadFormName: "Filedata",
			Namespaces:     strings.Split(*flNamespaces, ","),
			Extensions:     strings.Split(*flExtensions, ","),
			MimeTypes:      strings.Split(*flMimeTypes, ","),
			Database:       nil,
		},
	}

	// connect to DB
	var dbc *pg.DB
	if flDBConn != nil && *flDBConn != "" {
		cfg.Database, err = pg.ParseURL(*flDBConn)
		exitOnError(err)

		dbc = pg.Connect(cfg.Database)
		exitOnError(dbc.Ping(ctx))
		if *flDev {
			dbc.AddQueryHook(db.NewQueryLogger(sl))
		}
	}

	// create app
	a, err := app.New(appName, sl, cfg, dbc)
	exitOnError(err)

	sl.Print(ctx, "starting", "app", appName, "version", appVersion(), "addr", cfg.Server.Addr, "jwtHeader", cfg.Server.JWTHeader)
	sl.Print(ctx, "app features", "rpc", dbc != nil, "indexer", cfg.Server.Index, "indexBlurhash", cfg.Server.Index && cfg.Server.IndexBlurhash)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// run app
	go func() {
		if err := a.Run(ctx); err != nil {
			a.Print(ctx, "shutting down http server", "err", err)
		}
	}()
	<-quit
	a.Shutdown(5 * time.Second)
}

func exitOnError(err error) {
	if err != nil {
		//nolint:sloglint
		slog.Error("fatal error", "err", err)
		os.Exit(1)
	}
}

// appVersion returns app version from VCS info.
func appVersion() string {
	result := "devel"
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return result
	}

	for _, v := range info.Settings {
		if v.Key == "vcs.revision" {
			result = v.Value
		}
	}

	if len(result) > 8 {
		result = result[:8]
	}

	return result
}
