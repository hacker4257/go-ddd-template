package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/hacker4257/go-ddd-template/internal/pkg/metrics"
)

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytes += n
	return n, err
}

func AccessLog(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			metrics.HTTPInFlight.Add(1)
			metrics.HTTPRequestsTotal.Add(1)
			start := time.Now()
			
			sw := &statusWriter{ResponseWriter: w}

			next.ServeHTTP(sw, r)

			metrics.HTTPInFlight.Add(-1)
    		metrics.IncStatus(sw.status)
    		metrics.ObserveHTTPLatency(time.Since(start))

			rid := GetRequestID(r.Context())
			log.Info("http_request",
				slog.String("request_id", rid),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", sw.status),
				slog.Int("bytes", sw.bytes),
				slog.Duration("latency", time.Since(start)),
				slog.String("remote", r.RemoteAddr),
			)
		})
	}
}
