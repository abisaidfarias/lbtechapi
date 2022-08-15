package services

import (
	"fmt"

	"github.com/abisaidfarias/lbtechapi/repositories"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

// IAuthService auth interface
type IAuthService interface {
	SignIn(*request.AuthCredentials) (*responses.AuthResponse, error)
}

type authService struct {
	userRepository repositories.IUserRepository
}

// NewAuthService is a constructor
func NewAuthService(userRepository repositories.IUserRepository) IAuthService {
	return &authService{
		userRepository: userRepository,
	}
}

// AuthClaims claims to be added in the json payload

// SignIn sign the user in
func (s *authService) SignIn(credentials *request.AuthCredentials) (*responses.AuthResponse, error) {

	user, err := s.userRepository.GetByEmail(credentials.Email)
	if err != nil {
		return nil, fmt.Errorf("%w", utils.ErrorInvalidCredentials)
	}
	err = functions.ValidateUserCredentials(user.PasswordHash, credentials.Password)
	if err != nil {
		return nil, fmt.Errorf("%w", utils.ErrorInvalidCredentials)
	}
	token := functions.GenerateJWT(user)

	profile, err := s.userRepository.GetProfileByID(user.ID)

	if err != nil {
		return nil, fmt.Errorf("%w", utils.ErrorInQuery)
	}

	return &responses.AuthResponse{
		ID:       user.ID,
		Email:    user.Email,
		Profile:  *profile,
		Token:    token,
		Name:     user.Name,
		LastName: user.LastName,
		Company:  user.Company,
	}, nil
}
