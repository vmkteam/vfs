package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/vmkteam/vfs"
	"github.com/vmkteam/vfs/db"

	"github.com/go-pg/pg/v10"
	monitor "github.com/hypnoglow/go-pg-monitor"
	"github.com/labstack/echo/v4"
	"github.com/vmkteam/appkit"
	"github.com/vmkteam/embedlog"
	"github.com/vmkteam/rpcgen/v2"
	"github.com/vmkteam/rpcgen/v2/golang"
	zm "github.com/vmkteam/zenrpc-middleware"
	"github.com/vmkteam/zenrpc/v2"
)

type ServerConfig struct {
	Host      string
	Port      int
	IsDevel   bool
	JWTHeader string
	JWTKey    string

	// Index indexes files on start: width, height, blurhash.
	Index bool

	// IndexBlurhash calculates blurhash (could be long operation on large files)
	IndexBlurhash bool

	// IndexWorkers is total running indexer workers, default is cores
	IndexWorkers int

	// IndexBatchSize is Indexer batch size for files, default is 64
	IndexBatchSize uint64
}

type Config struct {
	Server   ServerConfig
	Database *pg.Options
	VFS      vfs.Config
}

type App struct {
	embedlog.Logger
	appName string
	cfg     Config
	db      db.DB
	dbc     *pg.DB
	repo    *db.VfsRepo
	vfs     vfs.VFS
	mon     *monitor.Monitor
	echo    *echo.Echo
}

func New(appName string, sl embedlog.Logger, cfg Config, dbc *pg.DB) (*App, error) {
	a := &App{
		Logger:  sl,
		appName: appName,
		cfg:     cfg,
		db:      db.New(dbc),
		dbc:     dbc,
		echo:    appkit.NewEcho(),
	}

	// init vfs
	if v, err := vfs.New(cfg.VFS, sl); err != nil {
		return nil, err
	} else {
		a.vfs = v
	}

	// set repo if db conn
	if a.dbc != nil {
		repo := db.NewVfsRepo(a.db)
		a.repo = &repo
	}

	// add services

	return a, nil
}

// Run is a function that runs application.
func (a *App) Run(ctx context.Context) error {
	a.registerMiddlewares()
	a.registerMetrics()
	a.registerDebugHandlers()
	a.registerHandlers()
	a.registerMetadata()
	if a.dbc != nil {
		a.registerAPIHandlers()
	}

	return a.runHTTPServer(ctx, a.cfg.Server.Host, a.cfg.Server.Port)
}

func (a *App) registerAPIHandlers() {
	var allowDebugFn = func() zm.AllowDebugFunc {
		return func(req *http.Request) bool {
			return req != nil && req.FormValue("__level") == "5"
		}
	}

	repo := db.NewVfsRepo(a.db)
	srv := zenrpc.NewServer(zenrpc.Options{ExposeSMD: true, AllowCORS: true})
	srv.Use(
		zm.WithHeaders(),
		zm.WithDevel(a.cfg.Server.IsDevel),
		zm.WithNoCancelContext(),
		zm.WithMetrics(zm.DefaultServerName),
		zm.WithSLog(a.Print, zm.DefaultServerName, nil),
		zm.WithTiming(a.cfg.Server.IsDevel, allowDebugFn()),
		zm.WithSentry(zm.DefaultServerName),
	)

	srv.Register("vfs", vfs.NewService(repo, a.vfs, a.dbc))

	gen := rpcgen.FromSMD(srv.SMD())

	a.echo.Any("/rpc/", echo.WrapHandler(a.authMiddleware(appkit.XRequestID(srv))))
	a.echo.GET("/rpc/doc/", appkit.EchoHandlerFunc(zenrpc.SMDBoxHandler))
	a.echo.GET("/rpc/openrpc.json", appkit.EchoHandlerFunc(rpcgen.Handler(gen.OpenRPC(a.appName, "http://localhost:8075/rpc"))))
	a.echo.GET("/rpc/api.ts", appkit.EchoHandlerFunc(rpcgen.Handler(gen.TSClient(nil))))
	a.echo.GET("/rpc/api.go", appkit.EchoHandlerFunc(rpcgen.Handler(gen.GoClient(golang.Settings{Package: a.appName}))))

	a.echo.Any("/upload/file", echo.WrapHandler(a.authMiddleware(a.vfs.UploadHandler(repo))))
}

// registerHandlers registers base vfs handlers
func (a *App) registerHandlers() {
	// enable base handlers
	a.echo.Any("/auth-token", a.issueTokenHandler)
	a.echo.Any("/upload/hash", echo.WrapHandler(a.authMiddleware(a.vfs.HashUploadHandler(a.repo))))
	a.echo.Static(a.cfg.VFS.WebPath, a.cfg.VFS.Path)

	// enabled indexer
	sc := a.cfg.Server
	if sc.Index {
		hi := vfs.NewHashIndexer(a.Logger, a.db, a.repo, a.vfs, sc.IndexWorkers, sc.IndexBatchSize, sc.IndexBlurhash)

		a.echo.Any("/scan-files", hi.ScanFilesHandler)
		a.echo.GET("/preview/:ns/:file", hi.Preview)
		a.echo.GET("/preview/:file", hi.Preview)

		go hi.Start()
		defer hi.Stop()
	}
}

// registerMetadata is a function that registers meta info from service. Must be updated.
func (a *App) registerMetadata() {
	opts := appkit.MetadataOpts{
		HasPublicAPI:  true,
		HasPrivateAPI: true,
		HasCronJobs:   a.cfg.Server.Index,
	}

	if a.dbc != nil {
		opts.DBs = []appkit.DBMetadata{
			appkit.NewDBMetadata(a.cfg.Database.Database, a.cfg.Database.PoolSize, false),
		}
	}

	md := appkit.NewMetadataManager(opts)
	md.RegisterMetrics()

	a.echo.GET("/debug/metadata", md.Handler)
}

// runHTTPServer is a function that starts http listener using labstack/echo.
func (a *App) runHTTPServer(ctx context.Context, host string, port int) error {
	listenAddress := fmt.Sprintf("%s:%d", host, port)
	a.Print(ctx, "starting http listener", "url", "http://"+listenAddress)

	return a.echo.Start(listenAddress)
}

// Shutdown is a function that gracefully stops HTTP server.
func (a *App) Shutdown(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if a.mon != nil {
		a.mon.Close()
	}

	if err := a.echo.Shutdown(ctx); err != nil {
		a.Error(ctx, "shutting down server", "err", err)
	}
}
