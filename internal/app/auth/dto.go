package auth

import (
	"errors"
	"net/mail"
	"regexp"
)

type RegisterRequest struct {
	Username string `json:"username" example:"john_doe"`
	Email    string `json:"email" example:"john@example.com"`
	Password string `json:"password" example:"securepassword"`
}

func (r *RegisterRequest) Validate() error {
	if len(r.Username) == 0 {
		return errors.New("username is required")
	}

	if err := validateUsername(r.Username); err != nil {
		return err
	}

	if err := validateEmail(r.Email); err != nil {
		return err
	}

	if err := validatePassword(r.Password); err != nil {
		return err
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < 5 || len(password) > 40 {
		return errors.New("password must be between 5 and 40 characters long")
	}

	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)

	if !hasLetter || !hasNumber {
		return errors.New("password must contain at least one letter and one number")
	}

	return nil
}

func validateUsername(username string) error {
	//containes only letters, numbers, and underscores
	ok, _ := regexp.MatchString("^[a-zA-Z0-9_]*$", username)
	if !ok {
		return errors.New("username can only contain letters, numbers, and underscores")
	}
	//between 3 and 20 characters long

	if len(username) < 3 || len(username) > 30 {
		return errors.New("username must be between 3 and 30 characters long")
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
