package repositories

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/models"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"gopkg.in/mgo.v2/bson"
)

// IUserRepository is the user repository interface
type IUserRepository interface {
	GetByEmail(email string) (*models.User, error)
	GetByOID(oid primitive.ObjectID) (*models.User, error)
	SaveUser(user *models.User) error
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

var userCollection = database.Conn().Collection("users")

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

// GetByOID checks database for credentials
func (r *userRepository) GetByOID(oid primitive.ObjectID) (*models.User, error) {

	var result models.User

	err := userCollection.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("%w", utils.ErrorInQuery)
	}

	return &result, nil
}

// SaveUSer saves the new user
func (r *userRepository) SaveUser(user *models.User) error {

	_, err := userCollection.InsertOne(context.TODO(), user)

	if err != nil {
		var merr mongo.WriteException
		merr = err.(mongo.WriteException)
		errCode := merr.WriteErrors[0].Code
		log.Println(err.Error())
		if errCode == 11000 {
			return fmt.Errorf("%w", utils.ErrorDuplicated)
		}
		return err
	}

	return nil
}
