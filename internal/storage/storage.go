package storage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"testing_trainer/internal/entities"
	"testing_trainer/internal/storage/transactor"
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
INSERT INTO goals (habit_id, frequency_type, times_per_frequency, total_tracking_periods, is_active, next_check_date, previous_goal_id, start_tracking_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id;
`

	var goalId int

	err := pool.QueryRow(ctx, query, habitID, goal.FrequencyType, goal.TimesPerFrequency, goal.TotalTrackingPeriods, true, goal.NextCheckDate, goal.PreviousGoalIDs, goal.StartTrackingAt).Scan(&goalId)
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
INSERT INTO goals (habit_id, frequency_type, times_per_frequency, total_tracking_periods, is_active, next_check_date, start_tracking_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);
`

	_, err := tx.Exec(ctx, query, habitID, goal.FrequencyType, goal.TimesPerFrequency, goal.TotalTrackingPeriods, true, goal.NextCheckDate, goal.StartTrackingAt)
	if err != nil {
		return fmt.Errorf("tx.Exec habit_id=%d frequency_type=%s times_per_frequency=%d total_tracking_periods=%d: %w", habitID, goal.FrequencyType, goal.TimesPerFrequency, goal.TotalTrackingPeriods, err)
	}

	return nil
}

func (s *Storage) GetUserByUsername(ctx context.Context, username string) (entities.User, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
select username,
       email, 
       password_hash 
from users
where username = $1;
`
	var user entities.User

	err := pool.QueryRow(ctx, query, username).Scan(&user.Name, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.User{}, ErrNotFound
		}
		return entities.User{}, fmt.Errorf("db.QueryRow: %w", err)
	}

	return user, nil
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (entities.User, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
select username,
       email, 
       password_hash 
from users 
where email = $1;
`
	var user entities.User

	err := pool.QueryRow(ctx, query, email).Scan(&user.Name, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.User{}, ErrNotFound
		}
		return entities.User{}, fmt.Errorf("db.QueryRow: %w", err)
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
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

func (s *Storage) ListUserCompletedHabits(ctx context.Context, username string) ([]entities.Habit, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
select 
    h.id as id, 
    h.description as description,
    g.id as goal_id,
    g.frequency_type as frequency_type, 
    g.times_per_frequency as times_per_frequency,
    g.total_tracking_periods as total_tracking_periods,
	g.next_check_date as next_check_date
from habits h 
    join goals g on h.id = g.habit_id and g.is_completed = true
where h.username = $1;
`
	rows, err := pool.Query(ctx, query, username)
	if err != nil {
		return nil, fmt.Errorf("db.Query username=%s: %w", username, err)
	}
	defer rows.Close()

	var result []entities.Habit
	for rows.Next() {
		var daoHabit habit

		err := rows.Scan(&daoHabit.id, &daoHabit.description, &daoHabit.goalId, &daoHabit.frequencyType, &daoHabit.timesPerFrequency, &daoHabit.totalTrackingPeriods, &daoHabit.nextCheckDate)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		result = append(result, toEntityHabit(daoHabit))
	}

	return result, nil
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
    g.total_tracking_periods as total_tracking_periods,
	g.next_check_date as next_check_date
from habits h 
    left join goals g on h.id = g.habit_id and g.is_active = true
where h.username = $1 and h.is_archived = false;
`

	rows, err := pool.Query(ctx, query, username)
	if err != nil {
		return nil, fmt.Errorf("db.Query username=%s: %w", username, err)
	}
	defer rows.Close()

	var result []entities.Habit
	for rows.Next() {
		var daoHabit habit

		err := rows.Scan(&daoHabit.id, &daoHabit.description, &daoHabit.goalId, &daoHabit.frequencyType, &daoHabit.timesPerFrequency, &daoHabit.totalTrackingPeriods, &daoHabit.nextCheckDate)
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
    g.total_tracking_periods as total_tracking_periods,
    g.next_check_date as next_check_date,
    g.is_completed as is_completed,
    g.previous_goal_id as previous_goal_id,
    g.start_tracking_at as start_tracking_at
from habits h 
    left join goals g on h.id = g.habit_id and g.is_active = true
where h.username = $1 
  and h.id = $2 and h.is_archived = false;`

	var daoHabit habit

	err := pool.QueryRow(ctx, query, username, habitId).Scan(&daoHabit.id, &daoHabit.description, &daoHabit.goalId, &daoHabit.frequencyType, &daoHabit.timesPerFrequency, &daoHabit.totalTrackingPeriods, &daoHabit.nextCheckDate, &daoHabit.isCompleted, &daoHabit.previousGoalIDs, &daoHabit.startTrackingAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.Habit{}, ErrNotFound
		}
		return entities.Habit{}, fmt.Errorf("db.QueryRow: %w", err)
	}

	return toEntityHabit(daoHabit), nil
}

func (s *Storage) GetHabitGoal(ctx context.Context, habitId int) (entities.Goal, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
select id, 
       frequency_type, 
       times_per_frequency, 
       total_tracking_periods, 
       created_at,
       next_check_date, 
       is_completed,
       is_active, 
       previous_goal_id, 
       start_tracking_at 
from goals 
where habit_id = $1 
  and is_active = true ;
`
	var (
		daoGoal goal
	)

	err := pool.QueryRow(ctx, query, habitId).Scan(&daoGoal.id, &daoGoal.frequencyType, &daoGoal.timesPerFrequency, &daoGoal.totalTrackingPeriods, &daoGoal.createdAt, &daoGoal.nextCheckDate, &daoGoal.isCompleted, &daoGoal.isActive, &daoGoal.previousGoalIDs, &daoGoal.startTrackingAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.Goal{}, ErrNotFound
		}
		return entities.Goal{}, fmt.Errorf("db.QueryRow: %w", err)
	}

	return toEntityGoal(daoGoal), nil
}

func (s *Storage) DeactivateGoalByID(ctx context.Context, id int) error {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `update goals set is_active = false where id = $1;`

	_, err := pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

func (s *Storage) AddProgressLog(ctx context.Context, goalId int, createdAt time.Time) error {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
insert into goal_logs (goal_id, record_created_at) values ($1, $2);
`

	_, err := pool.Exec(ctx, query, goalId, createdAt)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
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
select count(*) 
from goal_logs 
where goal_id = $1 
and record_created_at >= $2 
and record_created_at < $3;
    `

	var count int
	err := pool.QueryRow(ctx, query, goalId, start, end).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("db.QueryRow: %w", err)
	}

	return count, nil
}

func (s *Storage) GetCurrentDayExecutionCount(ctx context.Context, goal entities.Goal, currentTime time.Time) (int, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
select count(*) 
from goal_logs 
where goal_id = $1 
and record_created_at::date = $2::date
`
	var count int
	err := pool.QueryRow(ctx, query, goal.Id, currentTime).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("db.QueryRow: %w", err)
	}

	return count, nil
}

func (s *Storage) GetPreviousPeriodExecutionCount(ctx context.Context, goal entities.Goal, currentTime time.Time) (int, error) {
	// The first period has no previous period
	currentPeriod := goal.GetCurrentPeriod(currentTime)

	if currentPeriod == 0 {
		return 0, nil
	}

	start, end := calculatePeriodRange(goal.StartTrackingAt, goal.FrequencyType, currentPeriod-1)

	var executionCountForPreviousPeriod int

	currentGoalExecutionCount, err := s.getExecutionCountForPeriod(ctx, goal.Id, start, end)
	if err != nil {
		return 0, fmt.Errorf("s.getExecutionCountForPeriod: %w", err)
	}
	executionCountForPreviousPeriod += currentGoalExecutionCount
	for _, previousGoalID := range goal.PreviousGoalIDs {
		specificGoalExecutionCount, err := s.getExecutionCountForPeriod(ctx, previousGoalID, start, end)
		if err != nil {
			return 0, fmt.Errorf("s.getExecutionCountForPeriod: %w", err)
		}

		executionCountForPreviousPeriod += specificGoalExecutionCount
	}

	return executionCountForPreviousPeriod, nil
}

func (s *Storage) GetCurrentPeriodExecutionCount(ctx context.Context, goal entities.Goal, currentTime time.Time) (int, error) {
	currentPeriod := goal.GetCurrentPeriod(currentTime)

	start, end := calculatePeriodRange(goal.StartTrackingAt, goal.FrequencyType, currentPeriod)

	var executionCountForCurrentPeriod int

	// Получаем количество выполнений текущей цели за текущий период
	currentGoalExecutionCount, err := s.getExecutionCountForPeriod(ctx, goal.Id, start, end)
	if err != nil {
		return 0, fmt.Errorf("s.getExecutionCountForPeriod: %w", err)
	}
	executionCountForCurrentPeriod += currentGoalExecutionCount

	// Получаем количество выполнений предыдущих целей за текущий период
	for _, previousGoalID := range goal.PreviousGoalIDs {
		specificGoalExecutionCount, err := s.getExecutionCountForPeriod(ctx, previousGoalID, start, end)
		if err != nil {
			return 0, fmt.Errorf("s.getExecutionCountForPeriod: %w", err)
		}

		executionCountForCurrentPeriod += specificGoalExecutionCount
	}

	// В итоге получается общее количество выполнений целей за текущий период по всем целям, которые были активны
	return executionCountForCurrentPeriod, nil
}

func (s *Storage) GetCurrentProgress(ctx context.Context, goalId int) (entities.Progress, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
select 
    id,
    goal_id,
    total_completed_periods,
    total_skipped_periods, 
    total_completed_times,
    most_longest_streak, 
    current_streak 
from goal_stats 
where goal_id = $1;
`

	var progress entities.Progress
	err := pool.QueryRow(ctx, query, goalId).Scan(&progress.Id, &progress.GoalID, &progress.TotalCompletedPeriods, &progress.TotalSkippedPeriods, &progress.TotalCompletedTimes, &progress.MostLongestStreak, &progress.CurrentStreak)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.Progress{}, ErrNotFound
		}
		return entities.Progress{}, fmt.Errorf("db.QueryRow: %w", err)
	}

	return progress, nil
}

func (s *Storage) GetProgressesForAllGoals(ctx context.Context, username string, progressIDs []int64) ([]entities.Progress, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)
	query := `
select 
    h.username as username,
    gs.id as id,
    gs.goal_id as goal_id,
    gs.total_completed_periods as total_completed_periods,
    gs.total_completed_times as total_completed_times,
    gs.total_skipped_periods as total_skipped_periods,
    gs.most_longest_streak as most_longest_streak,
    gs.current_streak as current_streak
from habits h 
    left join goals g on h.id = g.habit_id and g.is_active = true
	left join goal_stats gs on g.id = gs.goal_id
where h.username = $1
  and h.is_archived = false 
  and g.is_active = true 
  and g.is_completed = false
`

	args := []interface{}{username}

	if len(progressIDs) > 0 {
		query += ` and gs.id = any($2);`
		args = append(args, progressIDs)
	}

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db.Query username=%s: %w", username, err)
	}
	defer rows.Close()

	var result []entities.Progress

	for rows.Next() {
		p := userProgress{}

		err := rows.Scan(&p.username, &p.id, &p.goalId, &p.totalCompletedPeriods, &p.totalCompletedTimes, &p.totalSkippedPeriods, &p.mostLongestStreak, &p.currentStreak)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		result = append(result, toEntityProgress(p))
	}
	return result, nil
}

func (s *Storage) SetGoalCompleted(ctx context.Context, goalId int) error {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
update goals 
set is_completed = true where id = $1;
`
	_, err := pool.Exec(ctx, query, goalId)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

func (s *Storage) GetAllGoalsNeedCheck(ctx context.Context, currentTime time.Time) ([]entities.Goal, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
select id, 
       frequency_type,
       times_per_frequency, 
       total_tracking_periods, 
       created_at, 
       next_check_date, 
       is_completed
from goals 
where is_active = true 
  and is_completed = false
  and next_check_date < $1;
`
	rows, err := pool.Query(ctx, query, currentTime)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var result []entities.Goal

	for rows.Next() {
		var daoGoal goal

		err := rows.Scan(&daoGoal.id, &daoGoal.frequencyType, &daoGoal.timesPerFrequency, &daoGoal.totalTrackingPeriods, &daoGoal.createdAt, &daoGoal.nextCheckDate, &daoGoal.isCompleted)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		result = append(result, toEntityGoal(daoGoal))
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows.Err: %w", rows.Err())
	}

	return result, nil
}

func (s *Storage) SetGoalNextCheckDate(ctx context.Context, goalId int, nextCheckDate time.Time) error {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
update goals set next_check_date = $2 where id = $1;
`
	_, err := pool.Exec(ctx, query, goalId, nextCheckDate)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}
	return nil
}

func (s *Storage) GetAllUserHabitsWithGoals(ctx context.Context, username string) ([]entities.Habit, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
select 
    g.id as goal_id, 
    g.frequency_type, 
    g.times_per_frequency, 
    g.total_tracking_periods, 
    g.next_check_date,
    g.created_at,
    g.previous_goal_id,
    g.start_tracking_at,
    h.id as habit_id, 
    h.description as habit_description
from 
    goals g
join 
    habits h 
on 
    g.habit_id = h.id 
where 
    g.is_active = true 
  	and g.is_completed = false
    and h.username = $1 and h.is_archived = false;
`
	rows, err := pool.Query(ctx, query, username)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var result []entities.Habit

	for rows.Next() {
		var h habit
		err := rows.Scan(&h.goalId, &h.frequencyType, &h.timesPerFrequency, &h.totalTrackingPeriods, &h.nextCheckDate, &h.createdAt, &h.previousGoalIDs, &h.startTrackingAt, &h.id, &h.description)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		result = append(result, toEntityHabit(h))
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows.Err: %w", rows.Err())
	}

	return result, nil
}

func (s *Storage) ArchiveHabitById(ctx context.Context, habitId int) error {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
update habits 
set is_archived = true 
where id = $1
`
	_, err := pool.Exec(ctx, query, habitId)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

func (s *Storage) UpdateHabit(ctx context.Context, habit entities.Habit) error {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
update habits 
set description = $1
where id = $2
`
	_, err := pool.Exec(ctx, query, habit.Description, habit.Id)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

func (s *Storage) AddToken(ctx context.Context, token entities.Token) error {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
insert into tokens (access_token, refresh_token, username, expires_at) 
values ($1, $2, $3, $4) on conflict (username) do update set access_token = $1, refresh_token = $2, expires_at = $4;
`

	_, err := pool.Exec(ctx, query, token.AccessToken, token.RefreshToken, token.Username, token.ExpiresAt)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

func (s *Storage) DeleteTokenByUsername(ctx context.Context, username string) error {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
delete from tokens where username = $1;
	`

	_, err := pool.Exec(ctx, query, username)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}
	return nil
}

func (s *Storage) GetTokenByUsername(ctx context.Context, username string) (entities.Token, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
select access_token, refresh_token, username, expires_at from tokens where username = $1;
`
	var token entities.Token

	err := pool.QueryRow(ctx, query, username).Scan(&token.AccessToken, &token.RefreshToken, &token.Username, &token.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.Token{}, ErrNotFound
		}
		return entities.Token{}, fmt.Errorf("db.QueryRow: %w", err)
	}
	return token, nil
}

func (s *Storage) CreateProgress(ctx context.Context, progress entities.Progress) (int64, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
insert into goal_stats (goal_id, total_completed_periods, total_completed_times, total_skipped_periods, most_longest_streak, current_streak, created_at, updated_at)
values ($1, $2, $3, $4, $5, $6, $7, $8) returning id;
`

	var id int64
	err := pool.QueryRow(ctx, query, progress.GoalID, progress.TotalCompletedPeriods, progress.TotalCompletedTimes, progress.TotalSkippedPeriods, progress.MostLongestStreak, progress.CurrentStreak, progress.CreatedAt, progress.UpdatedAt).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("db.QueryRow: %w", err)
	}

	return id, nil
}

func (s *Storage) UpdateProgressByID(ctx context.Context, progress entities.Progress) error {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
update goal_stats
set total_completed_periods = $2,
    total_completed_times = $3,
    total_skipped_periods = $4,
    most_longest_streak = $5,
    current_streak = $6,
    updated_at = $7
where id = $1;
	`
	_, err := pool.Exec(ctx, query, progress.Id, progress.TotalCompletedPeriods, progress.TotalCompletedTimes, progress.TotalSkippedPeriods, progress.MostLongestStreak, progress.CurrentStreak, progress.UpdatedAt)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

func (s *Storage) GetUserTimeOffset(ctx context.Context, username string) (int, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
select time_offset from users_time where username = $1;
`
	var timeOffset int
	err := pool.QueryRow(ctx, query, username).Scan(&timeOffset)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, fmt.Errorf("db.QueryRow: %w", err)
	}

	return timeOffset, nil
}

func (s *Storage) UpdateUserTimeOffset(ctx context.Context, username string, offset int) error {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
INSERT INTO users_time (username, time_offset)
VALUES ($1, $2)
ON CONFLICT (username)
DO UPDATE SET time_offset = CASE
    WHEN $2 = 0 THEN 0
    ELSE users_time.time_offset + $2
END;
`
	_, err := pool.Exec(ctx, query, username, offset)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

func (s *Storage) UpdateGoal(ctx context.Context, goal entities.Goal) error {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
update goals
set frequency_type = $2,
    times_per_frequency = $3,
    total_tracking_periods = $4
where id = $1;
`
	_, err := pool.Exec(ctx, query, goal.Id, goal.FrequencyType, goal.TimesPerFrequency, goal.TotalTrackingPeriods)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

func (s *Storage) CreateSnapshot(ctx context.Context, snapshot entities.ProgressSnapshot) error {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
	insert into progress_snapshots (username, progress_id, goal_id, created_at)
	values ($1, $2, $3, $4);
`
	_, err := pool.Exec(ctx, query, snapshot.Username, snapshot.ProgressID, snapshot.GoalID, snapshot.CreatedAt)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}
	return nil
}

func (s *Storage) GetMostRecentSnapshot(ctx context.Context, username string, goalID int, currentTime time.Time) (entities.ProgressSnapshot, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
select
	progress_id,
	goal_id,
	created_at,
	username
from progress_snapshots
where username = $1
	  and goal_id = $2
and created_at::date <= $3::date 
order by created_at desc
limit 1
`
	var snapshot entities.ProgressSnapshot
	err := pool.QueryRow(ctx, query, username, goalID, currentTime).Scan(&snapshot.ProgressID, &snapshot.GoalID, &snapshot.CreatedAt, &snapshot.Username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.ProgressSnapshot{}, ErrNotFound
		}
		return entities.ProgressSnapshot{}, fmt.Errorf("db.QueryRow: %w", err)
	}

	return snapshot, nil
}

func (s *Storage) GetCurrentSnapshot(ctx context.Context, username string, goalID int, currentTime time.Time) (entities.ProgressSnapshot, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
select
	progress_id,
	goal_id,
	created_at,
	username
from progress_snapshots
where username = $1
  and goal_id = $2 
  and created_at::date = $3::date
`

	var snapshot entities.ProgressSnapshot
	err := pool.QueryRow(ctx, query, username, goalID, currentTime).Scan(&snapshot.ProgressID, &snapshot.GoalID, &snapshot.CreatedAt, &snapshot.Username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.ProgressSnapshot{}, ErrNotFound
		}
		return entities.ProgressSnapshot{}, fmt.Errorf("db.QueryRow: %w", err)
	}

	return snapshot, nil
}

func (s *Storage) GetProgressByID(ctx context.Context, progressID int64) (entities.Progress, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
select 
	id,
	goal_id,
	total_completed_periods,
	total_completed_times,
	total_skipped_periods,
	most_longest_streak,
	current_streak
from goal_stats
where id = $1
`

	var progress entities.Progress
	err := pool.QueryRow(ctx, query, progressID).Scan(&progress.Id, &progress.GoalID, &progress.TotalCompletedPeriods, &progress.TotalCompletedTimes, &progress.TotalSkippedPeriods, &progress.MostLongestStreak, &progress.CurrentStreak)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.Progress{}, ErrNotFound
		}
		return entities.Progress{}, fmt.Errorf("db.QueryRow: %w", err)
	}

	return progress, nil
}

func (s *Storage) ApplyProgressChangeBySnapshotID(ctx context.Context, snapshotID int64, progressChange entities.ProgressChange) error {

	return nil
}

func (s *Storage) GetFutureSnapshots(ctx context.Context, username string, goalID int, currentTime time.Time) ([]entities.ProgressSnapshot, error) {
	pool := s.queryEngineProvider.GetQueryEngine(ctx)

	query := `
	select 
		progress_id,
		goal_id,
		created_at,
		username
	from progress_snapshots 
	where username = $1 
	  and goal_id = $2 
	  and created_at::date > $3::date
	order by created_at;
	`

	result := make([]entities.ProgressSnapshot, 0)

	rows, err := pool.Query(ctx, query, username, goalID, currentTime)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}

	for rows.Next() {
		var snapshot entities.ProgressSnapshot
		err := rows.Scan(&snapshot.ProgressID, &snapshot.GoalID, &snapshot.CreatedAt, &snapshot.Username)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		result = append(result, snapshot)
	}

	return result, nil
}
