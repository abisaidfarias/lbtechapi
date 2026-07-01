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

func GetDevicesByTechnicalModel(technicalModel string) primitive.M {
	return primitive.M{"technical_model": technicalModel}
}
func GetDevicesExpanded(brands []primitive.ObjectID) []bson.D {
	var pipeline mongo.Pipeline

	if len(brands) > 0 {
		matchStage := bson.D{
			primitive.E{Key: "$match", Value: bson.D{
				primitive.E{Key: "brand", Value: bson.D{
					primitive.E{Key: "$in", Value: brands},
				}},
			}},
		}
		pipeline = append(pipeline, matchStage)
	}

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
	sort := bson.D{
		primitive.E{Key: "$sort", Value: bson.D{
			primitive.E{Key: "commercial_model", Value: -1},
		}}}
	pipeline = append(pipeline, lookupStage, unwindStage, sort)
	return pipeline
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
