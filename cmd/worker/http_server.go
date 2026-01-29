package main

import (
	"context"
	"expvar"
	"log/slog"
	"net/http"
	"time"
)

func startWorkerHTTP(ctx context.Context, log *slog.Logger, addr string) *http.Server {
	mux := http.NewServeMux()

	// 基础健康
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// metrics：expvar
	mux.Handle("/debug/vars", expvar.Handler())
	mux.Handle("/metrics", expvar.Handler())

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		log.Info("worker_http_start", slog.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("worker_http_error", slog.Any("err", err))
		}
	}()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	return srv
}
