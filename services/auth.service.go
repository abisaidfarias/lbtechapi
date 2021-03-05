package services

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/viewmodels"
)

// IAuthService auth interface
type IAuthService interface {
	SignIn(*viewmodels.AuthCredentials) models.User
}

// AuthService is the auth service
type AuthService struct {
}

// SignIn sign the user in
func (s *AuthService) SignIn(credentials *viewmodels.AuthCredentials) (*viewmodels.UserResponse, error) {
	repository := repositories.AuthRepository{}

	user, err := repository.SignIn(credentials)

	if err != nil {
		return nil, err
	}

	return user, nil
}
