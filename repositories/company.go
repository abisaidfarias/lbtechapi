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

type ICompanyRepository interface {
	Create(*models.Company) error
	Get() ([]*responses.Company, error)
	Update(primitive.ObjectID, *models.Company) error
	Delete(primitive.ObjectID) error
	GetById(string) (*responses.Company, error)
}

type companyRepository struct {
}

func NewCompanyRepository() ICompanyRepository {
	return &companyRepository{}
}

var companyCollection = database.GetInstance().Collection("companies")

// Create a new tet case
func (r *companyRepository) Create(company *models.Company) error {

	_, err := companyCollection.InsertOne(context.TODO(), company)

	if err != nil {
		return err
	}
	return nil
}

// Get returns a list of all test cases
func (r *companyRepository) Get() ([]*responses.Company, error) {

	cursor, err := companyCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}
	var companies []*responses.Company = []*responses.Company{}
	if err = cursor.All(context.TODO(), &companies); err != nil {
		panic(err)
	}
	cursor.Close(context.TODO())
	return companies, nil
}
func (r *companyRepository) Update(id primitive.ObjectID, company *models.Company) error {

	filter, update := queries.UpdateCompany(company, id)

	_, err := companyCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	return nil
}
func (r *companyRepository) Delete(id primitive.ObjectID) error {

	_, err := companyCollection.DeleteOne(context.TODO(), queries.DeleteCountry(id))

	if err != nil {
		return err
	}
	return nil
}
func (r *companyRepository) GetById(id string) (*responses.Company, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}

	var result responses.Company

	err = companyCollection.FindOne(context.TODO(),
		queries.GetCompanyById(oid)).Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}
