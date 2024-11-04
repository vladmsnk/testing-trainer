package entities

type Habit struct {
	Id          string
	Description string
	Goal        *Goal
	IsArchived  bool
}

type Goal struct {
	Id                   int
	FrequencyType        FrequencyType
	TimesPerFrequency    int
	TotalTrackingPeriods int
	IsActive             bool
}

func IsGoalChanged(old, new *Goal) bool {
	if old == nil && new == nil {
		return false
	}

	if old == nil && new != nil || new == nil && old != nil {
		return true
	}

	if *old != *new {
		return true
	}

	return false
}

type FrequencyType int64

const (
	UndefinedFrequencyType FrequencyType = iota
	Daily
	Weekly
	Monthly
)

func (f FrequencyType) String() string {
	switch f {
	case Daily:
		return "daily"
	case Weekly:
		return "weekly"
	case Monthly:
		return "monthly"
	default:
		return "undefined"
	}
}

func FrequencyTypeFromString(s string) FrequencyType {
	switch s {
	case "daily":
		return Daily
	case "weekly":
		return Weekly
	case "monthly":
		return Monthly
	default:
		return UndefinedFrequencyType
	}

}
