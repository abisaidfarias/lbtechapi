package queries

import (
	"github.com/abisaidfarias/lbtechapi/models"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

func UpdateCompany(company *models.Company, oid primitive.ObjectID) (primitive.M, primitive.M) {

	filter := primitive.M{
		"_id": oid,
	}
	update := primitive.M{
		"$set": company,
	}
	return filter, update
}
func DeleteCompany(oid primitive.ObjectID) primitive.M {

	return primitive.M{
		"_id": oid,
	}
}
