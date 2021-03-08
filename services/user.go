package services

import (
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

// IUserService auth interface
type IUserService interface {
	Create(*request.UserRequest) error
	GetByOID(id string) (*responses.UserResponse, error)
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
	err := s.userRepository.SaveUser(user)
	if err != nil {
		return err
	}
	return nil
}

//GetByOID gets an user by ID
func (s *userService) GetByOID(id string) (*responses.UserResponse, error) {

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}
	user, err := s.userRepository.GetByOID(oid)
	if err != nil {
		return nil, err
	}
	return buildNewUserResponse(user), nil
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
		log.Println(err)
	}

	return string(hash)
}
