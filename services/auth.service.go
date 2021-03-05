package services

import (
	"fmt"
	"log"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/viewmodels"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

var (
	repository repositories.AuthRepository
)

// TODO move into correct package

// JWTKey is the key
var JWTKey = []byte("my_secret_key")

// IAuthService auth interface
type IAuthService interface {
	SignIn(*viewmodels.AuthCredentials) models.User
}

// AuthService is the auth service
type AuthService struct {
}

// AuthClaims claims to be added in the json payload
type AuthClaims struct {
	ID string `json:"id"`
	jwt.StandardClaims
}

// SignIn sign the user in
func (s *AuthService) SignIn(credentials *viewmodels.AuthCredentials) (*viewmodels.UserResponse, error) {

	user, err := repository.GetUserByEmail(credentials)

	if err != nil {
		return nil, fmt.Errorf("%w", models.ErrorInvalidCredentials)
	}

	err = validateUserCredentials(user, credentials)

	// invalid credentials
	if err != nil {
		return nil, fmt.Errorf("%w", models.ErrorInvalidCredentials)
	}

	token := generateJWT(user)

	// build user response
	userResponse := viewmodels.UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Token: token,
	}

	return &userResponse, nil
}

// SignUp creates and saves the new user
func (s *AuthService) SignUp(credentials *viewmodels.AuthCredentials) error {

	user, err := buildNewUser(credentials)

	err = repository.SaveUSer(user)

	// error saving the new user
	if err != nil {
		return err
	}

	return nil
}

// GetUserByID gets an user byID
func (s *AuthService) GetUserByID(id string) (*models.User, error) {

	user, err := repository.GetUserByID(id)

	return user, err
}

func generateJWT(user *models.User) string {

	expirationTime := time.Now().Add(5 * time.Minute)

	claims := &AuthClaims{
		ID: user.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(JWTKey)

	if err != nil {
		log.Println(err)
	}

	return tokenString
}

func validateUserCredentials(user *models.User, credentials *viewmodels.AuthCredentials) error {

	passwordMatches := compareHashAndPassword(user.PasswordHash, []byte(credentials.Password))

	if !passwordMatches {
		return fmt.Errorf("%w", models.ErrorInvalidCredentials)
	}

	return nil
}

func buildNewUser(credentials *viewmodels.AuthCredentials) (*models.User, error) {

	hashedPassword := hashPassword(credentials.Password)

	user := models.User{
		Email:        credentials.Email,
		PasswordHash: hashedPassword,
	}

	return &user, nil
}

func hashPassword(password string) string {

	passwordBytes := []byte(password)

	hash, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
	}

	return string(hash)
}

func compareHashAndPassword(hashedPassword string, incomingPassword []byte) bool {

	byteHash := []byte(hashedPassword)
	err := bcrypt.CompareHashAndPassword(byteHash, incomingPassword)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}
