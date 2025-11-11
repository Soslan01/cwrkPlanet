package domain

import "time"

type Room struct {
	ID              string    `db:"id"`
	Name            string    `db:"name"`
	MaxParticipants int64     `db:"max_participants"`
	CreatedAt       time.Time `db:"created_at"`
}
