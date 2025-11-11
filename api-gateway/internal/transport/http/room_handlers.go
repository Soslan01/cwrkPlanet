package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	approom "github.com/cwrk-planet/api-gateway/internal/app/room"
	"github.com/cwrk-planet/api-gateway/pkg/errs"
	"github.com/cwrk-planet/api-gateway/pkg/httputil"
)

type RoomHandlers struct {
	Room approom.Client
}

func bearer(r *http.Request) (string, bool) {
	authz := r.Header.Get("Authorization")
	parts := strings.SplitN(authz, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", false
	}

	return "Bearer " + strings.TrimSpace(parts[1]), true
}

func userID64(r *http.Request) (int64, bool) {
	s := strings.TrimSpace(r.Header.Get("X-User-ID"))
	if s == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false
	}

	return id, true
}

// POST /rooms
func (h *RoomHandlers) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var in approom.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "invalid JSON", nil)
		return
	}
	if strings.TrimSpace(in.Name) == "" {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "name is required", nil)
		return
	}
	// todo: перенести в конфиг
	if in.Max != 0 && (in.Max < 1 || in.Max > 10) {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "max must be in [1..10]", nil)
		return
	}
	auth, ok := bearer(r)
	if !ok {
		httputil.Error(r.Context(), w, http.StatusUnauthorized, "missing or invalid Authorization header", nil)
		return
	}
	uid, ok := userID64(r)
	if !ok {
		httputil.Error(r.Context(), w, http.StatusUnauthorized, "missing or invalid X-User-ID header", nil)
		return
	}

	out, err := h.Room.CreateRoom(r.Context(), auth, uid, in)
	if err != nil {
		status := errs.ToHTTP(err)
		httputil.Error(r.Context(), w, status, "create room failed", map[string]any{"reason": err.Error()})
		return
	}

	httputil.OK(w, out)
}

// GET /rooms?limit=&cursor=
func (h *RoomHandlers) ListRooms(w http.ResponseWriter, r *http.Request) {
	auth, ok := bearer(r)
	if !ok {
		httputil.Error(r.Context(), w, http.StatusUnauthorized, "missing or invalid Authorization header", nil)
		return
	}
	uid, ok := userID64(r)
	if !ok {
		httputil.Error(r.Context(), w, http.StatusUnauthorized, "missing or invalid X-User-ID header", nil)
		return
	}

	var limit int64
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 32); err == nil {
			limit = n
		}
	}
	cursor := r.URL.Query().Get("cursor")

	out, err := h.Room.ListRooms(r.Context(), auth, uid, limit, cursor)
	if err != nil {
		status := errs.ToHTTP(err)
		httputil.Error(r.Context(), w, status, "list rooms failed", map[string]any{"reason": err.Error()})
		return
	}

	httputil.OK(w, out)
}

// GET /rooms/{id}
func (h *RoomHandlers) GetRoom(w http.ResponseWriter, r *http.Request) {
	auth, ok := bearer(r)
	if !ok {
		httputil.Error(r.Context(), w, http.StatusUnauthorized, "missing or invalid Authorization header", nil)
		return
	}
	uid, ok := userID64(r)
	if !ok {
		httputil.Error(r.Context(), w, http.StatusUnauthorized, "missing or invalid X-User-ID header", nil)
		return
	}
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "id is required", nil)
		return
	}

	out, err := h.Room.GetRoom(r.Context(), auth, uid, id)
	if err != nil {
		status := errs.ToHTTP(err)
		httputil.Error(r.Context(), w, status, "get room failed", map[string]any{"reason": err.Error()})
		return
	}

	httputil.OK(w, out)
}

// POST /rooms/{id}/join
func (h *RoomHandlers) Join(w http.ResponseWriter, r *http.Request) {
	auth, ok := bearer(r)
	if !ok {
		httputil.Error(r.Context(), w, http.StatusUnauthorized, "missing or invalid Authorization header", nil)
		return
	}
	uid, ok := userID64(r)
	if !ok {
		httputil.Error(r.Context(), w, http.StatusUnauthorized, "missing or invalid X-User-ID header", nil)
		return
	}
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "id is required", nil)
		return
	}

	out, err := h.Room.Join(r.Context(), auth, uid, id)
	if err != nil {
		status := errs.ToHTTP(err)
		httputil.Error(r.Context(), w, status, "join failed", map[string]any{"reason": err.Error()})
		return
	}

	httputil.OK(w, out)
}

// POST /rooms/{id}/leave
func (h *RoomHandlers) Leave(w http.ResponseWriter, r *http.Request) {
	auth, ok := bearer(r)
	if !ok {
		httputil.Error(r.Context(), w, http.StatusUnauthorized, "missing or invalid Authorization header", nil)
		return
	}
	uid, ok := userID64(r)
	if !ok {
		httputil.Error(r.Context(), w, http.StatusUnauthorized, "missing or invalid X-User-ID header", nil)
		return
	}
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "id is required", nil)
		return
	}

	if err := h.Room.Leave(r.Context(), auth, uid, id); err != nil {
		status := errs.ToHTTP(err)
		httputil.Error(r.Context(), w, status, "leave failed", map[string]any{"reason": err.Error()})
		return
	}

	httputil.OK(w, map[string]string{"status": "left"})
}

// GET /rooms/{id}/participants
func (h *RoomHandlers) Participants(w http.ResponseWriter, r *http.Request) {
	auth, ok := bearer(r)
	if !ok {
		httputil.Error(r.Context(), w, http.StatusUnauthorized, "missing or invalid Authorization header", nil)
		return
	}
	uid, ok := userID64(r)
	if !ok {
		httputil.Error(r.Context(), w, http.StatusUnauthorized, "missing or invalid X-User-ID header", nil)
		return
	}
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "id is required", nil)
		return
	}

	out, err := h.Room.Participants(r.Context(), auth, uid, id)
	if err != nil {
		status := errs.ToHTTP(err)
		httputil.Error(r.Context(), w, status, "participants failed", map[string]any{"reason": err.Error()})
		return
	}

	httputil.OK(w, out)
}

// GET /rooms/{id}/chat?after=&limit=
func (h *RoomHandlers) ChatHistory(w http.ResponseWriter, r *http.Request) {
	auth, ok := bearer(r)
	if !ok {
		httputil.Error(r.Context(), w, http.StatusUnauthorized, "missing or invalid Authorization header", nil)
		return
	}
	uid, ok := userID64(r)
	if !ok {
		httputil.Error(r.Context(), w, http.StatusUnauthorized, "missing or invalid X-User-ID header", nil)
		return
	}
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "id is required", nil)
		return
	}
	after := r.URL.Query().Get("after")
	var limit int32 = 50
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 32); err == nil && n > 0 {
			limit = int32(n)
		}
	}

	out, err := h.Room.ChatHistory(r.Context(), auth, uid, id, after, limit)
	if err != nil {
		status := errs.ToHTTP(err)
		httputil.Error(r.Context(), w, status, "chat history failed", map[string]any{"reason": err.Error()})
		return
	}

	httputil.OK(w, out)
}
