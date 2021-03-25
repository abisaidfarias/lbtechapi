package repositories

import (
	"context"

	"gopkg.in/mgo.v2/bson"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

type IBrandRepository interface {
	Create(*models.Brand) error
	Get() ([]*responses.Brand, error)
}

type brandRepository struct {
}

func NewBrandRepository() IBrandRepository {
	return &brandRepository{}
}

var brandCollection = database.GetInstance().Collection("companies")

// Create a new tet case
func (r *brandRepository) Create(brand *models.Brand) error {

	_, err := brandCollection.InsertOne(context.TODO(), brand)

	if err != nil {
		return err
	}
	return nil
}

// Get returns a list of all test cases
func (r *brandRepository) Get() ([]*responses.Brand, error) {

	cursor, err := brandCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}
	var brands []*responses.Brand
	if err = cursor.All(context.TODO(), &brands); err != nil {
		panic(err)
	}
	cursor.Close(context.TODO())
	return brands, nil
}
