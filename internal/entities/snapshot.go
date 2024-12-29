package entities

import "time"

type ProgressSnapshot struct {
	Username   string
	CreatedAt  time.Time
	ProgressID int64
	GoalID     int
}
