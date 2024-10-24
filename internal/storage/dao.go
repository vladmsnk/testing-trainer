package storage

import (
	"database/sql"
)

type habit struct {
	id                   int              `db:"id"`
	name                 string           `db:"id"`
	description          string           `db:"description"`
	frequencyType        sql.Null[string] `db:"frequency_type"`
	timesPerFrequency    sql.Null[int]    `db:"times_per_frequency"`
	totalTrackingPeriods sql.Null[int]    `db:"total_tracking_periods"`
}

type goal struct {
	id                   int    `db:"id"`
	frequencyType        string `db:"frequency_type"`
	timesPerFrequency    int    `db:"times_per_frequency"`
	totalTrackingPeriods int    `db:"total_tracking_periods"`
}
