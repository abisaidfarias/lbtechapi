package repositories

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/database/queries"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

type IProfileRepository interface {
	Create(*models.Profile) error
	Get() ([]*bson.M, error)
	GetById(string) (*responses.Profile, error)
	GetByCompany(primitive.ObjectID) ([]*bson.M, error)
	Update(string, *models.Profile) error
	Delete(string) (error, bool)
}

type profileRepository struct {
}

func NewProfileRepository() IProfileRepository {
	return &profileRepository{}
}

var profileCollection = database.GetInstance().Collection("profiles")

// Create a new tet case
func (r *profileRepository) Create(profile *models.Profile) error {

	_, err := profileCollection.InsertOne(context.TODO(), profile)

	if err != nil {
		return err
	}
	return nil
}

// Get returns a list of all test cases
func (r *profileRepository) Get() ([]*bson.M, error) {

	cursor, err := profileCollection.Find(context.TODO(), bson.M{})

	if err != nil {
		panic(err)
	}
	var profiles []*bson.M = []*bson.M{}
	for cursor.Next(context.TODO()) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, &result)

	}
	cursor.Close(context.TODO())
	return profiles, nil
}

func (r *profileRepository) GetById(id string) (*responses.Profile, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}

	var result responses.Profile

	err = profileCollection.FindOne(context.TODO(), queries.GetProfileById(oid)).Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}
func (r *profileRepository) Update(id string, profile *models.Profile) error {

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter, update := queries.UpdateProfile(profile, oid)

	_, err = profileCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	return nil
}
func (r *profileRepository) Delete(id string) (error, bool) {
	oid, err := primitive.ObjectIDFromHex(id)

	cursor, err := userCollection.Find(context.TODO(), queries.GetUsersProfileId(id))
	if err != nil {
		return err, false
	}

	var users []*responses.User
	if err = cursor.All(context.TODO(), &users); err != nil {
		return err, false
	}
	if len(users) > 0 {
		return nil, false
	}
	_, err = profileCollection.DeleteOne(context.TODO(), queries.DeleteProfile(oid))

	if err != nil {
		return err, false
	}

	return nil, true
}
func (r *profileRepository) GetByCompany(companyID primitive.ObjectID) ([]*bson.M, error) {

	cursor, err := profileCollection.Find(context.TODO(), queries.GetProfileByCompany(companyID))

	if err != nil {

		panic(err)
	}
	var profiles []*bson.M = []*bson.M{}
	for cursor.Next(context.TODO()) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, &result)

	}
	cursor.Close(context.TODO())
	return profiles, nil
}
