package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test the NewHabit function
func TestNewHabit(t *testing.T) {
	name := "Exercise"
	description := "Daily workout routine"
	goal := &Goal{
		Frequency:       3,
		Duration:        24 * time.Hour,
		NumOfPeriods:    4,
		StartTrackingAt: time.Now(),
	}

	habit := NewHabit(name, description, goal)

	assert.Equal(t, name, habit.Name, "Habit Name should match")
	assert.Equal(t, description, habit.Description, "Habit Description should match")
	assert.NotNil(t, habit.Goal, "Goal should not be nil")
	assert.False(t, habit.IsArchived, "IsArchived should be false by default")
}

// Test the NewGoal function
func TestNewGoal(t *testing.T) {
	frequency := 3
	duration := 24 * time.Hour
	numOfPeriods := 4
	startTrackingAt := time.Now()

	goal := NewGoal(frequency, duration, numOfPeriods, startTrackingAt)

	assert.Equal(t, frequency, goal.Frequency, "Goal Frequency should match")
	assert.Equal(t, duration, goal.Duration, "Goal Duration should match")
	assert.Equal(t, numOfPeriods, goal.NumOfPeriods, "Goal NumOfPeriods should match")
	assert.True(t, goal.StartTrackingAt.Equal(startTrackingAt), "StartTrackingAt should match the input time")

	expectedEndTrackingAt := startTrackingAt.Add(duration * time.Duration(numOfPeriods))
	assert.True(t, goal.EndTrackingAt.Equal(expectedEndTrackingAt), "EndTrackingAt should be correctly calculated")
}
