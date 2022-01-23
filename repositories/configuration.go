package repositories

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

type IConfigurationRepository interface {
	Create(*models.Configuration) (*primitive.ObjectID, error)
	Get() ([]*responses.Configuration, error)
}

type configurationRepository struct {
}

func NewConfigurationRepository() IConfigurationRepository {
	return &configurationRepository{}
}

var configurationCollection = database.GetInstance().Collection("configurations")

// Create a new tet case
func (r *configurationRepository) Create(configuration *models.Configuration) (*primitive.ObjectID, error) {

	_, err := configurationCollection.DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	res, err := configurationCollection.InsertOne(context.TODO(), configuration)
	if err != nil {
		return nil, err
	}

	id := res.InsertedID.(primitive.ObjectID)
	return &id, nil
}

// Get returns a list of all test cases
func (r *configurationRepository) Get() ([]*responses.Configuration, error) {

	cursor, err := configurationCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}
	var configurations []*responses.Configuration = []*responses.Configuration{}
	if err = cursor.All(context.TODO(), &configurations); err != nil {
		panic(err)
	}
	cursor.Close(context.TODO())
	return configurations, nil
}
