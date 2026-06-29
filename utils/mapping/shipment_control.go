package mapping



import (

	"strings"
	"time"



	"github.com/abisaidfarias/lbtechapi/models"

	"github.com/abisaidfarias/lbtechapi/utils"

	"github.com/abisaidfarias/lbtechapi/utils/enums"

	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"

	"go.mongodb.org/mongo-driver/bson/primitive"

)



func ShipmentControlRequestToShipmentControl(

	req *request.ShipmentControl,

	multibandaID primitive.ObjectID,

	companyID primitive.ObjectID,

	countryID primitive.ObjectID,

) *models.ShipmentControl {

	now := time.Now().UTC()



	return &models.ShipmentControl{

		Multibanda:          multibandaID,

		Company:             companyID,

		Country:             countryID,

		CurrentPhase:        enums.ShipmentControlPhasePlanning,

		Status:              enums.ShipmentControlStatusInProgress,

		PlanningDate:        now,

		ImeiQuantity:        req.ImeiQuantity,

		ImeiFileUrl:         req.ImeiFileUrl,

		RegisteredImeiCount: 0,

		ReworkNumber:        req.ReworkNumber,

		OabiCertificate:     req.OabiCertificate,

		Client:              req.Client,

		RequestDelete:       false,

		CreatedDate:         now,

	}

}



func ShipmentControlRequestToShipmentControlUpdate(

	req *request.ShipmentControl,

	multibandaID primitive.ObjectID,

	companyID primitive.ObjectID,

	countryID primitive.ObjectID,

) *models.ShipmentControl {

	return &models.ShipmentControl{

		Multibanda:              multibandaID,

		Company:                 companyID,

		Country:                 countryID,

		CurrentPhase:            req.CurrentPhase,

		Status:                  req.Status,

		PlanningDate:            utils.NormalizeCalendarDateUTC(req.PlanningDate),

		ValidationStartDate:     utils.NormalizeCalendarDateUTC(req.ValidationStartDate),

		ValidationEndDate:       utils.NormalizeCalendarDateUTC(req.ValidationEndDate),

		UnderRevisionStartDate:  utils.NormalizeCalendarDateUTC(req.UnderRevisionStartDate),

		UnderRevisionEndDate:    utils.NormalizeCalendarDateUTC(req.UnderRevisionEndDate),

		CompletedDate:           utils.NormalizeCalendarDateUTC(req.CompletedDate),

		ImeiQuantity:            req.ImeiQuantity,

		ImeiFileUrl:             req.ImeiFileUrl,

		RegisteredImeiCount:     req.RegisteredImeiCount,

		ReworkNumber:            req.ReworkNumber,

		OabiCertificate:         req.OabiCertificate,

		Client:                  req.Client,

		SubtelCertificateUrl:    req.SubtelCertificateUrl,

		SubtelCertificateNumber: req.SubtelCertificateNumber,

		OabiCertificateUrl:      req.OabiCertificateUrl,

		OabiCertificateNumber:   req.OabiCertificateNumber,

		Comment:                 req.Comment,

	}

}



func ShipmentControlRequestToShipmentControlResume(

	req *request.ShipmentControlResume,

	countryID primitive.ObjectID,

) *models.ShipmentControl {

	shipmentControl := &models.ShipmentControl{

		CurrentPhase:            req.CurrentPhase,

		ValidationStartDate:     req.ValidationStartDate,

		ValidationEndDate:       req.ValidationEndDate,

		UnderRevisionStartDate:  req.UnderRevisionStartDate,

		UnderRevisionEndDate:    req.UnderRevisionEndDate,

		CompletedDate:           req.CompletedDate,

		ImeiQuantity:            req.ImeiQuantity,

		ImeiFileUrl:             req.ImeiFileUrl,

		RegisteredImeiCount:     req.RegisteredImeiCount,

		ReworkNumber:            req.ReworkNumber,

		OabiCertificate:         req.OabiCertificate,

		Client:                  req.Client,

		SubtelCertificateUrl:    req.SubtelCertificateUrl,

		SubtelCertificateNumber: req.SubtelCertificateNumber,

		OabiCertificateUrl:      req.OabiCertificateUrl,

		OabiCertificateNumber:   req.OabiCertificateNumber,

		Comment:                 req.Comment,

	}

	if !countryID.IsZero() {

		shipmentControl.Country = countryID

	}

	return shipmentControl
}

func ToShipmentControlMultibandaSummary(multibanda responses.MultibandaExpanded) responses.ShipmentControlMultibandaSummary {
	return responses.ShipmentControlMultibandaSummary{
		ID:                multibanda.ID,
		SubtelCertificateNumber: multibanda.SubtelCertificateNumber,
		SoftwareVersion:   multibanda.SoftwareVersion,
		HardwareVersion:   multibanda.HardwareVersion,
		OsVersion:         multibanda.OsVersion,
		OsVersionView:     multibanda.OsVersionView,
	}
}

func ShipmentControlToNotify(
	shipment *models.ShipmentControl,
	multibanda *responses.MultibandaExpanded,
	companyName string,
	countryName string,
) request.ShipmentControlNotify {
	notify := request.ShipmentControlNotify{
		CompanyName:             companyName,
		CountryName:             countryName,
		Client:                  shipment.Client,
		ReworkNumber:            shipment.ReworkNumber,
		ImeiQuantity:            shipment.ImeiQuantity,
		ImeiFileUrl:             shipment.ImeiFileUrl,
		RegisteredImeiCount:     shipment.RegisteredImeiCount,
		CurrentPhase:            shipment.CurrentPhase,
		PlanningDate:            shipment.PlanningDate,
		ValidationStartDate:     shipment.ValidationStartDate,
		ValidationEndDate:       shipment.ValidationEndDate,
		UnderRevisionStartDate:  shipment.UnderRevisionStartDate,
		UnderRevisionEndDate:    shipment.UnderRevisionEndDate,
		CompletedDate:           shipment.CompletedDate,
		OabiCertificate:         firstNonEmpty(shipment.OabiCertificateNumber, shipment.OabiCertificate),
		Comment:                 shipment.Comment,
		SubtelCertificateUrl:    shipment.SubtelCertificateUrl,
		OabiCertificateUrl:      shipment.OabiCertificateUrl,
	}
	if multibanda != nil {
		notify.SubtelCertificateNumber = firstNonEmpty(
			shipment.SubtelCertificateNumber,
			multibanda.SubtelCertificateNumber,
		)
		notify.MultibandCertificateUrl = multibanda.MultibandCertificateUrl
		notify.SoftwareVersion = multibanda.SoftwareVersion
		notify.HardwareVersion = multibanda.HardwareVersion
		notify.OsVersion = multibanda.OsVersion
		notify.OsVersionView = multibanda.OsVersionView
	} else {
		notify.SubtelCertificateNumber = strings.TrimSpace(shipment.SubtelCertificateNumber)
	}
	return notify
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

func ToShipmentControlDeviceSummary(multibanda responses.MultibandaExpanded) responses.ShipmentControlDeviceSummary {
	return responses.ShipmentControlDeviceSummary{
		CommercialModel: multibanda.Device.CommercialModel,
		TechnicalModel:  multibanda.Device.TechnicalModel,
		Brand:           multibanda.Brand.Name,
		PlatformOs:      multibanda.Device.PlatformOs,
	}
}

func SanitizeShipmentControlExpandedDates(item *responses.ShipmentControlExpanded) {
	if item == nil {
		return
	}
	item.ValidationStartDate = nilIfZeroTimePtr(item.ValidationStartDate)
	item.ValidationEndDate = nilIfZeroTimePtr(item.ValidationEndDate)
	item.UnderRevisionStartDate = nilIfZeroTimePtr(item.UnderRevisionStartDate)
	item.UnderRevisionEndDate = nilIfZeroTimePtr(item.UnderRevisionEndDate)
	item.CompletedDate = nilIfZeroTimePtr(item.CompletedDate)
}

func nilIfZeroTimePtr(value *time.Time) *time.Time {
	if value == nil || value.IsZero() {
		return nil
	}
	return value
}
