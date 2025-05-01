package app

import (
	"fmt"
	"strconv"
	"time"

	monitor "github.com/hypnoglow/go-pg-monitor"
	"github.com/hypnoglow/go-pg-monitor/gopgv10"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	zm "github.com/vmkteam/zenrpc-middleware"
)

// registerMetrics is a function that initializes a.stat* variables and adds /metrics endpoint to echo.
func (a *App) registerMetrics() {
	if a.dbc != nil {
		// add db conn metrics
		dbMetrics := monitor.NewMetrics(monitor.MetricsWithConstLabels(prometheus.Labels{"connection_name": "default"}))
		dbOpts := a.db.Options()
		a.mon = monitor.NewMonitor(
			gopgv10.NewObserver(a.db.DB),
			dbMetrics,
			monitor.MonitorWithPoolName(fmt.Sprintf("%s/%s", dbOpts.Addr, dbOpts.Database)),
		)
		a.mon.Open()
	}

	a.echo.Use(httpMetrics(zm.DefaultServerName))
	a.echo.Any("/metrics", echo.WrapHandler(promhttp.Handler()))
}

// httpMetrics is the middleware function that logs duration of responses.
func httpMetrics(appName string) echo.MiddlewareFunc {
	labels := []string{"method", "uri", "code"}

	if appName == "" {
		appName = "app"
	}

	echoRequests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: appName,
		Subsystem: "http",
		Name:      "requests_count",
		Help:      "Requests count by method/path/status.",
	}, labels)

	echoDurations := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: appName,
		Subsystem: "http",
		Name:      "responses_duration_seconds",
		Help:      "Response time by method/path/status.",
	}, labels)

	prometheus.MustRegister(echoRequests, echoDurations)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			if err := next(c); err != nil {
				c.Error(err)
			}

			metrics := []string{c.Request().Method, c.Path(), strconv.Itoa(c.Response().Status)}

			echoDurations.WithLabelValues(metrics...).Observe(time.Since(start).Seconds())
			echoRequests.WithLabelValues(metrics...).Inc()

			return nil
		}
	}
}
