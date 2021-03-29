package repositories

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/database/queries"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

// IUserRepository is the user repository interface
type IUserRepository interface {
	GetByID(string) (*responses.User, error)
	GetByEmail(string) (*responses.User, error)
	GetByCompany(primitive.ObjectID) ([]*responses.User, error)
	GetProfileByID(primitive.ObjectID) (*responses.Profile, error)
	Get() ([]*responses.User, error)
	Create(*models.User) error
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
func (r *userRepository) GetByEmail(email string) (*responses.User, error) {

	var user responses.User

	err := userCollection.FindOne(context.TODO(), queries.GetUserByEmail(email)).Decode(&user)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByID checks database for credentials
func (r *userRepository) GetByID(id string) (*responses.User, error) {

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}

	var result responses.User

	err = userCollection.FindOne(context.TODO(), queries.GetUserById(oid)).Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}
func (r *userRepository) GetByCompany(companyID primitive.ObjectID) ([]*responses.User, error) {

	var users []*responses.User = []*responses.User{}
	cursor, err := userCollection.Find(context.TODO(), queries.GetUserByCompany(companyID))
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.TODO(), &users); err != nil {
		return nil, err
	}
	return users, nil
}
func (r *userRepository) Get() ([]*responses.User, error) {

	var users []*responses.User = []*responses.User{}
	cursor, err := userCollection.Find(context.TODO(), primitive.M{})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.TODO(), &users); err != nil {
		return nil, err
	}
	return users, nil
}

// Save saves the new user
func (r *userRepository) Create(user *models.User) error {

	_, err := userCollection.InsertOne(context.TODO(), user)

	if err != nil {
		return err
	}

	return nil
}

// Update updates the given that
func (r *userRepository) Update(id string, user *models.User) error {

	oid, err := primitive.ObjectIDFromHex(id)

	filter, update := queries.UpdateUser(user, oid)

	_, err = userCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	return nil
}

// Delete deletes the user
func (r *userRepository) Delete(id string) error {
	oid, err := primitive.ObjectIDFromHex(id)

	_, err = userCollection.DeleteOne(context.TODO(), queries.DeleteUser(oid))

	if err != nil {
		return err
	}

	return nil
}
func (r *userRepository) GetProfileByID(oid primitive.ObjectID) (*responses.Profile, error) {

	var user responses.User
	err := userCollection.FindOne(context.TODO(), queries.GetUserById(oid)).Decode(&user)

	if err != nil {
		return nil, err
	}
	var profile *responses.Profile
	err = profileCollection.FindOne(context.TODO(), queries.GeProfileById(user.Profile)).Decode(&profile)

	return profile, nil
}
