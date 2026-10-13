// Command api is the composition root of the HTTP API (ADR-0001): it loads
// the configuration, wires adapters into the generated strict server, builds
// the middleware chain and serves until SIGINT/SIGTERM, then drains in-flight
// requests.
//
// Middleware order (docs/architecture/c4.md): request id → recover → rate
// limit → routing; then, per operation, flag gate → authentication →
// contract validation → handler (adapter/http/router).
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/adapter/http/health"
	"github.com/romodevx/control-contable/backend/adapter/http/middleware"
	"github.com/romodevx/control-contable/backend/adapter/http/router"
	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

// version is reported by /healthz; release builds set it with
// -ldflags "-X main.version=<semver>".
var version = "0.0.0-dev"

// http.Server limits. WriteTimeout bounds the slowest handler; shutdown
// fits inside Cloud Run's 10 s grace period after SIGTERM.
const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 120 * time.Second
	maxHeaderBytes    = 64 << 10
	shutdownTimeout   = 8 * time.Second
)

func main() {
	if err := runUntilSignal(); err != nil {
		fmt.Fprintln(os.Stderr, "api:", err)
		os.Exit(1)
	}
}

func runUntilSignal() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return run(ctx, os.LookupEnv, os.Stdout)
}

// run loads the configuration and serves until ctx is done.
func run(ctx context.Context, lookup config.Lookup, logOut io.Writer) error {
	cfg, err := config.Load(ctx, lookup, config.NewEnvSecrets(lookup))
	if err != nil {
		return err
	}
	logger := logging.New(logOut, cfg.LogLevel)
	handler, err := newHandler(cfg, deps{
		logger: logger,
		db:     health.NotWired{}, // the database adapter arrives with P0-12
		now:    time.Now,
	})
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", cfg.HTTPAddr, err)
	}
	logger.InfoContext(ctx, "api listening", slog.String("addr", ln.Addr().String()),
		slog.String("env", string(cfg.Env)), slog.String("version", version))
	return serve(ctx, newServer(handler, cfg.HTTPAddr, logger), ln, logger)
}

// deps are the adapters newHandler wires; tests substitute them.
type deps struct {
	logger *slog.Logger
	db     health.Pinger
	now    func() time.Time
}

// newHandler builds the full handler: the edge middlewares around the
// generated server with its per-operation pipeline.
func newHandler(cfg config.Config, d deps) (http.Handler, error) {
	spec, err := gen.GetSpec()
	if err != nil {
		return nil, fmt.Errorf("load embedded contract: %w", err)
	}
	server := api{Handler: health.New(d.db, version, d.logger)}
	routes, err := router.New(server, spec, router.Options{
		Flags:  router.AllModulesOff{},    // the flags service arrives with P0-15
		Authn:  router.NoAuthentication{}, // JWT authentication arrives with M1-07
		Logger: d.logger,
	})
	if err != nil {
		return nil, err
	}
	limiter := ratelimit.New(cfg.RateLimit.PerSecond, cfg.RateLimit.Burst, d.now)
	return middleware.Chain(routes,
		middleware.RequestID,
		middleware.Recover(d.logger),
		middleware.RateLimit(limiter, middleware.ClientIP(cfg.TrustedProxyHops)),
	), nil
}

func newServer(handler http.Handler, addr string, logger *slog.Logger) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelWarn),
	}
}

// serve runs srv on ln until ctx is done, then stops accepting connections
// and waits up to shutdownTimeout for in-flight requests.
func serve(ctx context.Context, srv *http.Server, ln net.Listener, logger *slog.Logger) error {
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()

	select {
	case err := <-errCh:
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
	}
	logger.Info("api shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-errCh; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve: %w", err)
	}
	return nil
}
