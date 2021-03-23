package queries

import (
	"github.com/abisaidfarias/lbtechapi/models"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

func GeProfileById(oid primitive.ObjectID) primitive.M {
	return primitive.M{"_id": oid}
}
func UpdateProfile(profile *models.Profile, oid primitive.ObjectID) (primitive.M, primitive.M) {

	filter := primitive.M{
		"_id": oid,
	}
	update := primitive.M{
		"$set": profile,
	}
	return filter, update
}
func DeleteProfile(oid primitive.ObjectID) primitive.M {

	return primitive.M{
		"_id": oid,
	}
}
