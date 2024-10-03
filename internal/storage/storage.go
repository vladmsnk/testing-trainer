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
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, fmt.Errorf("tx.Exec user_id=%s habit=%s description=%s: %w", username, habit.Name, habit.Description, err)
	}

	return habitID, nil
}

func (s *Storage) createGoalTx(ctx context.Context, tx pgx.Tx, habitID int64, goal *entities.Goal) error {
	if goal == nil {
		return nil
	}

	query := `
INSERT INTO goals (habit_id, frequency, duration, num_of_periods, start_tracking_at)
VALUES ($1, $2, $3, $4, $5);
`

	_, err := tx.Exec(ctx, query, habitID, goal.Frequency, goal.Duration, goal.NumOfPeriods, goal.StartTrackingAt)
	if err != nil {
		return fmt.Errorf("tx.Exec habit_id=%d frequency=%d duration=%d num_of_periods=%d start_tracking_at=%s: %w", habitID, goal.Frequency, goal.Duration, goal.NumOfPeriods, goal.StartTrackingAt, err)
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

func (s *Storage) CreateUser(ctx context.Context, user entities.User) error {
	query := `
insert into users (username, email, password_hash) 
values ($1, $2, $3);
`
	_, err := s.db.Exec(ctx, query, user.Name, user.Email, user.Password)
	if err != nil {
		return fmt.Errorf("s.db.Exec: %w", err)
	}

	return nil
}
