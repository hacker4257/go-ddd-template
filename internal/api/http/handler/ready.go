package handler

import (
	"net/http"

	"github.com/hacker4257/go-ddd-template/internal/pkg/health"
)

type ReadyHandler struct {
	Checker health.Checker
}

func (h ReadyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.Checker.Ready(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready"))
}
