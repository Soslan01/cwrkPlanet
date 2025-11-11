package postgres

import (
	"context"
	"fmt"

	"github.com/cwrk-planet/room-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepository struct {
	db *pgxpool.Pool
}

func NewChatRepository(db *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) Save(ctx context.Context, roomID string, userID int64, text string, replyTo *string) (*domain.ChatMessage, error) {
	// todo: перенести в отдельный файл queries.go
	row := r.db.QueryRow(ctx, `
		INSERT INTO room_messages (room_id, user_id, text, reply_to)
		VALUES ($1, $2, $3, $4)
		RETURNING id, room_id, user_id, text, reply_to, created_at
	`, roomID, userID, text, replyTo)

	var m domain.ChatMessage
	if err := row.Scan(&m.ID, &m.RoomID, &m.UserID, &m.Text, &m.ReplyTo, &m.CreatedAt); err != nil {
		return nil, err
	}
	return &m, nil
}

// History возвращает историю сообщений комнаты с курсорной пагинацией (created_at,id DESC).
func (r *ChatRepository) History(ctx context.Context, roomID, after string, limit int) ([]domain.ChatMessage, string, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	cur, err := DecodeCursor(after)
	if err != nil {
		return nil, "", fmt.Errorf("decode cursor: %w", err)
	}
	// todo: перенести в отдельный файл queries.go
	const baseQuery = `
		SELECT id, room_id, user_id, text, reply_to, created_at
		FROM room_messages
		WHERE room_id = $1
		  AND (
		    $2::timestamptz IS NULL
		    OR created_at < $2
		    OR (created_at = $2 AND id < $3)
		  )
		ORDER BY created_at DESC, id DESC
		LIMIT $4
	`

	var createdAt any
	var id any
	if cur != nil {
		createdAt = cur.CreatedAt
		id = cur.ID
	}

	rows, err := r.db.Query(ctx, baseQuery, roomID, createdAt, id, limit)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var out []domain.ChatMessage
	for rows.Next() {
		var m domain.ChatMessage
		if err := rows.Scan(&m.ID, &m.RoomID, &m.UserID, &m.Text, &m.ReplyTo, &m.CreatedAt); err != nil {
			return nil, "", err
		}
		out = append(out, m)
	}

	var next string
	if len(out) == limit {
		last := out[len(out)-1]
		if c, e := EncodeCursor(Cursor{CreatedAt: last.CreatedAt, ID: last.ID}); e == nil {
			next = c
		}
	}
	return out, next, nil
}
