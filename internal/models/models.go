package models

import "time"

type User struct {
	ID         int64
	Username   string
	ReferrerID *int64
	CreatedAt  time.Time
}

type Task struct {
	ID        int64
	Code      string
	Name      string
	Points    int
	Active    bool
	CreatedAt time.Time
}

type UserTask struct {
	UserID        int64
	TaskID        int64
	PointsAwarded int
	CompletedAt   time.Time
}

type LeaderboardEntry struct {
	UserID   int64
	Username string
	Balance  int
}
