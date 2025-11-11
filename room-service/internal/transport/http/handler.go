package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/cwrk-planet/room-service/internal/domain"
	"github.com/cwrk-planet/room-service/internal/postgres"
	"github.com/cwrk-planet/room-service/internal/service"
	httpmw "github.com/cwrk-planet/room-service/internal/transport/http/middleware"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	roomSvc   *service.RoomService
	memberSvc *service.MemberService
	chatSvc   *service.ChatService
}

func NewHandler(room *service.RoomService, member *service.MemberService, chat *service.ChatService) *Handler {
	return &Handler{
		roomSvc:   room,
		memberSvc: member,
		chatSvc:   chat,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// POST /rooms
func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("handler.CreateRoom.Decode:", slog.Any("err", err))
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid json"})
		return
	}
	room, err := h.roomSvc.CreateRoom(r.Context(), req.Name, req.Max)
	if err != nil {
		slog.Error("handler.CreateRoom:", slog.Any("err", err))
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, RoomItem{
		ID:              room.ID,
		Name:            room.Name,
		MaxParticipants: room.MaxParticipants,
		CreatedAt:       room.CreatedAt,
	})
}

// GET /rooms?limit=&cursor=
func (h *Handler) ListRooms(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			limit = n
		}
	}
	cursor := r.URL.Query().Get("cursor")

	rooms, next, err := h.roomSvc.ListRooms(r.Context(), limit, cursor)
	if err != nil {
		if errors.Is(err, postgres.ErrInvalidCursor) {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_cursor"})
			return
		}
		slog.Error("handler.ListRooms:", slog.Any("err", err))
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	resp := RoomsListResponse{Items: make([]RoomItem, 0, len(rooms)), NextCursor: next}
	for _, rm := range rooms {
		resp.Items = append(resp.Items, RoomItem{
			ID:              rm.ID,
			Name:            rm.Name,
			MaxParticipants: rm.MaxParticipants,
			CreatedAt:       rm.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

// GET /rooms/{id}
func (h *Handler) GetRoom(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	room, err := h.roomSvc.GetRoom(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrRoomNotFound) {
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "room not found"})
			return
		}
		slog.Error("handler.GetRoom:", slog.Any("err", err))
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, RoomItem{
		ID:              room.ID,
		Name:            room.Name,
		MaxParticipants: room.MaxParticipants,
		CreatedAt:       room.CreatedAt,
	})
}

// POST /rooms/{id}/join
func (h *Handler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "id")
	userID := httpmw.UserIDFromCtx(r.Context())
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "missing user id"})
		return
	}

	p, err := h.memberSvc.JoinRoom(r.Context(), roomID, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrRoomNotFound):
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "room not found"})
			return
		case errors.Is(err, domain.ErrRoomFull):
			writeJSON(w, http.StatusConflict, ErrorResponse{Error: "room full"})
			return
		case errors.Is(err, domain.ErrAlreadyJoined):
			// участник уже в комнате
		default:
			slog.Error("handler.JoinRoom:", slog.Any("err", err))
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}
	}
	_ = p

	resp := JoinRoomResponse{
		RoomID: roomID,
		PeerID: strconv.FormatInt(userID, 10),
	}

	writeJSON(w, http.StatusOK, resp)
}

// POST /rooms/{id}/leave
func (h *Handler) LeaveRoom(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "id")
	userID := httpmw.UserIDFromCtx(r.Context())
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "missing user id"})
		return
	}

	err := h.memberSvc.LeaveRoom(r.Context(), roomID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotInRoom) {
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "user not in room"})
			return
		}
	}
	if err != nil && !errors.Is(err, domain.ErrNotInRoom) {
		slog.Error("handler.LeaveRoom:", slog.Any("err", err))
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "left"})
}

// GET /rooms/{id}/participants
func (h *Handler) GetParticipants(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "id")

	items, err := h.memberSvc.ListParticipantsDetailed(r.Context(), roomID)
	if err != nil {
		slog.Error("handler.ListParticipants:", slog.Any("err", err))
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	resp := ParticipantsResponse{Items: make([]ParticipantItem, 0, len(items))}
	for _, it := range items {
		resp.Items = append(resp.Items, ParticipantItem{
			UserID:      strconv.FormatInt(it.UserID, 10),
			DisplayName: it.DisplayName,
			AvatarURL:   it.AvatarURL,
			JoinedAt:    it.JoinedAt,
			LastSeen:    it.LastSeen,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

// GET /rooms/{id}/chat?after=&limit=
func (h *Handler) GetChatHistory(w http.ResponseWriter, r *http.Request) {
	if h.chatSvc == nil {
		writeJSON(w, http.StatusNotImplemented, ErrorResponse{Error: "chat service disabled"})
		return
	}
	roomID := chi.URLParam(r, "id")
	after := r.URL.Query().Get("after")
	limit := 50
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			limit = n
		}
	}
	items, next, err := h.chatSvc.History(r.Context(), roomID, after, limit)
	if err != nil {
		slog.Error("handler.GetChatHistory:", slog.Any("err", err))
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	resp := ChatHistoryResponse{Items: make([]ChatMessageItem, 0, len(items)), NextCursor: next}
	for _, m := range items {
		resp.Items = append(resp.Items, ChatMessageItem{
			ID:        m.ID,
			RoomID:    m.RoomID,
			UserID:    strconv.FormatInt(m.UserID, 10),
			Text:      m.Text,
			CreatedAt: m.CreatedAt.Truncate(time.Millisecond),
			ReplyTo:   m.ReplyTo,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}
