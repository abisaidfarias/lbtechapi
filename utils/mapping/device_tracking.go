package mapping

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func DeviceTrackinRequestToDeviceTracking(deviceTraking *request.DeviceTracking, imei string) *models.DeviceTracking {
	companyID, _ := primitive.ObjectIDFromHex(deviceTraking.Company)
	deviceID, _ := primitive.ObjectIDFromHex(deviceTraking.Device)
	country, _ := CountryRequestToCountry(&deviceTraking.TrackingLog.Country)
	location, _ := LocationRequestToLocation(&deviceTraking.TrackingLog.Location)
	user := UserResumeToUser(&deviceTraking.TrackingLog.InternalResponsible)
	trackingLog := models.TrackingLog{
		Country:             *country,
		Location:            *location,
		InternalResponsible: *user,
		ExternalResponsible: deviceTraking.TrackingLog.ExternalResponsible,
		Comment:             deviceTraking.TrackingLog.Comment,
		DocumentUrl:         deviceTraking.TrackingLog.DocumentUrl,
		TrackingDate:        deviceTraking.TrackingLog.TrackingDate,
	}
	var trakings []models.TrackingLog = []models.TrackingLog{}

	trakings = append(trakings, trackingLog)
	return &models.DeviceTracking{
		Company:      companyID,
		Device:       deviceID,
		Imei:         imei,
		TrackingLogs: trakings,
	}
}
func TrackinLogRequestToTrackingLog(trackingLogReq *request.TrackingLog) *models.TrackingLog {

	country, _ := CountryRequestToCountry(&trackingLogReq.Country)
	location, _ := LocationRequestToLocation(&trackingLogReq.Location)
	user := UserResumeToUser(&trackingLogReq.InternalResponsible)

	return &models.TrackingLog{
		Country:             *country,
		Location:            *location,
		InternalResponsible: *user,
		ExternalResponsible: trackingLogReq.ExternalResponsible,
		Comment:             trackingLogReq.Comment,
		DocumentUrl:         trackingLogReq.DocumentUrl,
		TrackingDate:        trackingLogReq.TrackingDate,
	}
}
func DeviceTrackinRequestToDeviceTrackingUpdate(deviceTraking *request.DeviceTracking) *models.DeviceTracking {
	companyID, _ := primitive.ObjectIDFromHex(deviceTraking.Company)
	deviceID, _ := primitive.ObjectIDFromHex(deviceTraking.Device)
	country, _ := CountryRequestToCountry(&deviceTraking.TrackingLog.Country)
	location, _ := LocationRequestToLocation(&deviceTraking.TrackingLog.Location)
	user := UserResumeToUser(&deviceTraking.TrackingLog.InternalResponsible)
	trackingLog := models.TrackingLog{
		Country:             *country,
		Location:            *location,
		InternalResponsible: *user,
		ExternalResponsible: deviceTraking.TrackingLog.ExternalResponsible,
		Comment:             deviceTraking.TrackingLog.Comment,
		DocumentUrl:         deviceTraking.TrackingLog.DocumentUrl,
		TrackingDate:        deviceTraking.TrackingLog.TrackingDate,
	}
	var trakings []models.TrackingLog = []models.TrackingLog{}

	for _, trakingLog := range deviceTraking.TrackingLogs {

		trakings = append(trakings, *TrackinLogRequestToTrackingLog(&trakingLog))
	}
	trakings = append(trakings, trackingLog)
	return &models.DeviceTracking{
		Company:      companyID,
		Device:       deviceID,
		Imei:         deviceTraking.Imei,
		TrackingLogs: trakings,
	}
}
