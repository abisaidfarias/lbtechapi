package services

import (
	"fmt"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"
)

// IUserService auth interface
type IUserService interface {
	Create(*request.User) error
	GetByID(string) (*responses.User, error)
	Get(string) ([]*bson.M, error)
	GetByEmail(string) (*responses.AuthUser, error)
	GetProfileByID(string) (*responses.Profile, error)
	Update(string, *models.User) error
	Delete(string) error
	ChangePassword(string, request.ChangePassword) error
	Upgrade(*request.User) error
	GetInternalUser() ([]*responses.User, error)
	GetUsersByCompany(string) ([]*responses.User, error)
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
func (s *userService) Create(userRequest *request.User) error {

	userLogged, err := s.userRepository.GetByID(userRequest.UserID)

	if err != nil {
		return err
	}
	if userRequest.IsInternal {
		userRequest.Company = userLogged.Company.Hex()
	}

	user := mapping.UserRequestToUser(userRequest)

	err = s.userRepository.Create(user)

	if err != nil {
		return err
	}

	return nil
}
func (s *userService) Upgrade(userRequest *request.User) error {

	user := mapping.UserRequestToUser(userRequest)

	err := s.userRepository.Create(user)

	if err != nil {
		return err
	}

	return nil
}
func (s *userService) Get(userID string) ([]*bson.M, error) {

	user, err := s.userRepository.GetByID(userID)

	if err != nil {
		return nil, err
	}
	if user.IsInternal {
		users, err := s.userRepository.Get()
		if err != nil {
			return nil, err
		}
		return users, err
	}
	users, err := s.userRepository.GetByCompany(user.Company)
	if err != nil {
		return nil, err
	}
	return users, err
}

// GetByID gets an user by ID
func (s *userService) GetByID(id string) (*responses.User, error) {

	user, err := s.userRepository.GetByID(id)

	if err != nil {
		return nil, err
	}
	return user, nil
}
func (s *userService) GetByEmail(email string) (*responses.AuthUser, error) {

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
func (s *userService) GetProfileByID(id string) (*responses.Profile, error) {

	oid, _ := primitive.ObjectIDFromHex(id)
	profile, err := s.userRepository.GetProfileByID(oid)

	if err != nil {
		return nil, err
	}
	return profile, nil
}
func (s *userService) ChangePassword(email string, changePassword request.ChangePassword) error {

	user, err := s.userRepository.GetByEmail(email)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("%w", utils.ErrorResourceNotFound)
	}
	err = functions.ValidateUserCredentials(user.PasswordHash, changePassword.OldPassword)
	if err != nil {
		return fmt.Errorf("%w", utils.ErrorInvalidCredentials)
	}
	hashedPassword := functions.HashPassword(changePassword.NewPassword)

	err = s.userRepository.ChangePassword(hashedPassword, email)

	if err != nil {
		return err
	}
	return nil
}
func (s *userService) GetInternalUser() ([]*responses.User, error) {

	users, err := s.userRepository.GetInternalUser()

	if err != nil {
		return nil, err
	}
	return users, nil
}
func (s *userService) GetUsersByCompany(id string) ([]*responses.User, error) {

	company, _ := primitive.ObjectIDFromHex(id)
	users, err := s.userRepository.GetUsersByCompany(company)

	if err != nil {
		return nil, err
	}
	return users, nil
}
