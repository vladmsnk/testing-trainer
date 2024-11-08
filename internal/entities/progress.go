package entities

type Progress struct {
	TotalCompletedPeriods int
	TotalSkippedPeriods   int
	TotalCompletedTimes   int
	MostLongestStreak     int
	CurrentStreak         int
}

type ProgressWithGoal struct {
	Progress
	Goal
}
