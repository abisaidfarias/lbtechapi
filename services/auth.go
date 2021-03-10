package services

import (
	"fmt"
	"os"
	"time"

	"github.com/abisaidfarias/lbtechapi/repositories"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

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
	// TODO add companyId to authClaims
	jwt.StandardClaims
}

// SignIn sign the user in
func (s *authService) SignIn(credentials *viewmodels.AuthCredentials) (*responses.AuthResponse, error) {

	user, err := s.userRepository.GetByEmail(credentials.Email)
	if err != nil {
		return nil, fmt.Errorf("%w", utils.ErrorInvalidCredentials)
	}
	err = validateUserCredentials(user.PasswordHash, credentials.Password)
	if err != nil {
		return nil, fmt.Errorf("%w", utils.ErrorInvalidCredentials)
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

	var JWTKey = []byte(os.Getenv("SECRET_KEY"))
	// TODO move this into environment variable
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
		panic(err)
	}

	return tokenString
}
func validateUserCredentials(passwordHash string, password string) error {

	passwordMatches := compareHashAndPassword(passwordHash, []byte(password))

	if !passwordMatches {
		return fmt.Errorf("%w", utils.ErrorInvalidCredentials)
	}

	return nil
}

func compareHashAndPassword(hashedPassword string, incomingPassword []byte) bool {

	byteHash := []byte(hashedPassword)
	err := bcrypt.CompareHashAndPassword(byteHash, incomingPassword)
	if err != nil {
		return false
	}
	return true
}
