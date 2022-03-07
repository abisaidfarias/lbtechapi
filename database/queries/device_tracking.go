package queries

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"go.mongodb.org/mongo-driver/bson"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AddTrackingLog(trackingLog *models.TrackingLog, oid primitive.ObjectID) (primitive.M, primitive.M) {

	filter := primitive.M{
		"_id": oid,
	}
	update := primitive.M{
		"$push": primitive.M{
			"tracking_logs": trackingLog,
		},
	}
	return filter, update
}
func GetDeviceTrackingExpanded(isInternal bool, companyID primitive.ObjectID,brands []primitive.ObjectID) []bson.D {
	lookupCompany := bson.D{
		primitive.E{Key: "$lookup", Value: bson.D{
			primitive.E{Key: "from", Value: "companies"},
			primitive.E{Key: "localField", Value: "company"},
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "company"},
		}}}
	unwindCompany := bson.D{
		primitive.E{Key: "$unwind", Value: bson.D{
			primitive.E{Key: "path", Value: "$company"},
			primitive.E{Key: "preserveNullAndEmptyArrays", Value: true},
		}}}
	lookupDevice := bson.D{
		primitive.E{Key: "$lookup", Value: bson.D{
			primitive.E{Key: "from", Value: "devices"},
			primitive.E{Key: "localField", Value: "device"},
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "device"},
		}}}
	unwindDevice := bson.D{
		primitive.E{Key: "$unwind", Value: bson.D{
			primitive.E{Key: "path", Value: "$device"},
			primitive.E{Key: "preserveNullAndEmptyArrays", Value: true},
		}}}
	lookupBrand := bson.D{
		primitive.E{Key: "$lookup", Value: bson.D{
			primitive.E{Key: "from", Value: "brands"},
			primitive.E{Key: "localField", Value: "device.brand"},
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "device.brand"},
		}}}
	unwindBrand := bson.D{
		primitive.E{Key: "$unwind", Value: bson.D{
			primitive.E{Key: "path", Value: "$device.brand"},
			primitive.E{Key: "preserveNullAndEmptyArrays", Value: true},
		}}}
	var objectStage bson.D
	var matchStage bson.D
	if len(brands) > 0 {

		var brandStageIn bson.D
		brandStageIn = append(brandStageIn, primitive.E{Key: "$in", Value: brands})
		objectStage = append(matchStage, primitive.E{Key: "device.brand._id", Value: brandStageIn})
	}

	if !isInternal {
		objectStage = append(objectStage, primitive.E{Key: "company._id", Value: companyID})
		matchStage = append(matchStage, primitive.E{Key: "$match", Value: objectStage})
		// matchStage := bson.D{
		// 	primitive.E{Key: "$match", Value: bson.D{
		// 		primitive.E{Key: "company._id", Value: companyID},
		// 	}}}
		return mongo.Pipeline{lookupCompany, unwindCompany,
			lookupDevice, unwindDevice, lookupBrand, unwindBrand, matchStage}
	}
	return mongo.Pipeline{lookupCompany, unwindCompany,
		lookupDevice, unwindDevice, lookupBrand, unwindBrand}
}
func GetDeviceTrackingByCompany(companyId primitive.ObjectID) primitive.M {
	return primitive.M{"company": companyId}
}
func GetDeviceTrackingByDevice(deviceId primitive.ObjectID) primitive.M {
	return primitive.M{"device": deviceId}
}
func DeleteDeviceTracking(ids []primitive.ObjectID) primitive.M {

	filter := primitive.M{
		"_id": primitive.M{
			"$in": ids,
		},
	}
	return filter
}
func UpdateDeviceTracking(deviceTracking *models.DeviceTracking, oid primitive.ObjectID) (primitive.M, primitive.M) {

	filter := primitive.M{
		"_id": oid,
	}
	update := primitive.M{
		"$set": deviceTracking,
	}
	return filter, update
}