package storage

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type UserStorage struct {
	db *sqlx.DB
}

func NewUserStorage(db *sqlx.DB) *UserStorage {
	return &UserStorage{
		db: db,
	}
}

func (u *UserStorage) StoreChatID(ctx context.Context, chatID int64) error {
	conn, err := u.db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx,
		`INSERT INTO users (chat_id) VALUES ($1) ON CONFLICT (chat_id) DO NOTHING`, chatID)
	return err
}

func (u *UserStorage) RetrieveChatIDs(ctx context.Context) ([]int64, error) {
	conn, err := u.db.Connx(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var chatIDs []int64
	if err := conn.SelectContext(ctx, &chatIDs, `SELECT chat_id FROM users`); err != nil {
		return nil, err
	}

	return chatIDs, nil
}

func (s *UserStorage) IsBotStarted(ctx context.Context, chatID int64) (bool, error) {
	var started bool
	err := s.db.GetContext(ctx, &started, `SELECT started FROM users WHERE chat_id = $1`, chatID)
	return started, err
}

func (s *UserStorage) SetBotStarted(ctx context.Context, chatID int64, started bool) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET started = $1 WHERE chat_id = $2`, started, chatID)
	return err
}
