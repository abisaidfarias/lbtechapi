package repositories

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/database/queries"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

type IHomologationRepository interface {
	Create(*models.Homologation) error
	GetPrevious(primitive.ObjectID,
		primitive.ObjectID, primitive.ObjectID) (*responses.Homologation, error)
	GetByInternal([]primitive.ObjectID, []primitive.ObjectID,
		[]primitive.ObjectID) ([]*responses.HomologationExpanded, error)
	//GetByExternal(primitive.ObjectID, []primitive.ObjectID, []primitive.ObjectID) ([]*responses.Homologation, error)
}

type homologationRepository struct {
}

func NewHomologationRepository() IHomologationRepository {
	return &homologationRepository{}
}

var homologationCollection = database.GetInstance().Collection("homologations")

// Create a new tet case
func (r *homologationRepository) Create(homologation *models.Homologation) error {

	_, err := homologationCollection.InsertOne(context.TODO(), homologation)

	if err != nil {
		return err
	}
	return nil
}

// Get returns a list of all test cases
func (r *homologationRepository) Get() ([]*responses.Homologation, error) {

	cursor, err := homologationCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}
	var homologations []*responses.Homologation = []*responses.Homologation{}
	if err = cursor.All(context.TODO(), &homologations); err != nil {
		panic(err)
	}
	cursor.Close(context.TODO())
	return homologations, nil
}

// Get returns a list of all test cases
func (r *homologationRepository) GetPrevious(deviceId primitive.ObjectID,
	countryId primitive.ObjectID, companyId primitive.ObjectID) (*responses.Homologation, error) {

	var homologation *responses.Homologation
	err := homologationCollection.FindOne(context.TODO(),
		queries.GetHomologationValidations(deviceId, countryId, companyId)).Decode(&homologation)

	switch err {
	case mongo.ErrNoDocuments:
		return homologation, nil
	default:
		return homologation, err
	}
}
func (r *homologationRepository) GetByInternal(companies []primitive.ObjectID,
	devices []primitive.ObjectID, countries []primitive.ObjectID) ([]*responses.HomologationExpanded, error) {

	cursor, err := homologationCollection.Aggregate(context.TODO(),
		queries.GetHomologationsInternal(companies, devices, countries))
	if err != nil {
		return nil, err
	}
	var homologations []*responses.HomologationExpanded = []*responses.HomologationExpanded{}
	for cursor.Next(context.TODO()) {
		var homologation responses.HomologationExpanded
		err := cursor.Decode(&homologation)
		if err != nil {
			return nil, err
		}
		homologations = append(homologations, &homologation)
	}

	return homologations, nil

}

// func (r *homologationRepository) GetByExternal(deviceId primitive.ObjectID,
// 	countryId primitive.ObjectID, companyId primitive.ObjectID) (*responses.Homologation, error) {

// 	var homologation *responses.Homologation
// 	err := homologationCollection.FindOne(context.TODO(),
// 		queries.GetHomologationValidations(deviceId, countryId, companyId)).Decode(&homologation)

// 	switch err {
// 	case mongo.ErrNoDocuments:
// 		return homologation, nil
// 	default:
// 		return homologation, err
// 	}
// }
