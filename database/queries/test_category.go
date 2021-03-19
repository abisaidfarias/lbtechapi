package queries

import (
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

func InsertTestCase(oid primitive.ObjectID, oidTestCase primitive.ObjectID) (primitive.M, primitive.M) {

	filter := primitive.M{
		"_id": oid,
	}
	update := primitive.M{
		"$addToSet": primitive.M{"test_cases": oidTestCase},
	}
	return filter, update
}
