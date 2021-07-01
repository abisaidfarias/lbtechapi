package repositories

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

type ILocationRepository interface {
	Create(*models.Location) (*primitive.ObjectID, error)
	Get() ([]*responses.Location, error)
}

type locationRepository struct {
}

func NewLocationRepository() ILocationRepository {
	return &locationRepository{}
}

var locationCollection = database.GetInstance().Collection("locations")

// Create a new tet case
func (r *locationRepository) Create(location *models.Location) (*primitive.ObjectID, error) {

	res, err := locationCollection.InsertOne(context.TODO(), location)

	if err != nil {
		return nil, err
	}

	id := res.InsertedID.(primitive.ObjectID)
	return &id, nil
}

// Get returns a list of all test cases
func (r *locationRepository) Get() ([]*responses.Location, error) {

	cursor, err := locationCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}
	var locations []*responses.Location = []*responses.Location{}
	if err = cursor.All(context.TODO(), &locations); err != nil {
		panic(err)
	}
	cursor.Close(context.TODO())
	return locations, nil
}
