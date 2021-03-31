package services

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/abisaidfarias/lbtechapi/repositories"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
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
type AuthClaims struct {
	ID         string `json:"id"`
	CompanyID  string
	IsInternal string
	jwt.StandardClaims
}

// SignIn sign the user in
func (s *authService) SignIn(credentials *request.AuthCredentials) (*responses.AuthResponse, error) {

	user, err := s.userRepository.GetByEmail(credentials.Email)
	if err != nil {
		return nil, fmt.Errorf("%w", utils.ErrorInvalidCredentials)
	}
	err = validateUserCredentials(user.PasswordHash, credentials.Password)
	if err != nil {
		return nil, fmt.Errorf("%w", utils.ErrorInvalidCredentials)
	}
	token := generateJWT(user)

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
	}, nil
}

// generateJWT create a token
func generateJWT(user *responses.AuthUser) string {

	var JWTKey = []byte(os.Getenv("SECRET_KEY"))
	// TODO move this into environment variable
	expirationTime := time.Now().Add(1000 * time.Hour)

	claims := &AuthClaims{
		ID:         user.ID.Hex(),
		CompanyID:  user.Company.Hex(),
		IsInternal: strconv.FormatBool(user.IsInternal),
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
