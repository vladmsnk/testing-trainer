package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgxpool"
	"testing_trainer/internal/entities"
)

var (
	ErrNotFound = errors.New("error no rows")
)

type Storage struct {
	db *pgxpool.Pool
}

func NewStorage(db *pgxpool.Pool) *Storage {
	return &Storage{db: db}
}

func (s *Storage) CreateHabit(ctx context.Context, username string, habit entities.Habit) (int64, error) {
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

func (s *Storage) createHabitTx(ctx context.Context, tx pgx.Tx, username string, habit entities.Habit) (int64, error) {
	query := `
INSERT INTO habits (username, name, description)
VALUES ($1, $2, $3)
returning id;
`

	var habitID int64

	err := tx.QueryRow(ctx, query, username, habit.Name, habit.Description).Scan(&habitID)
	if err != nil {
		return 0, fmt.Errorf("tx.Exec user_id=%s habit=%s description=%s: %w", username, habit.Name, habit.Description, err)
	}

	return habitID, nil
}

func (s *Storage) createGoalTx(ctx context.Context, tx pgx.Tx, habitID int64, goal *entities.Goal) error {
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
	query := `
select username, email, password_hash from users where username = $1;
`
	var user entities.User

	err := s.db.QueryRow(ctx, query, username).Scan(&user.Name, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.User{}, ErrNotFound
		}
		return entities.User{}, fmt.Errorf("s.db.QueryRow: %w", err)
	}

	return user, nil
}

func (s *Storage) CreateUser(ctx context.Context, user entities.RegisterUser) error {
	query := `
insert into users (username, email, password_hash) 
values ($1, $2, $3);
`
	_, err := s.db.Exec(ctx, query, user.Name, user.Email, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("s.db.Exec: %w", err)
	}

	return nil
}

func (s *Storage) ListUserHabits(ctx context.Context, username string) ([]entities.Habit, error) {
	query := `
select 
    h.id as id, 
    h.name as name, 
    h.description as description,
    g.frequency_type as frequency_type, 
    g.times_per_frequency as times_per_frequency,
    g.total_tracking_days as total_tracking_days
from habits h 
    left join goals g on h.id = g.habit_id 
where h.username = $1;
`

	rows, err := s.db.Query(ctx, query, username)
	if err != nil {
		return nil, fmt.Errorf("s.db.Query: %w", err)
	}
	defer rows.Close()

	var result []entities.Habit
	for rows.Next() {
		var daoHabit habit

		err := rows.Scan(&daoHabit.id, &daoHabit.name, &daoHabit.description, &daoHabit.frequencyType, &daoHabit.timesPerFrequency, &daoHabit.totalTrackingPeriods)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		result = append(result, toEntityHabit(daoHabit))
	}

	return result, nil
}

func (s *Storage) GetHabitByName(ctx context.Context, username, habitName string) (entities.Habit, error) {
	query := `
select 
    h.id as id, 
    h.name as name, 
    h.description as description,
    g.frequency_type as frequency_type, 
    g.times_per_frequency as times_per_frequency,
    g.total_tracking_days as total_tracking_days
from habits h 
    left join goals g on h.id = g.habit_id 
where h.username = $1 
  and h.name = $2;`

	var daoHabit habit

	err := s.db.QueryRow(ctx, query, username, habitName).Scan(&daoHabit.id, &daoHabit.name, &daoHabit.description, &daoHabit.frequencyType, &daoHabit.timesPerFrequency, &daoHabit.totalTrackingPeriods)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.Habit{}, ErrNotFound
		}
		return entities.Habit{}, fmt.Errorf("s.db.QueryRow: %w", err)
	}

	return toEntityHabit(daoHabit), nil
}

func (s *Storage) UpdateHabit(ctx context.Context, username string, habit entities.Habit) error {
	return nil
}

func (s *Storage) GetHabitGoal(ctx context.Context, habitName string) (entities.Goal, error) {
	query := `
select id, frequency_type, times_per_frequency, total_tracking_periods from goals where habit_id = (select id from habits where name = $1);
	`
	var daoGoal goal

	err := s.db.QueryRow(ctx, query, habitName).Scan(&daoGoal.id, &daoGoal.frequencyType, &daoGoal.timesPerFrequency, &daoGoal.totalTrackingPeriods)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.Goal{}, ErrNotFound
		}
		return entities.Goal{}, fmt.Errorf("s.db.QueryRow: %w", err)
	}

	return toEntityGoal(daoGoal), nil
}

func (s *Storage) deactivateGoalByID(ctx context.Context, id int64) error {
	query := `update goals set is_active = false where id = $1;`

	_, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("s.db.Exec: %w", err)
	}

	return nil
}

func (s *Storage) AddHabitProgress(ctx context.Context, goalId int) error {
	query := `
insert into goal_logs (goal_id) values ($1);
`

	_, err := s.db.Exec(ctx, query, goalId)
	if err != nil {
		return fmt.Errorf("s.db.Exec: %w", err)
	}

	return nil
}

func (s *Storage) UpdateGoalStat(ctx context.Context, goalId int, progress entities.Progress) error {
	query := `
	INSERT INTO goal_stats (goal_id, total_completed_periods, total_completed_times, most_longest_streak, current_streak)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (goal_id) 
	DO UPDATE SET 
		total_completed_periods = $2,
		total_completed_times = $3,
		most_longest_streak = $4,
		current_streak = $5;
	`

	_, err := s.db.Exec(ctx, query, goalId, progress.TotalCompletedPeriods, progress.TotalCompletedTimes, progress.MostLongestStreak, progress.CurrentStreak)
	if err != nil {
		return fmt.Errorf("s.db.Exec: %w", err)
	}

	return nil
}

func (s *Storage) GetHabitProgress(ctx context.Context, username, habitName string) (entities.Progress, error) {
	return entities.Progress{}, nil
}

func (s *Storage) GetPreviousPeriodExecutionCount(ctx context.Context, goalId int, frequencyType entities.FrequencyType) (int, error) {
	var query string

	switch frequencyType {
	case entities.Daily:
		query = `
		SELECT COUNT(*) 
		FROM goal_logs 
		WHERE goal_id = $1 
		AND record_created_at::date = current_date - interval '1 day';
		`
	case entities.Weekly:
		query = `
		SELECT COUNT(*) 
		FROM goal_logs 
		WHERE goal_id = $1 
		AND record_created_at >= date_trunc('week', current_date - interval '1 week') 
		AND record_created_at < date_trunc('week', current_date);
		`
	case entities.Monthly:
		query = `
		SELECT COUNT(*) 
		FROM goal_logs 
		WHERE goal_id = $1 
		AND record_created_at >= date_trunc('month', current_date - interval '1 month') 
		AND record_created_at < date_trunc('month', current_date);
		`
	default:
		return 0, fmt.Errorf("unsupported frequncy type: %v", frequencyType)
	}

	var count int
	err := s.db.QueryRow(ctx, query, goalId).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("db.QueryRow: %w", err)
	}

	return count, nil
}
func (s *Storage) GetCurrentPeriodExecutionCount(ctx context.Context, goalId int, frequencyType entities.FrequencyType) (int, error) {
	var query string

	switch frequencyType {
	case entities.Daily:
		query = `
		SELECT COUNT(*) 
		FROM goal_logs 
		WHERE goal_id = $1 
		AND record_created_at::date = current_date;
		`
	case entities.Weekly:
		query = `
		SELECT COUNT(*) 
		FROM goal_logs 
		WHERE goal_id = $1 
		AND record_created_at >= date_trunc('week', current_date) 
		AND record_created_at < date_trunc('week', current_date + interval '1 week');
		`
	case entities.Monthly:
		query = `
		SELECT COUNT(*) 
		FROM goal_logs 
		WHERE goal_id = $1 
		AND record_created_at >= date_trunc('month', current_date) 
		AND record_created_at < date_trunc('month', current_date + interval '1 month');
		`
	default:
		return 0, fmt.Errorf("unsupported frequency type: %v", frequencyType)
	}

	var count int
	err := s.db.QueryRow(ctx, query, goalId).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("db.QueryRow: %w", err)
	}

	return count, nil
}

func (s *Storage) GetCurrentProgress(ctx context.Context, goalId int) (entities.Progress, error) {
	query := `
select total_completed_periods,total_skipped_periods, total_completed_times, most_longest_streak, current_streak from goal_stats where goal_id = $1;
`

	var progress entities.Progress
	err := s.db.QueryRow(ctx, query, goalId).Scan(&progress.TotalCompletedPeriods, &progress.TotalSkippedPeriods, &progress.TotalCompletedTimes, &progress.MostLongestStreak, &progress.CurrentStreak)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.Progress{}, nil
		}
		return entities.Progress{}, fmt.Errorf("db.QueryRow: %w", err)
	}

	return progress, nil
}
