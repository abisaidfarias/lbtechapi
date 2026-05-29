package mapping



import (

	"time"



	"github.com/abisaidfarias/lbtechapi/models"

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

		RegisteredImeiCount: 0,

		ReworkNumber:        req.ReworkNumber,

		OabiCertificate:     req.OabiCertificate,

		Client:              req.Client,

		CreatedDate:         now,

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
		CertificateNumber: multibanda.CertificateNumber,
		SoftwareVersion:   multibanda.SoftwareVersion,
		HardwareVersion:   multibanda.HardwareVersion,
		OsVersion:         multibanda.OsVersion,
		OsVersionView:     multibanda.OsVersionView,
	}
}

func ToShipmentControlDeviceSummary(multibanda responses.MultibandaExpanded) responses.ShipmentControlDeviceSummary {
	return responses.ShipmentControlDeviceSummary{
		CommercialModel: multibanda.Device.CommercialModel,
		TechnicalModel:  multibanda.Device.TechnicalModel,
		Brand:           multibanda.Brand.Name,
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
