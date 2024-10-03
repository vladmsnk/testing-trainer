package entities

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
	AccessToken  string
	RefreshToken string
}
