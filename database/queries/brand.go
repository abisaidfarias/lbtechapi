package queries

import (
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

func GetBrandById(oid primitive.ObjectID) primitive.M {
	return primitive.M{"_id": oid}
}
