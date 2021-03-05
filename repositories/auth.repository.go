package repositories

import (
	"context"
	"fmt"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels"
	"gopkg.in/mgo.v2/bson"
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

	var result models.User

	filter := bson.M{
		"email": credentials.Email,
	}

	err := userCollection.FindOne(context.TODO(), filter).Decode(&result)

	if err != nil {
		return nil, fmt.Errorf("%w", models.ErrorInvalidCredentials)
	}

	return &result, nil
}

// SaveUSer saves the new user
func (r *AuthRepository) SaveUSer(user *models.User) error {

	_, err := userCollection.InsertOne(context.TODO(), user)

	// error registering user
	if err != nil {
		return err
	}

	return nil
}
