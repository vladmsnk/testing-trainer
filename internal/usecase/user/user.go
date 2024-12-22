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
	ErrInvalidPassword   = fmt.Errorf("invalid password")
	ErrTokenNotFound     = fmt.Errorf("token not found")
	ErrInvalidToken      = fmt.Errorf("invalid token")
)

type UseCase interface {
	GetUserByUsername(ctx context.Context, username string) (entities.User, error)
	RegisterUser(ctx context.Context, user entities.RegisterUser) error
	Login(ctx context.Context, user entities.User) (entities.Token, error)
	Logout(ctx context.Context, token entities.Token) error
	GetToken(ctx context.Context, username string) (entities.Token, error)
	RefreshToken(ctx context.Context, tkn entities.Token) (entities.Token, error)
}

type Storage interface {
	GetUserByEmail(ctx context.Context, email string) (entities.User, error)
	GetUserByUsername(ctx context.Context, username string) (entities.User, error)
	CreateUser(ctx context.Context, user entities.RegisterUser) error
	AddToken(ctx context.Context, token entities.Token) error
	DeleteTokenByUsername(ctx context.Context, username string) error
	GetTokenByUsername(ctx context.Context, username string) (entities.Token, error)
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
	_, err = i.storage.GetUserByEmail(ctx, user.Email)
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
		return entities.Token{}, ErrInvalidPassword
	}

	accessToken, refreshToken, err := token.GenerateTokens(userFromStorage.Name)
	if err != nil {
		return entities.Token{}, fmt.Errorf("token.GenerateTokens: %w", err)
	}

	tkn := entities.Token{Username: userFromStorage.Name, AccessToken: accessToken, RefreshToken: refreshToken}

	err = i.storage.AddToken(ctx, tkn)
	if err != nil {
		return entities.Token{}, fmt.Errorf("i.storage.AddToken: %w", err)
	}

	return entities.Token{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (i *Implementation) Logout(ctx context.Context, tkn entities.Token) error {

	err := token.ValidateToken(tkn.AccessToken, token.KeyEnvApiSecret)
	if err != nil {
		return ErrInvalidToken
	}

	userNameFromToken, err := token.ExtractUserNameFromToken(tkn.AccessToken, token.KeyEnvApiSecret)
	if err != nil {
		return ErrInvalidToken
	}

	tokenFromStorage, err := i.storage.GetTokenByUsername(ctx, userNameFromToken)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrTokenNotFound
		}
		return fmt.Errorf("i.storage.GetTokenByUsername: %w", err)
	}

	if tkn.AccessToken != tokenFromStorage.AccessToken {
		return ErrInvalidToken
	}

	_, err = i.storage.GetUserByUsername(ctx, userNameFromToken)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("i.storage.GetUserByUsername: %w", err)
	}

	err = i.storage.DeleteTokenByUsername(ctx, userNameFromToken)
	if err != nil {
		return fmt.Errorf("i.storage.DeleteTokenByUsername: %w", err)
	}

	return nil
}

func (i *Implementation) RefreshToken(ctx context.Context, tkn entities.Token) (entities.Token, error) {
	err := token.ValidateToken(tkn.RefreshToken, token.KeyEnvRefreshSecret)
	if err != nil {
		return entities.Token{}, err
	}

	userNameFromToken, err := token.ExtractUserNameFromToken(tkn.RefreshToken, token.KeyEnvRefreshSecret)
	if err != nil {
		return entities.Token{}, fmt.Errorf("token.ExtractUserNameFromToken: %w", err)
	}

	tokenFromStorage, err := i.storage.GetTokenByUsername(ctx, userNameFromToken)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return entities.Token{}, ErrTokenNotFound
		}
		return entities.Token{}, fmt.Errorf("i.storage.GetTokenByUsername: %w", err)
	}

	if tkn.RefreshToken != tokenFromStorage.RefreshToken {
		return entities.Token{}, fmt.Errorf("invalid refresh token")
	}

	_, err = i.storage.GetUserByUsername(ctx, userNameFromToken)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return entities.Token{}, ErrUserNotFound
		}
		return entities.Token{}, fmt.Errorf("i.storage.GetUserByUsername: %w", err)
	}

	accessToken, refreshToken, err := token.GenerateTokens(userNameFromToken)
	if err != nil {
		return entities.Token{}, fmt.Errorf("token.GenerateTokens: %w", err)
	}

	updatedToken := entities.Token{Username: userNameFromToken, AccessToken: accessToken, RefreshToken: refreshToken}

	err = i.storage.DeleteTokenByUsername(ctx, userNameFromToken)
	if err != nil {
		return entities.Token{}, fmt.Errorf("i.storage.DeleteTokenByUsername: %w", err)
	}

	err = i.storage.AddToken(ctx, updatedToken)
	if err != nil {
		return entities.Token{}, fmt.Errorf("i.storage.AddToken: %w", err)
	}

	return updatedToken, nil
}

func (i *Implementation) GetToken(ctx context.Context, username string) (entities.Token, error) {
	_, err := i.storage.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return entities.Token{}, ErrUserNotFound
		}
		return entities.Token{}, fmt.Errorf("i.storage.GetUserByUsername: %w", err)
	}

	tkn, err := i.storage.GetTokenByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return entities.Token{}, ErrTokenNotFound
		}
		return entities.Token{}, fmt.Errorf("i.storage.GetTokenByUsername: %w", err)
	}

	return tkn, err
}
