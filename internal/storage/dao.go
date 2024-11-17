package storage

import (
	"database/sql"
	"time"
)

type habit struct {
	id                   int                 `db:"id"`
	name                 string              `db:"name"`
	description          string              `db:"description"`
	goalId               sql.Null[int]       `db:"goal_id"`
	frequencyType        sql.Null[string]    `db:"frequency_type"`
	timesPerFrequency    sql.Null[int]       `db:"times_per_frequency"`
	totalTrackingPeriods sql.Null[int]       `db:"total_tracking_periods"`
	nextCheckDate        sql.Null[time.Time] `db:"next_check_date"`
	createdAt            sql.Null[time.Time] `db:"created_at"`
	previousGoalId       sql.Null[int]       `db:"previous_goal_id"`
}

type goal struct {
	id                   int       `db:"id"`
	frequencyType        string    `db:"frequency_type"`
	timesPerFrequency    int       `db:"times_per_frequency"`
	totalTrackingPeriods int       `db:"total_tracking_periods"`
	createdAt            time.Time `db:"created_at"`
	nextCheckDate        time.Time `db:"next_check_date"`
	isCompleted          bool      `db:"is_completed"`
	isActive             bool      `db:"is_active"`
	previousGoalId       int       `db:"previous_goal_id"`
}
