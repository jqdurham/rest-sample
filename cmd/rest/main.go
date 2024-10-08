package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/jqdurham/rest-sample/internal/api"
	"github.com/jqdurham/rest-sample/internal/api/oapi"
	"github.com/jqdurham/rest-sample/internal/post"
	"github.com/jqdurham/rest-sample/internal/user"
	middleware "github.com/oapi-codegen/nethttp-middleware"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=../../.oapi-codegen.yaml ../../docs/openapi.json

func main() {
	defer func(start time.Time) {
		slog.Info("Application shutdown", "uptime", time.Since(start))
	}(time.Now())

	logr := slog.New(console.NewHandler(os.Stderr, &console.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logr)

	slog.Info("Application starting")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()

	var addr string
	flag.StringVar(&addr, "addr", ":8080", "Server listen address")

	userSvc := user.NewService()
	postSvc := post.NewService(userSvc)
	srvHandler := api.NewServerHandler(userSvc, postSvc)

	router := http.NewServeMux()
	oapi.HandlerFromMux(srvHandler, router)

	swagger, err := oapi.GetSwagger()
	if err != nil {
		fatal(err)
	}

	// https://github.com/oapi-codegen/oapi-codegen/issues/882
	swagger.Servers = nil

	h := middleware.OapiRequestValidator(swagger)(router)
	h = logRequestHandler(h)

	srv := &http.Server{
		Addr:    addr,
		Handler: h,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("API server starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("API server error", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
}

func logRequestHandler(h http.Handler) http.Handler {
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

		slog.Log(r.Context(), level,
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

func fatal(err error) {
	slog.Error(err.Error())
	os.Exit(1)
}
