package auth

import (
	"errors"
	"net/mail"
)

type RegisterRequest struct {
	Username string `json:"username" example:"john_doe"`
	Email    string `json:"email" example:"john@example.com"`
	Password string `json:"password" example:"securepassword"`
}

func (r *RegisterRequest) Validate() error {
	if len(r.Username) == 0 {
		return errors.New("username cannot be empty")
	}

	if err := validateEmail(r.Email); err != nil {
		return err
	}

	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	return nil
}

// validateEmail checks if the provided email address is valid
func validateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return errors.New("invalid email address")
	}
	return nil
}

type LoginRequest struct {
	Username string `json:"username" example:"john_doe"`
	Password string `json:"password" example:"securepassword"`
}

func (r *LoginRequest) Validate() error {
	if len(r.Username) == 0 {
		return errors.New("username cannot be empty")
	}

	return nil
}

type RegisterResponse struct{}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	AccessToken string `json:"access_token"`
}

type LogoutResponse struct{}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" example:"refresh_token"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
