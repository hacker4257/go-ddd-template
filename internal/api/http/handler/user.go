package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	userapp "github.com/hacker4257/go-ddd-template/internal/app/user"
	"github.com/hacker4257/go-ddd-template/internal/domain/user"
)

type UserHandler struct {
	svc *userapp.Service
}

func NewUserHandler(svc *userapp.Service) *UserHandler {
	return &UserHandler{svc: svc}
}

type createUserReq struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type userResp struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	u, err := h.svc.Create(r.Context(), userapp.CreateUserCmd{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		writeUserErr(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toUserResp(u))
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	u, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeUserErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResp(u))
}

func toUserResp(u user.User) userResp {
	return userResp{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
	}
}

func writeUserErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, user.ErrInvalidInput):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, user.ErrEmailExists):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, user.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
