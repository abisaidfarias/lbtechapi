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
	shipmentControlEmailCreate                  = "Create"
	shipmentControlEmailValidationStart         = "ValidationStart"
	shipmentControlEmailValidationEnd           = "ValidationEnd"
	shipmentControlEmailComplete                = "Complete"
	shipmentControlEmailRequestDeleteInternal   = "RequestDeleteInternal"
	shipmentControlEmailRequestDeleteClient     = "RequestDeleteClient"
	shipmentControlEmailDeleted                 = "Deleted"
)

// Exported email kinds for delete notifications.
const (
	ShipmentControlEmailRequestDeleteInternal = shipmentControlEmailRequestDeleteInternal
	ShipmentControlEmailRequestDeleteClient   = shipmentControlEmailRequestDeleteClient
	ShipmentControlEmailDeleted               = shipmentControlEmailDeleted
)

func ShipmentControlPhaseLabel(phase int) string {
	if label, ok := enums.ShipmentControlPhaseLabels[phase]; ok {
		return label
	}
	return fmt.Sprintf("Phase %d", phase)
}

func FormatShipmentControlEmailDate(date time.Time) string {
	if date.IsZero() {
		return "-"
	}
	return fmt.Sprintf("%02d/%02d/%d", date.Day(), date.Month(), date.Year())
}

// formatEmailCommentMultiline preserves explicit newlines and, when comments store
// space-separated numeric tokens (e.g. IMEIs), renders each token on its own line.
func formatEmailCommentMultiline(comment string) string {
	comment = strings.TrimSpace(comment)
	if comment == "" || comment == "-" {
		return comment
	}

	comment = strings.ReplaceAll(comment, "\r\n", "\n")
	comment = strings.ReplaceAll(comment, "\r", "\n")
	if strings.Contains(comment, "\n") {
		return comment
	}

	parts := strings.Fields(comment)
	if len(parts) < 3 {
		return comment
	}

	splitIdx := len(parts)
	for i := len(parts) - 1; i >= 0; i-- {
		if isNumericCommentToken(parts[i]) {
			splitIdx = i
			continue
		}
		break
	}

	numbers := parts[splitIdx:]
	if len(numbers) < 2 {
		return comment
	}

	header := strings.Join(parts[:splitIdx], " ")
	var b strings.Builder
	b.WriteString(header)
	for _, number := range numbers {
		b.WriteByte('\n')
		b.WriteString(number)
	}
	return b.String()
}

func isNumericCommentToken(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
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
	case shipmentControlEmailRequestDeleteInternal, shipmentControlEmailRequestDeleteClient:
		return "Delete Requested"
	case shipmentControlEmailDeleted:
		return "Deleted"
	default:
		return "—"
	}
}

func shipmentControlEmailSubjectIdent(brand, commercialModel, softwareVersion string) string {
	sw := strings.TrimSpace(softwareVersion)
	if sw == "" {
		sw = "—"
	}
	return fmt.Sprintf("%s %s SW version %s", strings.TrimSpace(brand), strings.TrimSpace(commercialModel), sw)
}

func GetShipmentControlNotificationMessageAndSubject(
	emailKind string,
	brand string,
	commercialModel string,
	softwareVersion string,
) (string, string) {
	ident := shipmentControlEmailSubjectIdent(brand, commercialModel, softwareVersion)

	switch emailKind {
	case shipmentControlEmailCreate:
		return utils.SHIPMENT_CONTROL_CREATE_MAIN_MESSAGE,
			fmt.Sprintf("Subject: New Shipment Control request was created %s", ident)
	case shipmentControlEmailValidationStart:
		return utils.SHIPMENT_CONTROL_VALIDATION_START_MAIN_MESSAGE,
			fmt.Sprintf("Subject: Control Shipment Validation Start Date %s", ident)
	case shipmentControlEmailValidationEnd:
		return utils.SHIPMENT_CONTROL_VALIDATION_END_MAIN_MESSAGE,
			fmt.Sprintf("Subject: Control Shipment Validation End Date %s", ident)
	case shipmentControlEmailComplete:
		return utils.SHIPMENT_CONTROL_COMPLETE_MAIN_MESSAGE,
			fmt.Sprintf("Subject: Control Shipment Completed %s", ident)
	default:
		return "", ""
	}
}

func GetShipmentControlDeleteNotificationMessageAndSubject(
	emailKind string,
	brand string,
	commercialModel string,
	softwareVersion string,
	companyName string,
) (string, string) {
	ident := shipmentControlEmailSubjectIdent(brand, commercialModel, softwareVersion)
	companyName = strings.TrimSpace(companyName)
	if companyName == "" {
		companyName = "Client"
	}

	switch emailKind {
	case shipmentControlEmailRequestDeleteInternal:
		return fmt.Sprintf(utils.SHIPMENT_CONTROL_REQUEST_DELETE_INTERNAL_MAIN_MESSAGE, companyName),
			fmt.Sprintf("Subject: Shipment Control delete request from %s %s", companyName, ident)
	case shipmentControlEmailRequestDeleteClient:
		return utils.SHIPMENT_CONTROL_REQUEST_DELETE_CLIENT_MAIN_MESSAGE,
			fmt.Sprintf("Subject: Shipment Control delete request received %s", ident)
	case shipmentControlEmailDeleted:
		return utils.SHIPMENT_CONTROL_DELETED_MAIN_MESSAGE,
			fmt.Sprintf("Subject: Shipment Control project deleted %s", ident)
	default:
		return "", ""
	}
}

func ShipmentControlNotifiesExternalRecipients(notifyKey, emailKind string) bool {
	if notifyKey == utils.CREATE {
		return true
	}
	return emailKind == shipmentControlEmailComplete
}

func MultibandaNotifiesExternalRecipients(notifyKey string, currentPhase int) bool {
	if notifyKey == utils.CREATE {
		return true
	}
	return currentPhase == enums.HomologationPhase_value["COMPLETE"]
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
		ReferenceID:                 emptyAsDash(notify.ReferenceID),
		Validation:                  emptyAsDash(notify.Validation),
		ReworkNumber:                emptyAsDash(notify.ReworkNumber),
		MultibandaCertificateNumber: emptyAsDash(notify.SubtelCertificateNumber),
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
		Comments:                    formatEmailCommentMultiline(emptyAsDash(notify.Comment)),
		ExcelFileURL:                strings.TrimSpace(notify.SubtelCertificateUrl),
		ImeiFileURL:                 imeiFileURLForEmail(emailKind, notify.ImeiFileUrl),
		MultibandCertificateURL:     strings.TrimSpace(notify.MultibandCertificateUrl),
		OabiCertificateURL:          strings.TrimSpace(notify.OabiCertificateUrl),
		Year:                        now.Year(),
	}
}
