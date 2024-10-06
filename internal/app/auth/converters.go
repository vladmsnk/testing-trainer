package auth

import "testing_trainer/internal/entities"

func toEntityRegisterUser(req RegisterRequest) entities.RegisterUser {
	return entities.RegisterUser{
		Email:        req.Email,
		Name:         req.Username,
		PasswordHash: req.Password,
	}
}
