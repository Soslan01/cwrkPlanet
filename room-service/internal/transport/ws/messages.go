package ws

// Типы событий, которые поступают в WS
const (
	TypeState      = "state"       // снапшот всех участников
	TypePeerJoined = "peer_joined" // пользователь присоединился
	TypePeerLeft   = "peer_left"   // пользователь покинул
	TypeChat       = "chat"        // чат-сообщение
	TypeChatAck    = "chat_ack"    // подтверждение отправки (НЕ сообщение)
)

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type StatePayload struct {
	RoomID       string                 `json:"room_id"`
	Participants []ParticipantStateItem `json:"participants"`
}

type ParticipantStateItem struct {
	UserID   string `json:"user_id"`
	JoinedAt int64  `json:"joined_at_unix"`
	LastSeen int64  `json:"last_seen_unix"`
}

type PeerEventPayload struct {
	RoomID string `json:"room_id"`
	UserID string `json:"user_id"`
}

type ChatPayload struct {
	RoomID  string `json:"room_id"`
	UserID  string `json:"user_id"`
	Message string `json:"message"`

	MsgID  string `json:"msg_id,omitempty"`
	TSUnix int64  `json:"ts_unix,omitempty"`
}

// для client: использует для снятия pending и дедупликации;
type ChatAckPayload struct {
	MsgID string `json:"msg_id"`
}
