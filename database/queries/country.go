package queries

import (
	"github.com/abisaidfarias/lbtechapi/models"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

func UpdateCountry(country *models.Country, oid primitive.ObjectID) (primitive.M, primitive.M) {

	filter := primitive.M{
		"_id": oid,
	}
	update := primitive.M{
		"$set": country,
	}
	return filter, update
}
func DeleteCountry(oid primitive.ObjectID) primitive.M {

	return primitive.M{
		"_id": oid,
	}
}
func GetCountryById(oid primitive.ObjectID) primitive.M {
	return primitive.M{"_id": oid}
}
