package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/hacker4257/go-ddd-template/internal/api/http/handler"
	appmw "github.com/hacker4257/go-ddd-template/internal/api/http/middleware"
)

func NewRouter(log *slog.Logger, uh *handler.UserHandler) http.Handler {
	r := chi.NewRouter()

	// 基础稳定中间件
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)

	// 我们自己的
	r.Use(appmw.RequestID)
	r.Use(appmw.AccessLog(log))

	r.Get("/healthz", handler.Healthz)
	r.Get("/readyz", handler.Readyz)

	// 给个根路由，方便确认服务启动
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("go-ddd-template"))
	})

	r.Route("/users", func(r chi.Router) {
        r.Post("/", uh.Create)
        r.Get("/{id}", uh.Get)
    })

	return r
}
