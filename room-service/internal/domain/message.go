package domain

import "time"

type ChatMessage struct {
	ID        string    `db:"id"`
	RoomID    string    `db:"room_id"`
	UserID    int64     `db:"user_id"`
	Text      string    `db:"text"`
	ReplyTo   *string   `db:"reply_to"`
	CreatedAt time.Time `db:"created_at"`
}
