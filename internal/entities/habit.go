package entities

type Habit struct {
	Name        string
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

func NewHabit(name, description string, goal *Goal) Habit {
	return Habit{
		Name:        name,
		Description: description,
		Goal:        goal,
		IsArchived:  false,
	}
}
