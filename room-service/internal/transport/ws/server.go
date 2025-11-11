package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cwrk-planet/room-service/internal/domain"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

type MemberSvc interface {
	ListParticipants(ctx context.Context, roomID string) ([]domain.Participant, error)
	TouchHeartbeat(ctx context.Context, roomID string, userID int64) error
	LeaveRoom(ctx context.Context, roomID string, userID int64) error
}

type ChatSvc interface {
	Save(ctx context.Context, roomID string, userID int64, text string) (msgID string, createdAt time.Time, err error)
}

type Server struct {
	upgrader  websocket.Upgrader
	hub       *Hub
	memberSvc MemberSvc
	chatSvc   ChatSvc

	pingEvery time.Duration
}

func NewServer(hub *Hub, member MemberSvc, chat ChatSvc) *Server {
	return &Server{
		hub:       hub,
		memberSvc: member,
		chatSvc:   chat,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		pingEvery: 15 * time.Second,
	}
}

// WS endpoint: GET /ws/rooms/{id}?access_token=...&user_id=...
func (s *Server) HandleWS(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	accessToken := strings.TrimSpace(q.Get("access_token"))
	userIDStr := strings.TrimSpace(q.Get("user_id"))
	if accessToken == "" {
		http.Error(w, "missing access_token", http.StatusUnauthorized)
		return
	}
	uid, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || uid <= 0 {
		http.Error(w, "invalid user_id", http.StatusUnauthorized)
		return
	}
	roomID := chi.URLParam(r, "id")
	if roomID == "" {
		http.Error(w, "missing room id", http.StatusBadRequest)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("ws upgrade failed", "err", err)
		http.Error(w, "upgrade failed", http.StatusBadRequest)
		return
	}

	c := newWsConn(conn, roomID, uid)
	s.hub.Add(c)

	if err := s.sendState(r.Context(), c); err != nil {
		slog.Warn("ws send initial state failed", "room", roomID, "user", uid, "err", err)
	}

	// peer_joined
	s.hub.Broadcast(roomID, Message{
		Type: TypePeerJoined,
		Payload: PeerEventPayload{
			RoomID: roomID,
			UserID: userIDStr,
		},
	})

	go s.writeLoop(r.Context(), c)
	s.readLoop(r.Context(), c)

	s.hub.Remove(c)

	if err := s.memberSvc.LeaveRoom(r.Context(), roomID, uid); err != nil {
		slog.Debug("ws leave room failed", "room", roomID, "user", uid, "err", err)
	}
	s.hub.Broadcast(roomID, Message{
		Type: TypePeerLeft,
		Payload: PeerEventPayload{
			RoomID: roomID,
			UserID: userIDStr,
		},
	})

	if err := c.Close(); err != nil {
		slog.Debug("ws close failed", "room", roomID, "user", uid, "err", err)
	}
}

func (s *Server) sendState(ctx context.Context, c *wsConn) error {
	parts, err := s.memberSvc.ListParticipants(ctx, c.roomID)
	if err != nil {
		return err
	}
	items := make([]ParticipantStateItem, 0, len(parts))
	for _, p := range parts {
		idStr := strconv.FormatInt(p.UserID, 10)
		items = append(items, ParticipantStateItem{
			UserID:   idStr,
			JoinedAt: p.JoinedAt.Unix(),
			LastSeen: p.LastSeen.Unix(),
		})
	}

	return c.Send(Message{
		Type: TypeState,
		Payload: StatePayload{
			RoomID:       c.roomID,
			Participants: items,
		},
	})
}

func (s *Server) readLoop(ctx context.Context, c *wsConn) {
	defer func() { _ = c.Close() }()
	idStr := strconv.FormatInt(c.userID, 10)

	_ = s.memberSvc.TouchHeartbeat(ctx, c.roomID, c.userID)

	c.conn.SetReadLimit(1 << 20)
	c.conn.SetReadDeadline(time.Now().Add(2 * s.pingEvery))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(2 * s.pingEvery))
		_ = s.memberSvc.TouchHeartbeat(ctx, c.roomID, c.userID)
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case TypeChat:
			var p ChatPayload
			if decode(msg.Payload, &p) == nil {
				p.RoomID = c.roomID
				p.UserID = idStr
				text := strings.TrimSpace(p.Message)
				if text == "" {
					continue
				}

				var (
					msgID string
					ts    time.Time
				)
				if s.chatSvc != nil {
					if id, createdAt, err := s.chatSvc.Save(ctx, c.roomID, c.userID, text); err == nil {
						msgID, ts = id, createdAt
					} else {
						slog.Warn("ws chat save failed", "room", c.roomID, "user", c.userID, "err", err)
						ts = time.Now()
					}
				} else {
					// если chatSvc == nil, всё равно рассылаем (без id), но добавим ts=now для фронта
					ts = time.Now()
				}

				// ЕДИНЫЙ broadcast всем (включая отправителя). Никаких c.Send(...) такого же TypeChat.
				out := ChatPayload{
					RoomID:  c.roomID,
					UserID:  idStr,
					Message: text,
				}
				if msgID != "" {
					out.MsgID = msgID
				}
				if !ts.IsZero() {
					out.TSUnix = ts.Unix()
				}
				s.hub.Broadcast(c.roomID, Message{Type: TypeChat, Payload: out})

				// Лёгкий ACK только отправителю, чтобы снять pending на клиенте.
				if msgID != "" {
					_ = c.Send(Message{
						Type:    TypeChatAck,
						Payload: ChatAckPayload{MsgID: msgID},
					})
				}
			}
		default:
			// ignore
		}
	}
}

func (s *Server) writeLoop(ctx context.Context, c *wsConn) {
	ticker := time.NewTicker(s.pingEvery)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_ = c.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second))
			_ = s.memberSvc.TouchHeartbeat(ctx, c.roomID, c.userID)
		case <-ctx.Done():
			return
		case <-c.closed:
			return
		}
	}
}

// --- helpers ---

func decode(payload interface{}, dst interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, dst)
}

type wsConn struct {
	conn   *websocket.Conn
	roomID string
	userID int64
	sendMu chan struct{}
	closed chan struct{}
}

func newWsConn(c *websocket.Conn, roomID string, userID int64) *wsConn {
	return &wsConn{
		conn:   c,
		roomID: roomID,
		userID: userID,
		sendMu: make(chan struct{}, 1),
		closed: make(chan struct{}),
	}
}

func (c *wsConn) Send(msg Message) error {
	c.sendMu <- struct{}{}
	defer func() { <-c.sendMu }()
	c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	return c.conn.WriteJSON(msg)
}

func (c *wsConn) Close() error {
	select {
	case <-c.closed:
	default:
		close(c.closed)
	}

	return c.conn.Close()
}

func (c *wsConn) UserID() string { return strconv.FormatInt(c.userID, 10) }
func (c *wsConn) RoomID() string { return c.roomID }
