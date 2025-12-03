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

type IBrandRepository interface {
	Create(*models.Brand) (*primitive.ObjectID, error)
	Get() ([]*responses.Brand, error)
	GetById(string) (*responses.Brand, error)
}

type brandRepository struct {
}

func NewBrandRepository() IBrandRepository {
	return &brandRepository{}
}

var brandCollection = database.GetInstance().Collection("brands")

// Create a new tet case
func (r *brandRepository) Create(brand *models.Brand) (*primitive.ObjectID, error) {

	res, err := brandCollection.InsertOne(context.TODO(), brand)

	if err != nil {
		return nil, err
	}
	id := res.InsertedID.(primitive.ObjectID)
	return &id, nil
}

// Get returns a list of all test cases
func (r *brandRepository) Get() ([]*responses.Brand, error) {

	cursor, err := brandCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}
	var brands []*responses.Brand = []*responses.Brand{}
	if err = cursor.All(context.TODO(), &brands); err != nil {
		panic(err)
	}
	cursor.Close(context.TODO())
	return brands, nil
}
func (r *brandRepository) GetById(id string) (*responses.Brand, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}

	var result responses.Brand

	err = brandCollection.FindOne(context.TODO(),
		queries.GetBrandById(oid)).Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}
