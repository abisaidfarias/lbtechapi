package services

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

// IUserService auth interface
type IUserService interface {
	Create(*request.UserRequest) error
	GetByID(string) (*responses.User, error)
	GetByEmail(string) (*responses.User, error)
	Update(string, *models.User) error
	Delete(string) error
}
type userService struct {
	userRepository repositories.IUserRepository
}

// NewUserService is a constructor
func NewUserService(userRepository repositories.IUserRepository) IUserService {
	return &userService{
		userRepository: userRepository,
	}
}

// Create creates and saves the new user
func (s *userService) Create(userRequest *request.UserRequest) error {

	user := mapping.UserRequestToUser(userRequest)

	err := s.userRepository.Create(user)

	if err != nil {
		return err
	}

	return nil
}

// GetByID gets an user by ID
func (s *userService) GetByID(id string) (*responses.User, error) {

	user, err := s.userRepository.GetByID(id)

	if err != nil {
		return nil, err
	}
	return user, nil
}
func (s *userService) GetByEmail(email string) (*responses.User, error) {

	user, err := s.userRepository.GetByEmail(email)

	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) Update(id string, user *models.User) error {

	err := s.userRepository.Update(id, user)

	if err != nil {
		return err
	}
	return nil
}

func (s *userService) Delete(id string) error {

	err := s.userRepository.Delete(id)

	if err != nil {
		return err
	}
	return nil
}
