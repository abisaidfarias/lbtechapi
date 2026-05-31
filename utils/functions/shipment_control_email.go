package functions

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

const (
	shipmentControlEmailCreate          = "Create"
	shipmentControlEmailValidationStart = "ValidationStart"
	shipmentControlEmailValidationEnd   = "ValidationEnd"
	shipmentControlEmailComplete        = "Complete"
)

func ShipmentControlPhaseLabel(phase int) string {
	if label, ok := enums.ShipmentControlPhaseLabels[phase]; ok {
		return label
	}
	return fmt.Sprintf("Phase %d", phase)
}

func FormatShipmentControlEmailDate(date time.Time) string {
	if date.IsZero() {
		return "N/A"
	}
	return fmt.Sprintf("%02d/%02d/%d", date.Day(), date.Month(), date.Year())
}

func shipmentControlDisplayPhase(emailKind string) string {
	switch emailKind {
	case shipmentControlEmailCreate:
		return ShipmentControlPhaseLabel(enums.ShipmentControlPhasePlanning)
	case shipmentControlEmailValidationStart:
		return ShipmentControlPhaseLabel(enums.ShipmentControlPhaseValidation)
	case shipmentControlEmailValidationEnd:
		return ShipmentControlPhaseLabel(enums.ShipmentControlPhaseUnderRevision)
	case shipmentControlEmailComplete:
		return ShipmentControlPhaseLabel(enums.ShipmentControlPhaseCompleted)
	default:
		return "—"
	}
}

func GetShipmentControlNotificationMessageAndSubject(
	emailKind string,
	brand string,
	commercialModel string,
	reworkNumber string,
) (string, string) {
	rework := strings.TrimSpace(reworkNumber)
	if rework == "" {
		rework = "—"
	}

	switch emailKind {
	case shipmentControlEmailCreate:
		return utils.SHIPMENT_CONTROL_CREATE_MAIN_MESSAGE,
			fmt.Sprintf("Subject: New Shipment Control request was created %s %s %s", brand, commercialModel, rework)
	case shipmentControlEmailValidationStart:
		return utils.SHIPMENT_CONTROL_VALIDATION_START_MAIN_MESSAGE,
			fmt.Sprintf("Subject: Control Shipment Validation Start Date %s %s %s", brand, commercialModel, rework)
	case shipmentControlEmailValidationEnd:
		return utils.SHIPMENT_CONTROL_VALIDATION_END_MAIN_MESSAGE,
			fmt.Sprintf("Subject: Control Shipment Validation End Date %s %s %s", brand, commercialModel, rework)
	case shipmentControlEmailComplete:
		return utils.SHIPMENT_CONTROL_COMPLETE_MAIN_MESSAGE,
			fmt.Sprintf("Subject: Control Shipment Completed %s %s %s", brand, commercialModel, rework)
	default:
		return "", ""
	}
}

func ResolveShipmentControlPhaseEmailKind(
	notifyKey string,
	existing *request.ShipmentControlNotify,
	updated *request.ShipmentControlNotify,
) string {
	if notifyKey == utils.CREATE {
		return shipmentControlEmailCreate
	}

	if updated == nil {
		return ""
	}

	if updated.CurrentPhase == enums.ShipmentControlPhaseCompleted ||
		!updated.UnderRevisionEndDate.IsZero() {
		return shipmentControlEmailComplete
	}

	if existing != nil && !existing.ValidationEndDate.IsZero() {
		// already had validation end; check other transitions below
	} else if !updated.ValidationEndDate.IsZero() {
		return shipmentControlEmailValidationEnd
	}

	if existing != nil && !existing.ValidationStartDate.IsZero() {
		// fall through to phase switch
	} else if !updated.ValidationStartDate.IsZero() {
		return shipmentControlEmailValidationStart
	}

	switch updated.CurrentPhase {
	case enums.ShipmentControlPhaseValidation:
		return shipmentControlEmailValidationStart
	case enums.ShipmentControlPhaseUnderRevision:
		return shipmentControlEmailValidationEnd
	case enums.ShipmentControlPhaseCompleted:
		return shipmentControlEmailComplete
	default:
		return ""
	}
}

func imeiFileURLForEmail(emailKind, imeiFileURL string) string {
	if emailKind != shipmentControlEmailCreate {
		return ""
	}
	return strings.TrimSpace(imeiFileURL)
}

func BuildShipmentControlPhaseEmailData(
	notify *request.ShipmentControlNotify,
	brandName string,
	technicalModel string,
	commercialModel string,
	platformOs string,
	userName string,
	mainMessage string,
	emailKind string,
) ShipmentControlPhaseEmailData {
	now := time.Now()
	dearName := strings.TrimSpace(notify.CompanyName)
	if dearName == "" {
		dearName = "Client"
	}

	showImei := emailKind == shipmentControlEmailValidationStart || emailKind == shipmentControlEmailComplete
	imeiSubmitted := "—"
	if emailKind == shipmentControlEmailComplete && notify.RegisteredImeiCount > 0 {
		imeiSubmitted = strconv.Itoa(notify.RegisteredImeiCount)
	}

	return ShipmentControlPhaseEmailData{
		ClientName:                  dearName,
		MainMessage:                 mainMessage,
		NotificationDate:            FormatShipmentControlEmailDate(now),
		CurrentPhase:                shipmentControlDisplayPhase(emailKind),
		Country:                     emptyAsDash(notify.CountryName),
		Client:                      emptyAsDash(notify.Client),
		ImeiQuantity:                strconv.Itoa(notify.ImeiQuantity),
		ReworkNumber:                emptyAsDash(notify.ReworkNumber),
		MultibandaCertificateNumber: emptyAsDash(notify.CertificateNumber),
		Brand:                       brandName,
		TechnicalModel:              technicalModel,
		CommercialModel:             commercialModel,
		SoftwareVersion:             emptyAsDash(notify.SoftwareVersion),
		HardwareVersion:             emptyAsDash(notify.HardwareVersion),
		OsVersion:                   FormatMultibandaOsVersion(notify.OsVersionView, platformOs, notify.OsVersion),
		UpdatedBy:                   emptyAsDash(userName),
		PlanningDate:                FormatShipmentControlEmailDate(notify.PlanningDate),
		ValidationStartDate:         FormatShipmentControlEmailDate(notify.ValidationStartDate),
		ValidationEndDate:           FormatShipmentControlEmailDate(notify.ValidationEndDate),
		UnderRevisionStartDate:      FormatShipmentControlEmailDate(notify.UnderRevisionStartDate),
		UnderRevisionEndDate:        FormatShipmentControlEmailDate(notify.UnderRevisionEndDate),
		ResultDate:                  FormatShipmentControlEmailDate(notify.CompletedDate),
		ShowImeiSubmitted:           showImei,
		ImeiSubmitted:               imeiSubmitted,
		ShowCompletedFields:         emailKind == shipmentControlEmailComplete,
		OabiCertificate:             emptyAsDash(notify.OabiCertificate),
		Comments:                    emptyAsDash(notify.Comment),
		ExcelFileURL:                strings.TrimSpace(notify.SubtelCertificateUrl),
		ImeiFileURL:                 imeiFileURLForEmail(emailKind, notify.ImeiFileUrl),
		MultibandCertificateURL:     strings.TrimSpace(notify.MultibandCertificateUrl),
		OabiCertificateURL:          strings.TrimSpace(notify.OabiCertificateUrl),
		Year:                        now.Year(),
	}
}
