package repo

import (
	"context"

	"github.com/NetPo4ki/reward-system/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserTasksRepo struct {
	pool *pgxpool.Pool
}

func NewUserTasksRepo(pool *pgxpool.Pool) *UserTasksRepo {
	return &UserTasksRepo{pool: pool}
}

func (r *UserTasksRepo) CompleteTask(ctx context.Context, userID int64, taskCode string) (bool, error) {
	const q = `INSERT INTO user_tasks (user_id, task_id, points_awarded)
	SELECT $1, t.id, t.points
	FROM tasks t
	WHERE t.code=$2 AND t.active = TRUE
	ON CONFLICT (user_id, task_id) DO NOTHING
	RETURNING user_id, task_id, points_awarded, completed_at`
	var ut models.UserTask
	err := r.pool.QueryRow(ctx, q, userID, taskCode).Scan(&ut.UserID, &ut.TaskID, &ut.PointsAwarded, &ut.CompletedAt)
	if err != nil {
		const existsQ = `SELECT 1 FROM tasks WHERE code=$1 AND active=TRUE`
		var one int
		if err2 := r.pool.QueryRow(ctx, existsQ, taskCode).Scan(&one); err2 != nil {
			return false, ErrNotFound
		}
		return false, nil
	}
	return true, nil
}

func (r *UserTasksRepo) ListCompleted(ctx context.Context, userID int64) ([]models.UserTask, error) {
	const q = `SELECT user_id, task_id, points_awarded, completed_at
	FROM user_tasks
	WHERE user_id=$1
	ORDER BY completed_at DESC`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []models.UserTask
	for rows.Next() {
		var ut models.UserTask
		if err := rows.Scan(&ut.UserID, &ut.TaskID, &ut.PointsAwarded, &ut.CompletedAt); err != nil {
			return nil, err
		}
		res = append(res, ut)
	}
	return res, rows.Err()
}
