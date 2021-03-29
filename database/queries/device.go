package queries

import (
	"github.com/abisaidfarias/lbtechapi/models"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

func GetDeviceById(oid primitive.ObjectID) primitive.M {
	return primitive.M{"_id": oid}
}
func UpdateDevice(device *models.Device, oid primitive.ObjectID) (primitive.M, primitive.M) {

	filter := primitive.M{
		"_id": oid,
	}
	update := primitive.M{
		"$set": device,
	}
	return filter, update
}
func DeleteDevice(oid primitive.ObjectID) primitive.M {

	return primitive.M{
		"_id": oid,
	}
}
