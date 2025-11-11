package domain

import "errors"

var (
	ErrRoomNotFound  = errors.New("room not found")
	ErrRoomFull      = errors.New("room is full")
	ErrAlreadyJoined = errors.New("user already joined the room")
	ErrNotInRoom     = errors.New("user not in the room")
)
