package auth

import "testing_trainer/internal/entities"

func toEntityRegisterUser(req RegisterRequest) entities.RegisterUser {
	return entities.RegisterUser{
		Email:        req.Email,
		Name:         req.Username,
		PasswordHash: req.Password,
	}
}

func toEntityUser(req LoginRequest) entities.User {
	return entities.User{
		Name:     req.Username,
		Password: req.Password,
	}
}

func toLoginResponse(token entities.Token) LoginResponse {
	return LoginResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}
}

func toRefreshResponse(token entities.Token) RefreshTokenResponse {
	return RefreshTokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}
}
