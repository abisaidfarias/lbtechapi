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

type ICountryRepository interface {
	Create(*models.Country) error
	Get() ([]*responses.Country, error)
	Update(primitive.ObjectID, *models.Country) error
	Delete(primitive.ObjectID) error
}

type countryRepository struct {
}

func NewCountryRepository() ICountryRepository {
	return &countryRepository{}
}

var countryCollection = database.GetInstance().Collection("countries")

// Create a new tet case
func (r *countryRepository) Create(country *models.Country) error {

	_, err := countryCollection.InsertOne(context.TODO(), country)

	if err != nil {
		return err
	}
	return nil
}

// Get returns a list of all test cases
func (r *countryRepository) Get() ([]*responses.Country, error) {

	cursor, err := countryCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}
	var countries []*responses.Country = []*responses.Country{}
	if err = cursor.All(context.TODO(), &countries); err != nil {
		panic(err)
	}
	cursor.Close(context.TODO())
	return countries, nil
}
func (r *countryRepository) Update(id primitive.ObjectID, country *models.Country) error {

	filter, update := queries.UpdateCountry(country, id)

	_, err := countryCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	return nil
}
func (r *countryRepository) Delete(id primitive.ObjectID) error {

	_, err := countryCollection.DeleteOne(context.TODO(), queries.DeleteCountry(id))

	if err != nil {
		return err
	}
	return nil
}
