package postgres

import (
	"context"
	"time"

	"github.com/cwrk-planet/room-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// todo: перенести все query в queries.go

type ParticipantRepository struct {
	db *pgxpool.Pool
}

func NewParticipantRepository(db *pgxpool.Pool) *ParticipantRepository {
	return &ParticipantRepository{db: db}
}

func (r *ParticipantRepository) CountInRoom(ctx context.Context, roomID string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM room_participants WHERE room_id=$1`, roomID).Scan(&count)
	return count, err
}

func (r *ParticipantRepository) Exists(ctx context.Context, roomID string, userID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM room_participants WHERE room_id=$1 AND user_id=$2)`,
		roomID, userID).Scan(&exists)
	return exists, err
}

// Join — защищён от гонок по max_participants.
// Это гарантирует, что два параллельных Join по одной комнате не пробьют лимит.
func (r *ParticipantRepository) Join(ctx context.Context, p *domain.Participant, maxParticipants int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Блокируем строку комнаты. Параллельные транзакции по той же комнате будут ждать.
	var mp int64
	if err := tx.QueryRow(ctx, `SELECT max_participants FROM rooms WHERE id=$1 FOR UPDATE`, p.RoomID).Scan(&mp); err != nil {
		return err
	}
	if mp > 0 {
		maxParticipants = mp
	}

	var count int64
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM room_participants WHERE room_id=$1`, p.RoomID).Scan(&count); err != nil {
		return err
	}
	if count >= maxParticipants {
		return domain.ErrRoomFull
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO room_participants (room_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, p.RoomID, p.UserID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *ParticipantRepository) Leave(ctx context.Context, roomID string, userID int64) error {
	cmd, err := r.db.Exec(ctx, `DELETE FROM room_participants WHERE room_id=$1 AND user_id=$2`, roomID, userID)
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotInRoom
	}
	if err != nil {
		return err
	}

	return nil
}

func (r *ParticipantRepository) ListByRoom(ctx context.Context, roomID string) ([]domain.Participant, error) {
	rows, err := r.db.Query(ctx,
		`SELECT room_id, user_id, joined_at, last_seen FROM room_participants WHERE room_id=$1 ORDER BY joined_at ASC`,
		roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Participant
	for rows.Next() {
		var p domain.Participant
		if err := rows.Scan(&p.RoomID, &p.UserID, &p.JoinedAt, &p.LastSeen); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (r *ParticipantRepository) TouchHeartbeat(ctx context.Context, roomID string, userID int64) error {
	cmd, err := r.db.Exec(ctx,
		`UPDATE room_participants SET last_seen=now() WHERE room_id=$1 AND user_id=$2`,
		roomID, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotInRoom
	}
	return nil
}

type ParticipantDetailedRow struct {
	UserID      int64
	DisplayName *string
	AvatarURL   *string
	JoinedAt    time.Time
	LastSeen    time.Time
}

// activeWithin — окно «онлайн»; участники со свежим last_seen.
func (r *ParticipantRepository) ListDetailed(ctx context.Context, roomID string, activeWithin time.Duration) ([]ParticipantDetailedRow, error) {
	secs := int64(activeWithin / time.Second)

	const q = `
SELECT m.user_id,
       u.display_name,
       u.avatar_url,
       m.joined_at,
       m.last_seen
FROM public.room_participants AS m
JOIN public.users AS u ON u.id = m.user_id
WHERE m.room_id = $1
  AND m.last_seen > NOW() - ($2::int * INTERVAL '1 second')
ORDER BY u.display_name NULLS LAST, m.joined_at;
`
	rows, err := r.db.Query(ctx, q, roomID, secs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ParticipantDetailedRow, 0, 16)
	for rows.Next() {
		var row ParticipantDetailedRow
		if err := rows.Scan(
			&row.UserID,
			&row.DisplayName,
			&row.AvatarURL,
			&row.JoinedAt,
			&row.LastSeen,
		); err != nil {
			return nil, err
		}
		out = append(out, row)
	}

	return out, rows.Err()
}
