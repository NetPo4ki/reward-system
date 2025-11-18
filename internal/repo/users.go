package repo

import (
	"context"
	"errors"

	"github.com/NetPo4ki/reward-system/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersRepo struct {
	pool *pgxpool.Pool
}

func NewUsersRepo(pool *pgxpool.Pool) *UsersRepo {
	return &UsersRepo{pool: pool}
}

func (r *UsersRepo) Create(ctx context.Context, username string) (*models.User, error) {
	const q = `INSERT INTO users (username)
	VALUES ($1)
	RETURNING id, username, referrer_id, created_at`
	var u models.User
	if err := r.pool.QueryRow(ctx, q, username).Scan(&u.ID, &u.Username, &u.ReferrerID, &u.CreatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UsersRepo) GetByID(ctx context.Context, id int64) (*models.User, error) {
	const q = `SELECT id, username, referrer_id, created_at
	FROM users WHERE  id=$1`
	var u models.User
	if err := r.pool.QueryRow(ctx, q, id).Scan(&u.ID, &u.Username, &u.ReferrerID, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UsersRepo) SetReferrer(ctx context.Context, userID, referrerID int64) error {
	if userID == referrerID {
		return ErrInvalid
	}

	const existsQ = `SELECT 1 FROM users WHERE id=$1`
	var one int
	if err := r.pool.QueryRow(ctx, existsQ, referrerID).Scan(&one); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	const upd = `UPDATE users SET referrer_id=$2
	WHERE id=$1 AND referrer_id IS NULL`
	tag, err := r.pool.Exec(ctx, upd, userID, referrerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrConflict
	}
	return nil
}

func (r *UsersRepo) Balance(ctx context.Context, userID int64) (int, error) {
	const q = `SELECT COALESCE(SUM(points_awarded),0)
	FROM user_tasks WHERE user_id=$1`
	var balance int
	if err := r.pool.QueryRow(ctx, q, userID).Scan(&balance); err != nil {
		return 0, err
	}
	return balance, nil
}

func (r *UsersRepo) Leaderboard(ctx context.Context, limit int) ([]models.LeaderboardEntry, error) {
	const q = `SELECT u.id, u.username, COALESCE(SUM(ut.points_awarded),0) AS balance
	FROM users u
	LEFT JOIN user_tasks ut ON ut.user_id = u.id
	GROUP BY u.id, u.username
	ORDER BY balance DESC, u.id ASC
	LIMIT $1`
	rows, err := r.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []models.LeaderboardEntry
	for rows.Next() {
		var e models.LeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.Username, &e.Balance); err != nil {
			return nil, err
		}
		res = append(res, e)
	}
	return res, rows.Err()
}
