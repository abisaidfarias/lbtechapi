package queries

import (
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"
)

func GetUsersProfileId(id string) primitive.M {

	return primitive.M{"profile": bson.ObjectIdHex(id)}

}
