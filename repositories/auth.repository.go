package repositories

import (
	"context"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels"
)

// IAuthRepository is the auth repository interface
type IAuthRepository interface {
	getUserByEmail(*viewmodels.AuthCredentials) viewmodels.UserResponse
}

// AuthRepository is the auth repository implementation
type AuthRepository struct {
}

var userCollection = database.Conn().Collection("users")

// GetUserByEmail checks database for credentials
func (r *AuthRepository) GetUserByEmail(credentials *viewmodels.AuthCredentials) (*models.User, error) {

	user := models.User{
		Email:        "this is an email mock",
		ID:           "1",
		PasswordHash: "password",
	}

	return &user, nil
}

// SaveUSer saves the new user
func (r *AuthRepository) SaveUSer(user *models.User) error {

	_, err := userCollection.InsertOne(context.TODO(), user)

	// error registering user
	if err != nil {
		return err
	}

	// TODO change here to return user
	return nil
}
