package app

import (
	"context"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/vmkteam/appkit"
)

func (a *App) registerMiddlewares() {
	headers := []string{"Authorization", "Authorization2", "Origin", "X-Requested-With", "Content-Type", "Accept", "Platform", "Version", "X-Request-ID"}
	if a.cfg.Server.JWTHeader != "" {
		headers = append(headers, a.cfg.Server.JWTHeader)
	}

	a.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
		AllowHeaders: headers,
	}))

	a.echo.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:    true,
		LogURI:       true,
		LogError:     true,
		HandleError:  true,
		LogLatency:   true,
		LogRemoteIP:  true,
		LogRequestID: true,
		LogUserAgent: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			attrs := []slog.Attr{
				slog.String("ip", v.RemoteIP),
				slog.String("uri", v.URI),
				slog.Int("status", v.Status),
				slog.String("userAgent", v.UserAgent),
				slog.String("duration", v.Latency.String()),
				slog.String("xRequestId", v.RequestID),
			}

			if v.Error == nil {
				a.Log().LogAttrs(context.Background(), slog.LevelInfo, "http request", attrs...)
			} else {
				a.Log().LogAttrs(context.Background(), slog.LevelError, "http request error", append(attrs, slog.String("err", v.Error.Error()))...)
			}
			return nil
		},
	}))
}

// registerDebugHandlers adds /debug/pprof handlers into a.echo instance.
func (a *App) registerDebugHandlers() {
	dbg := a.echo.Group("/debug")
	dbg.Any("/pprof/*", appkit.PprofHandler)

	// add healthcheck
	a.echo.GET("/status", func(c echo.Context) error {
		// test postgresql connection
		err := a.db.Ping(c.Request().Context())
		if err != nil {
			a.Error(c.Request().Context(), "failed to check db connection", "err", err)
			return c.String(http.StatusInternalServerError, "DB error")
		}
		return c.String(http.StatusOK, "OK")
	})

	if a.cfg.Server.IsDevel {
		a.echo.GET("/", appkit.RenderRoutes(a.appName, a.echo))
	}
}

// issueTokenHandler issues new jwt token for 1 hour. Subject can be set by id GET/POST param
func (a *App) issueTokenHandler(c echo.Context) (err error) {
	id := c.FormValue("id")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "vfs",
		Subject:   id,
	})

	key := []byte(a.cfg.Server.JWTKey)
	tokenString, err := token.SignedString(key)
	sl := a.With("id", id)
	sl.PrintOrErr(c.Request().Context(), "issued new token", err, "token", tokenString)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, tokenString)
}

// authMiddleware checks JWT token if set in flag jwt.header.
func (a *App) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			isOK    = true
			errMsg  = ""
			errCode = http.StatusUnauthorized
		)

		defer func() {
			if isOK {
				next.ServeHTTP(w, r)
			} else {
				http.Error(w, errMsg, errCode)
			}
		}()

		if a.cfg.Server.JWTHeader != "" {
			isOK = false
			tokenString := r.Header.Get(a.cfg.Server.JWTHeader)
			if tokenString == "" {
				errMsg = "missing token"
				return
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(a.cfg.Server.JWTKey), nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

			switch {
			case err != nil:
				errMsg, errCode = err.Error(), http.StatusForbidden
			case !token.Valid:
				errMsg = "bad token"
			default:
				isOK = token.Valid
			}
		}
	})
}
