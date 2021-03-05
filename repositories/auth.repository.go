package repositories

import (
	"errors"
	"time"

	"github.com/abisaidfarias/lbtechapi/viewmodels"
	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte("my_secret_key")

// IAuthRepository is the auth repository interface
type IAuthRepository interface {
	SignIn(*viewmodels.AuthCredentials) viewmodels.UserResponse
}

// AuthRepository is the auth repository implementation
type AuthRepository struct {
}

// Claims claims to be added in the json payload
type Claims struct {
	ID string `json:"id"`
	jwt.StandardClaims
}

// SignIn checks database for credentials
func (r *AuthRepository) SignIn(creds *viewmodels.AuthCredentials) (*viewmodels.UserResponse, error) {

	user := viewmodels.UserResponse{
		Email: "this is an email mock",
		ID:    "1",
	}

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
		return nil, errors.New("Unable to sign token")
	}

	user.Token = tokenString

	return &user, nil
}
