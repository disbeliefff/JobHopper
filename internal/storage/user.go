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
