package domain

import "time"

type Participant struct {
	RoomID   string    `db:"room_id"`
	UserID   int64     `db:"user_id"`
	JoinedAt time.Time `db:"joined_at"`
	LastSeen time.Time `db:"last_seen"`
}
