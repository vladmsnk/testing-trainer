package storage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing_trainer/internal/storage/transactor"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"testing_trainer/internal/entities"
)

var (
	ErrNotFound = errors.New("error no rows")
)

type Storage struct {
	db                  *pgxpool.Pool
	queryEngineProvider transactor.QueryEngineProvider
}

func NewStorage(db *pgxpool.Pool) *Storage {
	txManager, err := transactor.New(db)
	if err != nil {
		log.Fatal(err)
	}

	return &Storage{db: db, queryEngineProvider: txManager}
}

func (s *Storage) CreateHabit(ctx context.Context, username string, habit entities.Habit) (int, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("db.BeginTx: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
			return
		}
		err = tx.Commit(ctx)
	}()

	habitId, err := s.createHabitTx(ctx, tx, username, habit)
	if err != nil {
		return 0, fmt.Errorf("s.createHabitTx: %w", err)
	}

	err = s.createGoalTx(ctx, tx, habitId, habit.Goal)
	if err != nil {
		return 0, fmt.Errorf("s.createGoalTx: %w", err)
	}

	return habitId, nil
}

func (s *Storage) createHabitTx(ctx context.Context, tx pgx.Tx, username string, habit entities.Habit) (int, error) {
	query := `
INSERT INTO habits (username, name, description)
VALUES ($1, $2, $3)
returning id;
`

	var habitID int

	err := tx.QueryRow(ctx, query, username, habit.Name, habit.Description).Scan(&habitID)
	if err != nil {
		return 0, fmt.Errorf("tx.Exec user_id=%s description=%s: %w", username, habit.Description, err)
	}

	return habitID, nil
}

func (s *Storage) CreateGoal(ctx context.Context, habitID int, goal entities.Goal) (int, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
INSERT INTO goals (habit_id, frequency_type, times_per_frequency, total_tracking_periods)
VALUES ($1, $2, $3, $4)
RETURNING id;
`

	var goalId int

	err := pool.QueryRow(ctx, query, habitID, goal.FrequencyType, goal.TimesPerFrequency, goal.TotalTrackingPeriods).Scan(&goalId)
	if err != nil {
		return 0, fmt.Errorf("tx.Exec habit_id=%d frequency_type=%s times_per_frequency=%d total_tracking_periods=%d: %w", habitID, goal.FrequencyType, goal.TimesPerFrequency, goal.TotalTrackingPeriods, err)
	}

	return goalId, nil
}

func (s *Storage) createGoalTx(ctx context.Context, tx pgx.Tx, habitID int, goal *entities.Goal) error {
	if goal == nil {
		return nil
	}

	query := `
INSERT INTO goals (habit_id, frequency_type, times_per_frequency, total_tracking_periods)
VALUES ($1, $2, $3, $4);
`

	_, err := tx.Exec(ctx, query, habitID, goal.FrequencyType, goal.TimesPerFrequency, goal.TotalTrackingPeriods)
	if err != nil {
		return fmt.Errorf("tx.Exec habit_id=%d frequency_type=%s times_per_frequency=%d total_tracking_periods=%d: %w", habitID, goal.FrequencyType, goal.TimesPerFrequency, goal.TotalTrackingPeriods, err)
	}

	return nil
}

func (s *Storage) GetUserByUsername(ctx context.Context, username string) (entities.User, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
select username, email, password_hash from users where username = $1;
`
	var user entities.User

	err := pool.QueryRow(ctx, query, username).Scan(&user.Name, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.User{}, ErrNotFound
		}
		return entities.User{}, fmt.Errorf("s.db.QueryRow: %w", err)
	}

	return user, nil
}

func (s *Storage) CreateUser(ctx context.Context, user entities.RegisterUser) error {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
insert into users (username, email, password_hash) 
values ($1, $2, $3);
`
	_, err := pool.Exec(ctx, query, user.Name, user.Email, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("s.db.Exec: %w", err)
	}

	return nil
}

func (s *Storage) ListUserHabits(ctx context.Context, username string) ([]entities.Habit, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
select 
    h.id as id, 
    h.description as description,
    g.id as goal_id,
    g.frequency_type as frequency_type, 
    g.times_per_frequency as times_per_frequency,
    g.total_tracking_periods as total_tracking_periods
from habits h 
    left join goals g on h.id = g.habit_id and g.is_active = true
where h.username = $1;
`

	rows, err := pool.Query(ctx, query, username)
	if err != nil {
		return nil, fmt.Errorf("s.db.Query: %w", err)
	}
	defer rows.Close()

	var result []entities.Habit
	for rows.Next() {
		var daoHabit habit

		err := rows.Scan(&daoHabit.id, &daoHabit.description, &daoHabit.goalId, &daoHabit.frequencyType, &daoHabit.timesPerFrequency, &daoHabit.totalTrackingPeriods)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		result = append(result, toEntityHabit(daoHabit))
	}

	return result, nil
}

func (s *Storage) GetHabitById(ctx context.Context, username string, habitId int) (entities.Habit, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
select 
    h.id as id, 
    h.description as description,
    g.id as goal_id,
    g.frequency_type as frequency_type, 
    g.times_per_frequency as times_per_frequency,
    g.total_tracking_periods as total_tracking_periods
from habits h 
    left join goals g on h.id = g.habit_id 
where h.username = $1 
  and h.id = $2;`

	var daoHabit habit

	err := pool.QueryRow(ctx, query, username, habitId).Scan(&daoHabit.id, &daoHabit.description, &daoHabit.goalId, &daoHabit.frequencyType, &daoHabit.timesPerFrequency, &daoHabit.totalTrackingPeriods)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.Habit{}, ErrNotFound
		}
		return entities.Habit{}, fmt.Errorf("s.db.QueryRow: %w", err)
	}

	return toEntityHabit(daoHabit), nil
}

func (s *Storage) GetHabitGoal(ctx context.Context, habitId int) (entities.Goal, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
select id, frequency_type, times_per_frequency, total_tracking_periods, created_at from goals where habit_id = $1 and is_active = true;
	`
	var daoGoal goal

	err := pool.QueryRow(ctx, query, habitId).Scan(&daoGoal.id, &daoGoal.frequencyType, &daoGoal.timesPerFrequency, &daoGoal.totalTrackingPeriods, &daoGoal.createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.Goal{}, ErrNotFound
		}
		return entities.Goal{}, fmt.Errorf("s.db.QueryRow: %w", err)
	}

	return toEntityGoal(daoGoal), nil
}

func (s *Storage) DeactivateGoalByID(ctx context.Context, id int) error {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `update goals set is_active = false where id = $1;`

	_, err := pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("s.db.Exec: %w", err)
	}

	return nil
}

func (s *Storage) AddHabitProgress(ctx context.Context, goalId int) error {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
insert into goal_logs (goal_id) values ($1);
`

	_, err := pool.Exec(ctx, query, goalId)
	if err != nil {
		return fmt.Errorf("s.db.Exec: %w", err)
	}

	return nil
}

func (s *Storage) UpdateGoalStat(ctx context.Context, goalId int, progress entities.Progress) error {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
	INSERT INTO goal_stats (goal_id, total_completed_periods, total_completed_times, most_longest_streak, current_streak)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (goal_id) 
	DO UPDATE SET 
		total_completed_periods = $2,
		total_completed_times = $3,
		most_longest_streak = $4,
		current_streak = $5,
		updated_at = now();
	`

	_, err := pool.Exec(ctx, query, goalId, progress.TotalCompletedPeriods, progress.TotalCompletedTimes, progress.MostLongestStreak, progress.CurrentStreak)
	if err != nil {
		return fmt.Errorf("s.db.Exec: %w", err)
	}

	return nil
}

func calculatePeriodRange(createdAt time.Time, frequencyType entities.FrequencyType, periodOffset int) (start time.Time, end time.Time) {
	switch frequencyType {
	case entities.Daily:
		start = createdAt.AddDate(0, 0, periodOffset)
		end = start.AddDate(0, 0, 1)
	case entities.Weekly:
		start = createdAt.AddDate(0, 0, periodOffset*7)
		end = start.AddDate(0, 0, 7)
	case entities.Monthly:
		start = createdAt.AddDate(0, periodOffset, 0)
		end = start.AddDate(0, 1, 0)
	}
	return start, end
}

func (s *Storage) getExecutionCountForPeriod(ctx context.Context, goalId int, start, end time.Time) (int, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
    SELECT COUNT(*) 
    FROM goal_logs 
    WHERE goal_id = $1 
    AND record_created_at >= $2 
    AND record_created_at < $3;
    `

	var count int
	err := pool.QueryRow(ctx, query, goalId, start, end).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("db.QueryRow: %w", err)
	}

	return count, nil
}

func (s *Storage) GetPreviousPeriodExecutionCount(ctx context.Context, goalId int, frequencyType entities.FrequencyType, createdAt time.Time, currentPeriod int) (int, error) {
	// The first period has no previous period
	if currentPeriod == 0 {
		return 0, nil
	}

	start, end := calculatePeriodRange(createdAt, frequencyType, currentPeriod-1)
	return s.getExecutionCountForPeriod(ctx, goalId, start, end)
}

func (s *Storage) GetCurrentPeriodExecutionCount(ctx context.Context, goalId int, frequencyType entities.FrequencyType, createdAt time.Time, currentPeriod int) (int, error) {
	start, end := calculatePeriodRange(createdAt, frequencyType, currentPeriod)
	return s.getExecutionCountForPeriod(ctx, goalId, start, end)
}

func (s *Storage) GetCurrentProgress(ctx context.Context, goalId int) (entities.Progress, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
select total_completed_periods,total_skipped_periods, total_completed_times, most_longest_streak, current_streak from goal_stats where goal_id = $1;
`

	var progress entities.Progress
	err := pool.QueryRow(ctx, query, goalId).Scan(&progress.TotalCompletedPeriods, &progress.TotalSkippedPeriods, &progress.TotalCompletedTimes, &progress.MostLongestStreak, &progress.CurrentStreak)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.Progress{}, nil
		}
		return entities.Progress{}, fmt.Errorf("db.QueryRow: %w", err)
	}

	return progress, nil
}

func (s *Storage) SetGoalCompleted(ctx context.Context, goalId int) error {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
update goals set is_completed = true where id = $1;
`
	_, err := pool.Exec(ctx, query, goalId)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}
