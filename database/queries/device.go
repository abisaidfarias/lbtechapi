package queries

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"go.mongodb.org/mongo-driver/bson"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetDeviceById(oid primitive.ObjectID) primitive.M {
	return primitive.M{"_id": oid}
}
func GetDevicesExpanded() []bson.D {
	lookupStage := bson.D{
		primitive.E{Key: "$lookup", Value: bson.D{
			primitive.E{Key: "from", Value: "brands"},
			primitive.E{Key: "localField", Value: "brand"},
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "brand"},
		}}}
	unwindStage := bson.D{
		primitive.E{Key: "$unwind", Value: bson.D{
			primitive.E{Key: "path", Value: "$brand"},
			primitive.E{Key: "preserveNullAndEmptyArrays", Value: true},
		}}}
	return mongo.Pipeline{lookupStage, unwindStage}
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
