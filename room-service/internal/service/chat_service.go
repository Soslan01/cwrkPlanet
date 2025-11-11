package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/cwrk-planet/room-service/internal/domain"
	"github.com/cwrk-planet/room-service/internal/postgres"
)

type ChatService struct {
	chatRepo *postgres.ChatRepository
}

func NewChatService(chatRepo *postgres.ChatRepository) *ChatService {
	return &ChatService{chatRepo: chatRepo}
}

func (s *ChatService) Save(ctx context.Context, roomID string, userID int64, text string) (string, time.Time, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", time.Time{}, errors.New("empty message")
	}
	// todo: вынести в конфиг
	if len(text) > 4000 {
		return "", time.Time{}, errors.New("message too long")
	}
	msg, err := s.chatRepo.Save(ctx, roomID, userID, text, nil)
	if err != nil {
		return "", time.Time{}, err
	}
	return msg.ID, msg.CreatedAt, nil
}

func (s *ChatService) History(ctx context.Context, roomID, after string, limit int) ([]domain.ChatMessage, string, error) {
	return s.chatRepo.History(ctx, roomID, after, limit)
}
