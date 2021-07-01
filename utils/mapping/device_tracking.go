package mapping

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func DeviceTrackinRequestToDeviceTracking(deviceTraking *request.DeviceTracking, imei string) *models.DeviceTracking {
	companyID, _ := primitive.ObjectIDFromHex(deviceTraking.Company)
	deviceID, _ := primitive.ObjectIDFromHex(deviceTraking.Device)
	countryID, _ := primitive.ObjectIDFromHex(deviceTraking.TrackingLog.Country)
	locationID, _ := primitive.ObjectIDFromHex(deviceTraking.TrackingLog.Location)
	userID, _ := primitive.ObjectIDFromHex(deviceTraking.TrackingLog.InternalResponsible)

	trackingLog := models.TrackingLog{
		Country:             countryID,
		Location:            locationID,
		InternalResponsible: userID,
		ExternalResponsible: deviceTraking.TrackingLog.ExternalResponsible,
		Comment:             deviceTraking.TrackingLog.Comment,
		DocumentUrl:         deviceTraking.TrackingLog.DocumentUrl,
		TrackingDate:        deviceTraking.TrackingLog.TrackingDate,
	}

	return &models.DeviceTracking{
		Company:     companyID,
		Device:      deviceID,
		Imei:        imei,
		TrackingLog: trackingLog,
	}
}
