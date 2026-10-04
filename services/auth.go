package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

// IAuthService auth interface
type IAuthService interface {
	SignIn(*request.AuthCredentials) (*responses.AuthResponse, error)
	ForgotPassword(*request.ForgotPassword) error
	SetPassword(*request.SetPassword) error
	ValidatePasswordToken(string) (*responses.PasswordTokenValidate, error)
}

type authService struct {
	userRepository       repositories.IUserRepository
	passwordTokenService IPasswordTokenService
}

// NewAuthService is a constructor
func NewAuthService(
	userRepository repositories.IUserRepository,
	passwordTokenService IPasswordTokenService,
) IAuthService {
	return &authService{
		userRepository:       userRepository,
		passwordTokenService: passwordTokenService,
	}
}

// SignIn sign the user in
func (s *authService) SignIn(credentials *request.AuthCredentials) (*responses.AuthResponse, error) {

	user, err := s.userRepository.GetByEmail(credentials.Email)
	if err != nil {
		return nil, fmt.Errorf("%w", utils.ErrorInvalidCredentials)
	}
	if functions.EffectiveAccountStatus(user) == models.AccountStatusPending {
		return nil, fmt.Errorf("%w", utils.ErrorPasswordNotSet)
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
		Brands:   user.Brands,
	}, nil
}

func (s *authService) ForgotPassword(req *request.ForgotPassword) error {
	s.sendForgotPasswordIfExists(req.Email)
	return nil
}

func (s *authService) SetPassword(req *request.SetPassword) error {
	return s.passwordTokenService.Consume(strings.TrimSpace(req.Token), req.NewPassword)
}

func (s *authService) ValidatePasswordToken(raw string) (*responses.PasswordTokenValidate, error) {
	token, err := s.passwordTokenService.ValidateRawToken(strings.TrimSpace(raw))
	if err != nil {
		return &responses.PasswordTokenValidate{Valid: false}, nil
	}
	return &responses.PasswordTokenValidate{
		Valid:     true,
		Purpose:   token.Purpose,
		ExpiresAt: token.ExpiresAt.UTC().Format(time.RFC3339),
	}, nil
}
