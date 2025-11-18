package repo

import (
	"context"
	"errors"

	"github.com/NetPo4ki/reward-system/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TasksRepo struct {
	pool *pgxpool.Pool
}

func NewTasksRepo(pool *pgxpool.Pool) *TasksRepo {
	return &TasksRepo{pool: pool}
}

func (r *TasksRepo) GetByCode(ctx context.Context, code string) (*models.Task, error) {
	const q = `SELECT id, code, name, points, active, created_at
	FROM tasks WHERE code=$1`
	var t models.Task
	if err := r.pool.QueryRow(ctx, q, code).Scan(&t.ID, &t.Code, &t.Name, &t.Points, &t.Active, &t.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}
