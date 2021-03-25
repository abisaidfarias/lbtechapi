package queries

import (
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

func GetUsersProfileId(id string) primitive.M {
	val, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		panic(err)
	}
	return primitive.M{"profile": val}
}
