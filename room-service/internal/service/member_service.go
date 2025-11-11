package service

import (
	"context"
	"time"

	"github.com/cwrk-planet/room-service/internal/domain"
	"github.com/cwrk-planet/room-service/internal/postgres"
)

type MemberService struct {
	roomRepo        *postgres.RoomRepository
	participantRepo *postgres.ParticipantRepository

	heartbeatWindow time.Duration
}

func NewMemberService(roomRepo *postgres.RoomRepository, participantRepo *postgres.ParticipantRepository) *MemberService {
	return &MemberService{
		roomRepo:        roomRepo,
		participantRepo: participantRepo,
		heartbeatWindow: 60 * time.Second, // окно «онлайн»
	}
}

func (s *MemberService) SetHeartbeatWindow(d time.Duration) {
	if d > 0 {
		s.heartbeatWindow = d
	}
}

func (s *MemberService) JoinRoom(ctx context.Context, roomID string, userID int64) (*domain.Participant, error) {
	room, err := s.roomRepo.Get(ctx, roomID)
	if err != nil {
		return nil, err
	}

	exists, err := s.participantRepo.Exists(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrAlreadyJoined
	}

	p := &domain.Participant{
		RoomID:   roomID,
		UserID:   userID,
		JoinedAt: time.Now(),
		LastSeen: time.Now(),
	}

	if err := s.participantRepo.Join(ctx, p, room.MaxParticipants); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *MemberService) LeaveRoom(ctx context.Context, roomID string, userID int64) error {
	return s.participantRepo.Leave(ctx, roomID, userID)
}

func (s *MemberService) ListParticipants(ctx context.Context, roomID string) ([]domain.Participant, error) {
	return s.participantRepo.ListByRoom(ctx, roomID)
}

func (s *MemberService) TouchHeartbeat(ctx context.Context, roomID string, userID int64) error {
	return s.participantRepo.TouchHeartbeat(ctx, roomID, userID)
}

type ParticipantDetailed struct {
	UserID      int64
	DisplayName *string
	AvatarURL   *string
	JoinedAt    time.Time
	LastSeen    time.Time
}

func (s *MemberService) ListParticipantsDetailed(ctx context.Context, roomID string) ([]ParticipantDetailed, error) {
	rows, err := s.participantRepo.ListDetailed(ctx, roomID, s.heartbeatWindow)
	if err != nil {
		return nil, err
	}
	out := make([]ParticipantDetailed, 0, len(rows))
	for _, r := range rows {
		out = append(out, ParticipantDetailed{
			UserID:      r.UserID,
			DisplayName: r.DisplayName,
			AvatarURL:   r.AvatarURL,
			JoinedAt:    r.JoinedAt,
			LastSeen:    r.LastSeen,
		})
	}

	return out, nil
}
