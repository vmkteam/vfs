package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/vmkteam/vfs"
	"github.com/vmkteam/vfs/db"
	"github.com/vmkteam/vfs/internal/app"

	"github.com/BurntSushi/toml"
	"github.com/go-pg/pg/v10"
	"github.com/namsral/flag"
	"github.com/vmkteam/appkit"
	"github.com/vmkteam/embedlog"
)

const (
	appName = "vfssrv"
)

var (
	fs           = flag.NewFlagSetWithEnvPrefix(os.Args[0], "VFS", 0)
	flConfigPath = fs.String("config", "config.toml", "path to config file")
	flInitConfig = fs.Bool("init", false, "write default config file")
	flVerbose    = fs.Bool("verbose", false, "enable debug output")
	flJSONLogs   = fs.Bool("json", false, "enable json output")
	flDev        = fs.Bool("dev", false, "enable dev mode")
	cfg          app.Config
)

func main() {
	flag.DefaultConfigFlagname = "config.flag"
	err := fs.Parse(os.Args[1:])
	exitOnError(err)

	// setup logger
	sl, ctx := embedlog.NewLogger(*flVerbose, *flJSONLogs), context.Background()
	if *flDev {
		sl = embedlog.NewDevLogger()
	}
	slog.SetDefault(sl.Log()) // set default logger
	ql := db.NewQueryLogger(sl)
	pg.SetLogger(ql)

	// check for default config
	if *flInitConfig && *flConfigPath != "" {
		exitOnError(writeConfig(*flConfigPath))
		sl.Print(ctx, "config file successfully written", "file", *flConfigPath)
		return
	}

	_, err = toml.DecodeFile(*flConfigPath, &cfg)
	exitOnError(err)

	// connect to DB
	var dbc *pg.DB
	if cfg.Database != nil {
		dbc = pg.Connect(cfg.Database)
		exitOnError(dbc.Ping(ctx))
		if *flDev {
			dbc.AddQueryHook(ql)
		}
	}

	// create app
	a, err := app.New(appName, sl, cfg, dbc)
	exitOnError(err)

	sl.Print(ctx, "starting", "app", appName, "version", appkit.Version(), "host", cfg.Server.Host, "port", cfg.Server.Port, "jwtHeader", cfg.Server.JWTHeader)
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

// writeConfig writes default config file to `-config` path.
func writeConfig(configPath string) error {
	var defaultConfig = app.Config{
		Server: app.ServerConfig{
			Host:           "0.0.0.0",
			Port:           9999,
			IsDevel:        false,
			JWTHeader:      "AuthorizationJWT",
			JWTKey:         randomString(16),
			Index:          false,
			IndexBlurhash:  true,
			IndexWorkers:   runtime.NumCPU() / 2,
			IndexBatchSize: 64,
		},
		Database: nil,
		VFS: vfs.Config{
			MaxFileSize:      32 << 20,
			Path:             "testdata",
			WebPath:          "/media/",
			PreviewPath:      "/media/small/",
			Database:         nil,
			Namespaces:       []string{"items", "test"},
			Extensions:       []string{"jpg", "jpeg", "png", "gif"},
			MimeTypes:        []string{"image/jpeg", "image/png", "image/gif"},
			UploadFormName:   "Filedata",
			SaltedFilenames:  false,
			SkipFolderVerify: false,
		},
	}

	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(defaultConfig); err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	// write default DB config
	buf.WriteString(`
[Database]
  Addr     = "localhost:5432"
  User     = "postgres"
  Password = ""
  Database = "apisrv"
  PoolSize = 10
  ApplicationName = "vfssrv"`)

	if err := os.WriteFile(configPath, buf.Bytes(), os.ModePerm); err != nil {
		return fmt.Errorf("failed to write file %s: %w", configPath, err)
	}

	return nil
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
