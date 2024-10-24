package habit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"testing_trainer/internal/entities"
	"testing_trainer/internal/storage"
	"testing_trainer/internal/usecase/user"
)

func TestCreateHabit(t *testing.T) {

	initFunc := func(t *testing.T) (*MockStorage, *MockUserUseCase) {
		mockStorage := NewMockStorage(t)
		mockUserUc := NewMockUserUseCase(t)

		return mockStorage, mockUserUc
	}

	var (
		ctx             = context.Background()
		username        = "username"
		expectedHabitID = int64(1)
		habitId         = "1"
	)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mockStorage, mockUserUc := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{Name: username}, nil)
		mockStorage.On("CreateHabit", ctx, username, entities.Habit{Id: habitId}).Return(expectedHabitID, nil)
		habituc := New(mockStorage, mockUserUc)

		habitID, err := habituc.CreateHabit(ctx, username, entities.Habit{Id: habitId})
		require.Nil(t, err, "unexpected error")
		require.Equal(t, expectedHabitID, habitID)
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		mockStorage, mockUserUc := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, user.ErrUserNotFound)
		habituc := New(mockStorage, mockUserUc)

		habitID, err := habituc.CreateHabit(ctx, username, entities.Habit{Id: habitId})
		require.ErrorIs(t, err, user.ErrUserNotFound, "unexpected error")
		require.Zero(t, habitID)
	})
}

func TestListUserHabits(t *testing.T) {
	var (
		ctx            = context.Background()
		username       = "username"
		expectedHabits = []entities.Habit{{Id: "1"}, {Id: "2"}}
	)

	initFunc := func(t *testing.T) (*MockStorage, *MockUserUseCase) {
		mockStorage := NewMockStorage(t)
		mockUserUc := NewMockUserUseCase(t)

		return mockStorage, mockUserUc
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mockStorage, mockUserUc := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{Name: username}, nil)
		mockStorage.On("ListUserHabits", ctx, username).Return(expectedHabits, nil)
		habituc := New(mockStorage, mockUserUc)

		habits, err := habituc.ListUserHabits(ctx, username)
		require.Nil(t, err, "unexpected error")
		require.Equal(t, expectedHabits, habits)
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		mockStorage, mockUserUc := initFunc(t)

		mockUserUc.On("GetUserByUsername", ctx, username).Return(entities.User{}, storage.ErrNotFound)
		habituc := New(mockStorage, mockUserUc)

		habits, err := habituc.ListUserHabits(ctx, username)
		require.ErrorIs(t, err, ErrUserNotFound, "unexpected error")
		require.Nil(t, habits)
	})
}
