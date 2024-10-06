package user

import (
	"context"
	"errors"
	"fmt"
	"testing_trainer/internal/entities"
	"testing_trainer/internal/storage"
	utils "testing_trainer/utils/password"
	"testing_trainer/utils/token"
)

var (
	ErrUserNotFound      = fmt.Errorf("user not found")
	ErrUserAlreadyExists = fmt.Errorf("user already exists")
)

type UseCase interface {
	GetUserByUsername(ctx context.Context, username string) (entities.User, error)
	RegisterUser(ctx context.Context, user entities.RegisterUser) error
}

type Storage interface {
	GetUserByUsername(ctx context.Context, username string) (entities.User, error)
	CreateUser(ctx context.Context, user entities.RegisterUser) error
}

type Implementation struct {
	storage Storage
}

func New(storage Storage) *Implementation {
	return &Implementation{storage: storage}
}

func (i *Implementation) GetUserByUsername(ctx context.Context, username string) (entities.User, error) {
	user, err := i.storage.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return entities.User{}, ErrUserNotFound
		}
		return entities.User{}, fmt.Errorf("i.storage.GetUserByUsername: %w", err)
	}

	return user, nil
}

func (i *Implementation) RegisterUser(ctx context.Context, user entities.RegisterUser) error {
	_, err := i.storage.GetUserByUsername(ctx, user.Name)
	if err != nil {
		if !errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("i.storage.GetUserByUsername: %w", err)
		}
	} else {
		return ErrUserAlreadyExists
	}

	user.PasswordHash, err = utils.HashPassword(user.PasswordHash)
	if err != nil {
		return fmt.Errorf("password.HashPassword: %w", err)
	}

	err = i.storage.CreateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("i.storage.CreateUser: %w", err)
	}
	return nil
}

func (i *Implementation) Login(ctx context.Context, user entities.User) (entities.Token, error) {
	userFromStorage, err := i.storage.GetUserByUsername(ctx, user.Name)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return entities.Token{}, ErrUserNotFound
		}
		return entities.Token{}, fmt.Errorf("i.storage.GetUserByUsername: %w", err)
	}

	err = utils.CheckPassword(userFromStorage.Password, user.Password)
	if err != nil {
		return entities.Token{}, fmt.Errorf("password.CheckPasswordHash: %w", err)
	}

	accessToken, refreshToken, err := token.GenerateTokens(userFromStorage.Name)
	if err != nil {
		return entities.Token{}, fmt.Errorf("token.GenerateTokens: %w", err)
	}

	return entities.Token{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
