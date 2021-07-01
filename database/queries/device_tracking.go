package queries

import (
	"github.com/abisaidfarias/lbtechapi/models"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

func GetDeviceTracking(oid primitive.ObjectID) primitive.M {
	return primitive.M{"company": oid}
}
func AddTrackingLog(trackingLog *models.TrackingLog, oid primitive.ObjectID) (primitive.M, primitive.M) {

	filter := primitive.M{
		"_id": oid,
	}
	update := primitive.M{
		"$push": primitive.M{
			"tracking_log": trackingLog,
		},
	}
	return filter, update
}
