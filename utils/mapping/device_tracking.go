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
	person, _ := PersonRequestToPerson(&deviceTraking.TrackingLog.Person)
	trackingLog := models.TrackingLog{
		TrackingID:          deviceTraking.TrackingLog.TrackingID,
		Country:             *country,
		Location:            *location,
		InternalResponsible: *user,
		Person:              *person,
		Comment:             deviceTraking.TrackingLog.Comment,
		DocumentUrl:         deviceTraking.TrackingLog.DocumentUrl,
		TrackingDate:        deviceTraking.TrackingLog.TrackingDate,
		ExternalDelivery:    deviceTraking.TrackingLog.ExternalDelivery,
		ProcessTypes:        append([]string{}, deviceTraking.TrackingLog.ProcessTypes...),
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
	person, _ := PersonRequestToPerson(&trackingLogReq.Person)

	return &models.TrackingLog{
		TrackingID:          trackingLogReq.TrackingID,
		Country:             *country,
		Location:            *location,
		InternalResponsible: *user,
		Person:              *person,
		Comment:             trackingLogReq.Comment,
		DocumentUrl:         trackingLogReq.DocumentUrl,
		TrackingDate:        trackingLogReq.TrackingDate,
		ExternalDelivery:    trackingLogReq.ExternalDelivery,
		ProcessTypes:        append([]string{}, trackingLogReq.ProcessTypes...),
	}
}
func DeviceTrackinRequestToDeviceTrackingUpdate(deviceTraking *request.DeviceTrackingExpanded) *models.DeviceTracking {
	companyID, _ := primitive.ObjectIDFromHex(string(deviceTraking.Company.ID))
	deviceID, _ := primitive.ObjectIDFromHex(string(deviceTraking.Device.ID))
	var trakings []models.TrackingLog = []models.TrackingLog{}

	for _, trakingLog := range deviceTraking.TrackingLogs {

		trakings = append(trakings, *TrackinLogRequestToTrackingLog(&trakingLog))
	}
	return &models.DeviceTracking{
		Company:      companyID,
		Device:       deviceID,
		Imei:         deviceTraking.Imei,
		TrackingLogs: trakings,
	}
}
