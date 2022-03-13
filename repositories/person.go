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

type IPersonRepository interface {
	Create(*models.Person) error
	Get() ([]*responses.Person, error)
	Update(primitive.ObjectID, *models.Person) error
	Delete(primitive.ObjectID) error
}

type personRepository struct {
}

func NewPersonRepository() IPersonRepository {
	return &personRepository{}
}

var personCollection = database.GetInstance().Collection("persons")

// Create a new tet case
func (r *personRepository) Create(person *models.Person) error {

	_, err := personCollection.InsertOne(context.TODO(), person)

	if err != nil {
		return err
	}
	return nil
}

// Get returns a list of all test cases
func (r *personRepository) Get() ([]*responses.Person, error) {

	cursor, err := personCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}
	var companies []*responses.Person = []*responses.Person{}
	if err = cursor.All(context.TODO(), &companies); err != nil {
		panic(err)
	}
	cursor.Close(context.TODO())
	return companies, nil
}
func (r *personRepository) Update(id primitive.ObjectID, person *models.Person) error {

	filter, update := queries.UpdatePerson(person, id)

	_, err := personCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	return nil
}
func (r *personRepository) Delete(id primitive.ObjectID) error {

	_, err := personCollection.DeleteOne(context.TODO(), queries.DeleteCountry(id))

	if err != nil {
		return err
	}
	return nil
}
