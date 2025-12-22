package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/jqdurham/rest-sample/cmd/rest/handlers"
	"github.com/jqdurham/rest-sample/cmd/rest/handlers/oapi"
	"github.com/jqdurham/rest-sample/internal/post"
	"github.com/jqdurham/rest-sample/internal/user"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/phsym/console-slog"
	"golang.org/x/sync/errgroup"
)

//go:generate go tool oapi-codegen --config=../../.oapi-codegen.yaml ../../docs/openapi.json

func main() {
	log := slog.New(console.NewHandler(os.Stderr, &console.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(log)

	if err := run(log); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	log.Info("Application starting")
	defer func(start time.Time) {
		log.Info("Application shutdown", "uptime", time.Since(start))
	}(time.Now())

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()

	var addr string
	flag.StringVar(&addr, "addr", "127.0.0.1:8080", "Server listen address")
	flag.Parse()

	userSvc := user.NewService()
	postSvc := post.NewService(userSvc)
	srvHandler := handlers.NewServerHandler(userSvc, postSvc)

	router := http.NewServeMux()
	oapi.HandlerFromMux(srvHandler, router)

	swagger, err := oapi.GetSwagger()
	if err != nil {
		return fmt.Errorf("load swagger: %w", err)
	}

	// Disable Host header validation. See https://github.com/deepmap/oapi-codegen/issues/882
	swagger.Servers = nil

	svr := &http.Server{
		Addr:    addr,
		Handler: requestLoggerHandler(log,  middleware.OapiRequestValidator(swagger)(router)),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		ReadHeaderTimeout: 5 * time.Second,
	}

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		log.Info("API server starting", "addr", addr)
		defer log.Info("API server shutdown")
		if err := svr.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		<-ctx.Done()
		stopCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return svr.Shutdown(stopCtx)
	})

	return errors.Unwrap(eg.Wait())
}

func requestLoggerHandler(log *slog.Logger, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var (
			level   = slog.LevelInfo
			metrics = httpsnoop.CaptureMetrics(h, w, r)
		)
		switch metrics.Code {
		case http.StatusBadRequest:
			level = slog.LevelWarn
		case http.StatusInternalServerError:
			level = slog.LevelError
		}

		log.Log(r.Context(), level,
			"request handled",
			slog.String("method", r.Method),
			slog.String("url", r.URL.String()),
			slog.Int("status", metrics.Code),
			slog.Duration("dur", metrics.Duration),
			slog.Int64("bytes", metrics.Written),
			slog.String("ua", r.UserAgent()),
			slog.String("ip", r.RemoteAddr),
		)
	}
	return http.HandlerFunc(fn)
}
