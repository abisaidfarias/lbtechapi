package services

import (
	"fmt"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/viewmodels"
	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte("my_secret_key")

// IAuthService auth interface
type IAuthService interface {
	SignIn(*viewmodels.AuthCredentials) models.User
}

// AuthService is the auth service
type AuthService struct {
}

// Claims claims to be added in the json payload
type Claims struct {
	ID string `json:"id"`
	jwt.StandardClaims
}

// SignIn sign the user in
func (s *AuthService) SignIn(credentials *viewmodels.AuthCredentials) (*viewmodels.UserResponse, error) {
	repository := repositories.AuthRepository{}

	user, err := repository.GetUserByEmail(credentials)

	err = validateUserCredentials(user, credentials)

	// invalid credentials
	if err != nil {
		return nil, fmt.Errorf("%w", models.ErrorInvalidCredentials)
	}

	token, err := generateJWT(user)

	// error singing the token
	if err != nil {
		return nil, err
	}

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
	repository := repositories.AuthRepository{}

	user, err := buildNewUser(credentials)

	err = repository.SaveUSer(user)

	// error saving the new user
	if err != nil {
		return err
	}

	return nil
}

func generateJWT(user *models.User) (string, error) {

	expirationTime := time.Now().Add(5 * time.Minute)

	claims := &Claims{
		ID: user.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		return "", fmt.Errorf("%f", models.ErrorUnableToSignToken)
	}

	return tokenString, nil
}

func validateUserCredentials(user *models.User, credentials *viewmodels.AuthCredentials) error {

	// TODO hash incoming password
	hashedPassword := credentials.Password

	if user.PasswordHash != hashedPassword {
		return fmt.Errorf("%w", models.ErrorInvalidCredentials)
	}

	return nil
}

func buildNewUser(credentials *viewmodels.AuthCredentials) (*models.User, error) {

	// TODO generate salt and hash password for the user

	user := models.User{
		Email:        "this is an email mock",
		ID:           "1",
		PasswordHash: "passwordHash",
	}

	return &user, nil
}
