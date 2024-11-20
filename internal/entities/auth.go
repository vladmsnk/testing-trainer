package entities

import "time"

type RegisterUser struct {
	Name         string
	Email        string
	PasswordHash string
}

type User struct {
	Name     string
	Email    string
	Password string
}

type Token struct {
	Username     string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}
