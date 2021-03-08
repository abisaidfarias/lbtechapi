package services

import (
	"fmt"
	"log"
	"time"

	"github.com/abisaidfarias/lbtechapi/repositories"
	util "github.com/abisaidfarias/lbtechapi/util/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// JWTKey is the key
var JWTKey = []byte("my_secret_key")

// IAuthService auth interface
type IAuthService interface {
	SignIn(*viewmodels.AuthCredentials) (*responses.AuthResponse, error)
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
type AuthClaims struct {
	ID string `json:"id"`
	jwt.StandardClaims
}

// SignIn sign the user in
func (s *authService) SignIn(credentials *viewmodels.AuthCredentials) (*responses.AuthResponse, error) {

	user, err := s.userRepository.GetByEmail(credentials.Email)
	if err != nil {
		return nil, fmt.Errorf("%w", util.ErrorInvalidCredentials)
	}
	err = validateUserCredentials(user.PasswordHash, credentials.Password)
	if err != nil {
		return nil, fmt.Errorf("%w", util.ErrorInvalidCredentials)
	}
	token := generateJWT(user.ID.Hex())
	return &responses.AuthResponse{
		ID:    user.ID,
		Email: user.Email,
		Token: token,
	}, nil
}

// generateJWT create a token
func generateJWT(userID string) string {

	expirationTime := time.Now().Add(999 * time.Minute)

	claims := &AuthClaims{
		ID: userID,
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
func validateUserCredentials(passwordHash string, password string) error {

	passwordMatches := compareHashAndPassword(passwordHash, []byte(password))

	if !passwordMatches {
		return fmt.Errorf("%w", util.ErrorInvalidCredentials)
	}

	return nil
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
