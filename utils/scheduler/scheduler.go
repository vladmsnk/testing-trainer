package scheduler

import (
	"fmt"

	"github.com/go-co-op/gocron/v2"
)

func NewScheduler(cron gocron.JobDefinition, tasks []gocron.Task) (gocron.Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("gocron.NewScheduler: %w", err)
	}

	for _, task := range tasks {
		_, err := s.NewJob(cron, task)
		if err != nil {
			return nil, fmt.Errorf("s.NewJob: %w", err)
		}
	}

	s.Start()

	return s, nil
}
