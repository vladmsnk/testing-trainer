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
INSERT INTO goals (habit_id, frequency, duration, num_of_periods, start_tracking_at, end_tracking_at)
VALUES ($1, $2, $3, $4, $5, $6);
`

	_, err := tx.Exec(ctx, query, habitID, goal.Frequency, goal.Duration, goal.NumOfPeriods, goal.StartTrackingAt, goal.EndTrackingAt)
	if err != nil {
		return fmt.Errorf("tx.Exec habit_id=%d frequency=%d duration=%d num_of_periods=%d start_tracking_at=%s end_tracking_at=%s: %w", habitID, goal.Frequency, goal.Duration, goal.NumOfPeriods, goal.StartTrackingAt, goal.EndTrackingAt, err)
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
    g.duration as duration, 
    g.frequency as frequency,
    g.num_of_periods as num_of_periods,
    g.start_tracking_at as start_tracking_at,
    g.end_tracking_at as end_tracking_at
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

		err := rows.Scan(&daoHabit.id, &daoHabit.name, &daoHabit.description, &daoHabit.duration, &daoHabit.frequency, &daoHabit.numOfPeriods, &daoHabit.startTrackingAt, &daoHabit.endTrackingAt)
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
    g.duration as duration, 
    g.frequency as frequency,
    g.num_of_periods as num_of_periods,
    g.start_tracking_at as start_tracking_at,
    g.end_tracking_at as end_tracking_at
from habits h 
    left join goals g on h.id = g.habit_id 
where h.username = $1 
  and h.name = $2;`

	var daoHabit habit

	err := s.db.QueryRow(ctx, query, username, habitName).Scan(&daoHabit.id, &daoHabit.name, &daoHabit.description, &daoHabit.duration, &daoHabit.frequency, &daoHabit.numOfPeriods, &daoHabit.startTrackingAt, &daoHabit.endTrackingAt)
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

func (s *Storage) getHabitGoal(ctx context.Context, habitName string) (entities.Goal, error) {
	query := `
select frequency, duration, num_of_periods, start_tracking_at, end_tracking_at from goals where habit_id = (select id from habits where name = $1);
	`
	var daoGoal goal

	err := s.db.QueryRow(ctx, query, habitName).Scan(&daoGoal.frequency, &daoGoal.duration, &daoGoal.numOfPeriods, &daoGoal.startTrackingAt, &daoGoal.endTrackingAt)
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
