package entities

import (
	"testing"
	"time"
)

func TestGetCurrentPeriod(t *testing.T) {
	tests := []struct {
		name            string
		frequencyType   FrequencyType
		startTrackingAt time.Time
		expected        int
	}{
		{
			name:            "Daily - 3 days since creation",
			frequencyType:   Daily,
			startTrackingAt: time.Now().AddDate(0, 0, -3), // 3 days ago
			expected:        3,
		},
		{
			name:            "Weekly - 0 week since creation",
			frequencyType:   Weekly,
			expected:        0,
			startTrackingAt: time.Now().AddDate(0, 0, -1),
		},
		{
			name:            "Weekly - 1 week since creation",
			frequencyType:   Weekly,
			expected:        1,
			startTrackingAt: time.Now().AddDate(0, 0, -8),
		},
		{
			name:            "Weekly - 2 weeks since creation",
			frequencyType:   Weekly,
			startTrackingAt: time.Now().AddDate(0, 0, -14), // 2 weeks ago
			expected:        2,
		},
		{
			name:            "Monthly - 5 months since creation",
			frequencyType:   Monthly,
			startTrackingAt: time.Now().AddDate(0, -5, 0), // 5 months ago
			expected:        4,
		},
		{
			name:            "Monthly - 1 year and 2 months since creation",
			frequencyType:   Monthly,
			startTrackingAt: time.Now().AddDate(-1, -2, 0), // 1 year and 2 months ago
			expected:        13,                            // 12 months + 2 months
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goal := Goal{
				FrequencyType:   tt.frequencyType,
				StartTrackingAt: tt.startTrackingAt,
			}

			result := goal.GetCurrentPeriod()
			if result != tt.expected {
				t.Errorf("got %d, want %d", result, tt.expected)
			}
		})
	}
}
