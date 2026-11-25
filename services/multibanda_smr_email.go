package services

import (
	"fmt"
	"log"
	"time"

	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SmrPendingMainMessage opens the "SMR without testing pending validation"
// email sent to every internal user.
const SmrPendingMainMessage = "A new SMR without testing request needs validation of FM Radio and the stamp. Any internal user can validate it in the platform. Requested on"

// Injection points so tests can check recipients and content without SMTP or
// a database.
var (
	smrInternalRecipients = func(companyID primitive.ObjectID) ([]string, bool) {
		return functions.GetEmails(true, companyID)
	}
	sendSmrEmail = func(to []string, subject string, data functions.MultibandaPhaseEmailData) error {
		return functions.SendMultibandaPhaseEmail(to, subject, data,
			utils.TEMPLATE_MULTIBANDA_PHASE_PATH, utils.LBOneTrackLogoPNG)
	}
)

// BuildSmrPendingEmail returns the subject and data of the pending email.
func BuildSmrPendingEmail(m *responses.MultibandaExpanded) (string, functions.MultibandaPhaseEmailData) {
	notify := mapping.MultibandaResponseToMultibandaNotify(*m)
	data := functions.BuildMultibandaPhaseEmailData(&notify, m.Brand.Name, m.Device.TechnicalModel,
		m.Device.CommercialModel, m.Device.PlatformOs, m.Device.SarValue, m.RequesterName,
		SmrPendingMainMessage, false, "", utils.PHASE)
	data.HeaderTitle = "SMR without testing · validation required"
	data.ClientName = "LB Technology Team"
	if m.CreatedDate != nil {
		data.NotificationDate = functions.FormatMultibandaEmailDate(*m.CreatedDate)
	} else {
		data.NotificationDate = functions.FormatMultibandaEmailDate(time.Now())
	}
	functions.ApplySkippedTestingDates(&data)
	data.ActionURL = functions.MultibandaRequestURL(m.ID.Hex())
	data.ActionLabel = "Validate request"
	subject := fmt.Sprintf("Subject: SMR without testing pending validation Multibanda %s %s",
		m.Brand.Name, m.Device.CommercialModel)
	return subject, data
}

// NotifySmrValidationPending emails every internal user that an SMR without
// testing waits for validation.
func (s *multibandaService) NotifySmrValidationPending(m *responses.MultibandaExpanded) {
	if m == nil {
		return
	}
	to, empty := smrInternalRecipients(m.Company.ID)
	if empty {
		return
	}
	subject, data := BuildSmrPendingEmail(m)
	if err := sendSmrEmail(to, subject, data); err != nil {
		log.Printf("smr pending validation email (%s): %v", m.ID.Hex(), err)
	}
}
