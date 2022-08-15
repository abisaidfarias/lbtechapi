package queries

import (
	"github.com/abisaidfarias/lbtechapi/models"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

func UpdatePerson(person *models.Person, oid primitive.ObjectID) (primitive.M, primitive.M) {

	filter := primitive.M{
		"_id": oid,
	}
	update := primitive.M{
		"$set": person,
	}
	return filter, update
}
func DeletePerson(oid primitive.ObjectID) primitive.M {

	return primitive.M{
		"_id": oid,
	}
}
