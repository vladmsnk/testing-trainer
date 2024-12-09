package storage

import (
	"database/sql"
	"github.com/jackc/pgx/v5/pgtype"
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
	startTrackingAt      sql.Null[time.Time] `db:"start_tracking_at"`
	isCompleted          sql.Null[bool]      `db:"is_completed"`
	previousGoalId       sql.Null[int]       `db:"previous_goal_id"`
	previousGoalIDs      []int               `db:"previous_goal_ids"`
}

type goal struct {
	id                   int           `db:"id"`
	frequencyType        string        `db:"frequency_type"`
	timesPerFrequency    int           `db:"times_per_frequency"`
	totalTrackingPeriods int           `db:"total_tracking_periods"`
	createdAt            time.Time     `db:"created_at"`
	nextCheckDate        time.Time     `db:"next_check_date"`
	isCompleted          bool          `db:"is_completed"`
	isActive             bool          `db:"is_active"`
	previousGoalId       sql.Null[int] `db:"previous_goal_id"`
	previousGoalIDs      []int         `db:"previous_goal_ids"`
	startTrackingAt      time.Time     `db:"start_tracking_at"`
}

type userProgress struct {
	username              string `db:"username"`
	id                    int    `db:"id"`
	goalId                int    `db:"goal_id"`
	totalCompletedPeriods int    `db:"total_completed_periods"`
	totalCompletedTimes   int    `db:"total_completed_times"`
	totalSkippedPeriods   int    `db:"total_skipped_periods"`
	mostLongestStreak     int    `db:"most_longest_streak"`
	currentStreak         int    `db:"current_streak"`
}

type snapshot struct {
	username           string              `db:"username"`
	currentProgressIDs pgtype.Array[int64] `db:"current_progress_ids"`
	createdAt          time.Time           `db:"time_swticher"`
}
