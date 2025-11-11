package room

import "time"

type CreateRoomRequest struct {
	Name string `json:"name"`
	Max  int64  `json:"max,omitempty"`
}

type RoomItem struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	MaxParticipants int64     `json:"max_participants"`
	CreatedAt       time.Time `json:"created_at"`
}

type RoomsListResponse struct {
	Items      []RoomItem `json:"items"`
	NextCursor string     `json:"next_cursor,omitempty"`
}

type JoinRoomResponse struct {
	RoomID string `json:"room_id"`
	PeerID string `json:"peer_id"`
}

type ParticipantItem struct {
	UserID   string    `json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`
	LastSeen time.Time `json:"last_seen"`
}

type ParticipantsResponse struct {
	Items []ParticipantItem `json:"items"`
}

type ChatMessageItem struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"room_id"`
	UserID    string    `json:"user_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	ReplyTo   string    `json:"reply_to,omitempty"`
}

type ChatHistoryResponse struct {
	Items      []ChatMessageItem `json:"items"`
	NextCursor string            `json:"next_cursor,omitempty"`
}
