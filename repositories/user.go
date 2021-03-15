package repositories

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/models"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"gopkg.in/mgo.v2/bson"
)

// IUserRepository is the user repository interface
type IUserRepository interface {
	GetByID(string) (*models.User, error)
	GetByEmail(string) (*models.User, error)
	Save(*models.User) error
	Update(string, *models.User) error
	Delete(string) error
}

type userRepository struct {
	ID    primitive.ObjectID
	Email string
	Token string
}

//NewUserRepository is a constructor
func NewUserRepository() IUserRepository {
	return &userRepository{
		Email: "",
		Token: "",
	}
}

var userCollection = database.GetInstance().Collection("users")

// GetByEmail checks database for credentials
func (r *userRepository) GetByEmail(email string) (*models.User, error) {

	var user models.User

	filter := bson.M{
		"email": email,
	}

	err := userCollection.FindOne(context.TODO(), filter).Decode(&user)

	if err != nil {
		return nil, fmt.Errorf("%w", utils.ErrorInQuery)
	}
	return &user, nil
}

// GetByID checks database for credentials
func (r *userRepository) GetByID(id string) (*models.User, error) {

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}

	var result models.User

	err = userCollection.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&result)

	if err != nil {
		return nil, fmt.Errorf("%w", utils.ErrorInQuery)
	}

	return &result, nil
}

// Save saves the new user
func (r *userRepository) Save(user *models.User) error {

	_, err := userCollection.InsertOne(context.TODO(), user)

	if err != nil {
		var merr mongo.WriteException
		merr = err.(mongo.WriteException)
		errCode := merr.WriteErrors[0].Code
		if errCode == 11000 {
			return fmt.Errorf("%w", utils.ErrorDuplicated)
		}
		return err
	}

	return nil
}

// Update updates the given that
func (r *userRepository) Update(id string, user *models.User) error {

	oid, err := primitive.ObjectIDFromHex(id)

	update := bson.M{
		"$set": user,
	}

	_, err = userCollection.UpdateOne(context.TODO(), bson.M{"_id": oid}, update)

	if err != nil {
		return err
	}

	return nil
}

// Delete deletes the user
func (r *userRepository) Delete(id string) error {
	oid, err := primitive.ObjectIDFromHex(id)

	_, err = userCollection.DeleteOne(context.TODO(), bson.M{"_id": oid})

	if err != nil {
		return err
	}

	return nil
}
