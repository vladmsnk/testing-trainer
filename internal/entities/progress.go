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

func (p *Progress) DeepCopy() Progress {
	return Progress{
		Id:                    p.Id,
		GoalID:                p.GoalID,
		Username:              p.Username,
		TotalCompletedPeriods: p.TotalCompletedPeriods,
		TotalSkippedPeriods:   p.TotalSkippedPeriods,
		TotalCompletedTimes:   p.TotalCompletedTimes,
		MostLongestStreak:     p.MostLongestStreak,
		CurrentStreak:         p.CurrentStreak,
		CreatedAt:             p.CreatedAt,
		UpdatedAt:             p.UpdatedAt,
	}
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

type ProgressChange struct {
	TotalCompletedPeriods int
	TotalSkippedPeriods   int
	TotalCompletedTimes   int
	MostLongestStreak     int
	CurrentStreak         int
}
