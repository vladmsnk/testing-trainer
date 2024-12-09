package entities

import "time"

type Progress struct {
	Id                    int
	GoalID                int
	Username              string
	TotalCompletedPeriods int
	TotalSkippedPeriods   int
	TotalCompletedTimes   int
	MostLongestStreak     int
	CurrentStreak         int
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type ProgressWithGoal struct {
	Habit
	Progress
	Goal
}

type CurrentPeriodProgress struct {
	Habit                       Habit
	CurrentPeriodCompletedTimes int
	NeedToCompleteTimes         int
	CurrentPeriod               int
}
