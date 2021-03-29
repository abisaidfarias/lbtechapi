package repositories

import (
	"context"

	"gopkg.in/mgo.v2/bson"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

type ICompanyRepository interface {
	Create(*models.Company) error
	Get() ([]*responses.Company, error)
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
