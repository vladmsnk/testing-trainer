package entities

type Progress struct {
	TotalCompletedPeriods int
	TotalSkippedPeriods   int
	TotalCompletedTimes   int
	MostLongestStreak     int
	CurrentStreak         int
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
