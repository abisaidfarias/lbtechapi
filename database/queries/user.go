package queries

import (
	"github.com/abisaidfarias/lbtechapi/models"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

func GetUsersProfileId(id string) primitive.M {
	val, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		panic(err)
	}
	return primitive.M{"profile": val}
}

func UpdateUser(user *models.User, oid primitive.ObjectID) (primitive.M, primitive.M) {

	filter := primitive.M{
		"_id": oid,
	}
	update := primitive.M{
		"$set": user,
	}
	return filter, update
}
func DeleteUser(oid primitive.ObjectID) primitive.M {

	return primitive.M{
		"_id": oid,
	}
}
func GetUserById(oid primitive.ObjectID) primitive.M {
	return primitive.M{"_id": oid}
}
func GetUserByEmail(email string) primitive.M {
	return primitive.M{"email": email}
}
