package services

import (
	"github.com/abisaidfarias/lbtechapi/models"
	repository "github.com/abisaidfarias/lbtechapi/repositories"
)

// IAuthService auth interface
type IAuthService interface {
	SignIn(*models.AuthCredentials) models.User
}

// AuthService is the auth service
type AuthService struct {
}

// SignIn sign the user in
func (s *AuthService) SignIn(credentials *models.AuthCredentials) (*models.UserResponse, error) {
	repository := repository.AuthRepository{}

	user, err := repository.SignIn(credentials)

	if err != nil {
		return nil, err
	}

	return user, nil
}
