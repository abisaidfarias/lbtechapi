package services

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

// IUserService auth interface
type IUserService interface {
	Create(*request.UserRequest) error
	GetByID(string) (*responses.UserResponse, error)
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

	user := buildNewUser(userRequest)

	err := s.userRepository.Save(user)

	if err != nil {
		return err
	}

	return nil
}

// GetByID gets an user by ID
func (s *userService) GetByID(id string) (*responses.UserResponse, error) {

	user, err := s.userRepository.GetByID(id)

	if err != nil {
		return nil, err
	}
	return buildNewUserResponse(user), nil
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

func buildNewUser(userRequest *request.UserRequest) *models.User {

	hashedPassword := hashPassword(userRequest.Password)

	return &models.User{
		Email:        userRequest.Email,
		PasswordHash: hashedPassword,
		Name:         userRequest.Name,
		LastName:     userRequest.LastName,
		Dni:          userRequest.Dni,
		Phone:        userRequest.Phone,
	}
}

func buildNewUserResponse(user *models.User) *responses.UserResponse {

	return &responses.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		LastName:  user.LastName,
		Dni:       user.Dni,
		Phone:     user.Phone,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func hashPassword(password string) string {

	passwordBytes := []byte(password)

	hash, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)

	if err != nil {
		panic(err)
	}

	return string(hash)
}
