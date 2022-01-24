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
func GetDeviceTrackingExpanded(isInternal bool, companyID primitive.ObjectID) []bson.D {
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

	if !isInternal {
		matchStage := bson.D{
			primitive.E{Key: "$match", Value: bson.D{
				primitive.E{Key: "company._id", Value: companyID},
			}}}
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