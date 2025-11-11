package postgres

import (
	"context"

	"github.com/cwrk-planet/room-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepository struct {
	db *pgxpool.Pool
}

func NewRoomRepository(db *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) Create(ctx context.Context, room *domain.Room) error {
	query := `
		INSERT INTO rooms (name, max_participants)
		VALUES ($1, $2)
		RETURNING id, created_at`
	err := r.db.QueryRow(ctx, query, room.Name, room.MaxParticipants).Scan(&room.ID, &room.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *RoomRepository) Get(ctx context.Context, id string) (*domain.Room, error) {
	var rm domain.Room
	query := `SELECT id, name, max_participants, created_at FROM rooms WHERE id=$1`
	err := r.db.QueryRow(ctx, query, id).
		Scan(&rm.ID, &rm.Name, &rm.MaxParticipants, &rm.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrRoomNotFound
		}
		return nil, err
	}
	return &rm, nil
}

func (r *RoomRepository) List(ctx context.Context, limit int, cursorStr string) ([]domain.Room, string, error) {
	cur, err := DecodeCursor(cursorStr)
	if err != nil {
		return nil, "", err
	}

	query := `
		SELECT id, name, max_participants, created_at
		FROM rooms
		WHERE ($1::timestamptz IS NULL OR created_at < $1
		       OR (created_at = $1 AND id < $2))
		ORDER BY created_at DESC, id DESC
		LIMIT $3`

	var createdAt any
	var id any
	if cur != nil {
		createdAt = cur.CreatedAt
		id = cur.ID
	}

	rows, err := r.db.Query(ctx, query, createdAt, id, limit)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var rooms []domain.Room
	for rows.Next() {
		var r domain.Room
		if err := rows.Scan(&r.ID, &r.Name, &r.MaxParticipants, &r.CreatedAt); err != nil {
			return nil, "", err
		}
		rooms = append(rooms, r)
	}

	var nextCursor string
	if len(rooms) == limit {
		last := rooms[len(rooms)-1]
		cur := Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
		nextCursor, _ = EncodeCursor(cur)
	}

	return rooms, nextCursor, nil
}

func (r *RoomRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM rooms WHERE id=$1`, id)
	return err
}
