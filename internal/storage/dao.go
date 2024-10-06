package storage

import (
	"database/sql"
	"time"
)

type habit struct {
	id              int           `db:"id"`
	name            string        `db:"id"`
	description     string        `db:"id"`
	frequency       sql.NullInt64 `db:"frequency"`
	duration        sql.NullInt64 `db:"duration"`
	numOfPeriods    sql.NullInt64 `db:"num_of_periods"`
	startTrackingAt sql.NullTime  `db:"start_tracking_at"`
	endTrackingAt   sql.NullTime  `db:"end_tracking_at"`
}

type goal struct {
	frequency       int       `db:"frequency"`
	duration        int64     `db:"duration"`
	numOfPeriods    int       `db:"num_of_periods"`
	startTrackingAt time.Time `db:"start_tracking_at"`
	endTrackingAt   time.Time `db:"end_tracking_at"`
}
