package entities

import "time"

type ProgressSnapshot struct {
	Username           string
	CurrentProgressIDs []int64
	CreatedAt          time.Time
	ProgressID         int64
	GoalID             int
}
