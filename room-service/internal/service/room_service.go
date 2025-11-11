package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/cwrk-planet/room-service/internal/domain"
	"github.com/cwrk-planet/room-service/internal/postgres"
)

type RoomService struct {
	roomRepo *postgres.RoomRepository
}

func NewRoomService(roomRepo *postgres.RoomRepository) *RoomService {
	return &RoomService{roomRepo: roomRepo}
}

// CreateRoom создаёт комнату с заданным именем и лимитом участников.
func (s *RoomService) CreateRoom(ctx context.Context, name string, max int64) (*domain.Room, error) {
	if max <= 0 || max > 10 {
		max = 10
	}

	room := &domain.Room{
		Name:            name,
		MaxParticipants: max,
	}

	if err := s.roomRepo.Create(ctx, room); err != nil {
		return nil, fmt.Errorf("roomRepo.Create: %w", err)
	}
	return room, nil
}

// GetRoom возвращает комнату по ID.
func (s *RoomService) GetRoom(ctx context.Context, id string) (*domain.Room, error) {
	room, err := s.roomRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrRoomNotFound) {
			return nil, domain.ErrRoomNotFound
		}
		return nil, err
	}
	return room, nil
}

// ListRooms возвращает список комнат с курсорной пагинацией.
func (s *RoomService) ListRooms(ctx context.Context, limit int, cursor string) ([]domain.Room, string, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	rooms, nextCursor, err := s.roomRepo.List(ctx, limit, cursor)
	if err != nil {
		return nil, "", err
	}
	return rooms, nextCursor, nil
}

// DeleteRoom — опционально, если создатель покинул комнату
func (s *RoomService) DeleteRoom(ctx context.Context, id string) error {
	return s.roomRepo.Delete(ctx, id)
}
