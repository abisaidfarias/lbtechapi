package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ISmrValidationService validates SMR requests without testing: any internal
// user with canWriteMultibanda confirms FM Radio and the stamp (approve) or
// closes the request (reject). There is no assignment: whoever decides is
// recorded.
type ISmrValidationService interface {
	Get(multibandaID, userID string) (*responses.SmrValidation, error)
	Approve(multibandaID, userID string, req *request.SmrValidationApprove) (*responses.SmrValidation, error)
	Reject(multibandaID, userID string, req *request.SmrValidationReject) (*responses.SmrValidation, error)
}

type smrMultibandaRepo interface {
	GetByIdExpanded(primitive.ObjectID) (*responses.MultibandaExpanded, error)
	ApplySmrDecision(primitive.ObjectID, primitive.M, models.SmrValidationEntry) (bool, error)
	LastGeneratedReportForTechnicalModel(string, primitive.ObjectID) (*repositories.SmrSourceReport, error)
}

type smrDeviceRepo interface {
	SetCommunicationRadio(primitive.ObjectID, bool) error
}

type smrUserRepo interface {
	GetByID(string) (*responses.User, error)
	GetProfileByID(primitive.ObjectID) (*responses.Profile, error)
}

type smrValidationService struct {
	multibandas smrMultibandaRepo
	devices     smrDeviceRepo
	users       smrUserRepo
	notifier    reportGeneratedNotifier
	now         func() time.Time
}

func NewSmrValidationService(
	multibandas smrMultibandaRepo, devices smrDeviceRepo, users smrUserRepo, notifier reportGeneratedNotifier,
) ISmrValidationService {
	return &smrValidationService{multibandas: multibandas, devices: devices, users: users, notifier: notifier, now: time.Now}
}

// smrContext is what every action needs.
type smrContext struct {
	user       *responses.User
	multibanda *responses.MultibandaExpanded
	source     *repositories.SmrSourceReport
}

// load checks the caller (internal user with the claim), that the request is
// an SMR without testing, and finds the reference report.
func (s *smrValidationService) load(multibandaID, userID, claim string) (*smrContext, error) {
	user, err := s.users.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsInternal {
		return nil, fmt.Errorf("%w", utils.ErrorForbidden)
	}
	profile, err := s.users.GetProfileByID(user.ID)
	if err != nil {
		return nil, err
	}
	claims := make([]models.Claim, 0, len(profile.Claims))
	for _, c := range profile.Claims {
		claims = append(claims, models.Claim{Name: c.Name, Allow: c.Allow})
	}
	if !functions.HasProfileClaim(claims, claim) {
		return nil, fmt.Errorf("%w", utils.ErrorForbidden)
	}

	oid, err := primitive.ObjectIDFromHex(strings.TrimSpace(multibandaID))
	if err != nil {
		return nil, utils.NewValidationError("invalid multibanda id")
	}
	m, err := s.multibandas.GetByIdExpanded(oid)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, utils.NewValidationError("multibanda not found")
	}
	if !functions.IsSmrWithoutTesting(m.RequiresTesting) || m.SmrValidation == nil {
		return nil, utils.NewValidationError(functions.SmrErrNotWithoutTesting)
	}
	if !functions.UserHasClientAccess(user, m.Company.ID) {
		return nil, fmt.Errorf("%w", utils.ErrorForbidden)
	}
	source, err := s.multibandas.LastGeneratedReportForTechnicalModel(strings.TrimSpace(m.Device.TechnicalModel), m.ID)
	if err != nil {
		return nil, err
	}
	return &smrContext{user: user, multibanda: m, source: source}, nil
}

func (s *smrValidationService) Get(multibandaID, userID string) (*responses.SmrValidation, error) {
	ctx, err := s.load(multibandaID, userID, enums.CanReadMultibanda)
	if err != nil {
		return nil, err
	}
	return buildSmrValidationResponse(ctx), nil
}

func (s *smrValidationService) Approve(
	multibandaID, userID string, req *request.SmrValidationApprove,
) (*responses.SmrValidation, error) {
	ctx, err := s.load(multibandaID, userID, enums.CanWriteMultibanda)
	if err != nil {
		return nil, err
	}
	if ctx.multibanda.SmrValidation.Status != models.SmrValidationPending {
		return nil, utils.NewValidationError(functions.SmrErrNotPending)
	}
	if req == nil {
		req = &request.SmrValidationApprove{}
	}

	originalFM := ctx.multibanda.Device.CommunicationRadio
	originalStamp := sourceStamp(ctx.source)
	if err := functions.SmrApproveCheck(originalFM, originalStamp, req.FMRadio, req.StampCode, req.Comment); err != nil {
		return nil, err
	}
	stamp := strings.TrimSpace(req.StampCode)
	comment := strings.TrimSpace(req.Comment)

	// Filling or correcting FM Radio updates the device.
	if originalFM == nil || *originalFM != *req.FMRadio {
		if err := s.devices.SetCommunicationRadio(ctx.multibanda.Device.ID, *req.FMRadio); err != nil {
			return nil, err
		}
	}

	now := s.now().UTC()
	name := strings.TrimSpace(ctx.user.Name + " " + ctx.user.LastName)
	set := primitive.M{
		"smr_validation.status":          models.SmrValidationApproved,
		"smr_validation.decided_at":      now,
		"smr_validation.decided_by":      ctx.user.ID,
		"smr_validation.decided_by_name": name,
		"smr_validation.comment":         comment,
		"current_phase":                  enums.HomologationPhase_value["UNDER_EVALUATION"],
		"dashboard_phase":                enums.DashBoardPhase_value["ONGOING"],
		"under_start_date":               now,
		"phase_states":                   functions.SmrApprovedStates(),
	}
	entry := smrLogEntry("approve", now, ctx, name, comment, originalFM, req.FMRadio, originalStamp, stamp)
	if ok, err := s.multibandas.ApplySmrDecision(ctx.multibanda.ID, set, entry); err != nil {
		return nil, err
	} else if !ok {
		return nil, utils.NewValidationError(functions.SmrErrNotPending)
	}

	if s.notifier != nil {
		s.notifier.NotifyReportGenerated(ctx.multibanda.ID, userID, functions.MultibandaReportEmailDetails{
			FMRadio:        enums.ReportFMRadioFromDevice(req.FMRadio),
			StampCode:      stamp,
			WithoutTesting: true,
		})
	}
	return s.Get(multibandaID, userID)
}

func (s *smrValidationService) Reject(
	multibandaID, userID string, req *request.SmrValidationReject,
) (*responses.SmrValidation, error) {
	ctx, err := s.load(multibandaID, userID, enums.CanWriteMultibanda)
	if err != nil {
		return nil, err
	}
	if ctx.multibanda.SmrValidation.Status != models.SmrValidationPending {
		return nil, utils.NewValidationError(functions.SmrErrNotPending)
	}
	comment := ""
	if req != nil {
		comment = strings.TrimSpace(req.Comment)
	}
	if comment == "" {
		return nil, utils.NewValidationError(functions.SmrErrCommentOnReject)
	}

	now := s.now().UTC()
	name := strings.TrimSpace(ctx.user.Name + " " + ctx.user.LastName)
	set := primitive.M{
		"smr_validation.status":          models.SmrValidationRejected,
		"smr_validation.decided_at":      now,
		"smr_validation.decided_by":      ctx.user.ID,
		"smr_validation.decided_by_name": name,
		"smr_validation.comment":         comment,
		// Closed like the existing rejections: phase COMPLETE + status REJECTED.
		"current_phase":   enums.HomologationPhase_value["COMPLETE"],
		"status":          enums.HomologationStatus_value["REJECTED"],
		"dashboard_phase": enums.DashBoardPhase_value["COMPLETE"],
		"completed_date":  now,
		"phase_states":    functions.SmrRejectedStates(),
	}
	fm := ctx.multibanda.Device.CommunicationRadio
	stamp := sourceStamp(ctx.source)
	entry := smrLogEntry("reject", now, ctx, name, comment, fm, fm, stamp, stamp)
	if ok, err := s.multibandas.ApplySmrDecision(ctx.multibanda.ID, set, entry); err != nil {
		return nil, err
	} else if !ok {
		return nil, utils.NewValidationError(functions.SmrErrNotPending)
	}
	return s.Get(multibandaID, userID)
}

func sourceStamp(source *repositories.SmrSourceReport) string {
	if source == nil {
		return ""
	}
	return strings.TrimSpace(source.StampCode)
}

func smrLogEntry(
	action string, at time.Time, ctx *smrContext, name, comment string,
	fmOriginal, fmFinal *bool, stampOriginal, stampFinal string,
) models.SmrValidationEntry {
	entry := models.SmrValidationEntry{
		Action: action, At: at, By: ctx.user.ID, ByName: name, Comment: comment,
		FMRadioOriginal: copyBool(fmOriginal), FMRadioFinal: copyBool(fmFinal),
		StampOriginal: stampOriginal, StampFinal: stampFinal,
	}
	if ctx.source != nil {
		entry.SourceReportID = ctx.source.ReportID
		entry.SourceMultibanda = ctx.source.MultibandaID
	}
	return entry
}

func copyBool(v *bool) *bool {
	if v == nil {
		return nil
	}
	c := *v
	return &c
}

func buildSmrValidationResponse(ctx *smrContext) *responses.SmrValidation {
	v := ctx.multibanda.SmrValidation
	out := &responses.SmrValidation{
		Status:        v.Status,
		FMRadio:       copyBool(ctx.multibanda.Device.CommunicationRadio),
		StampCode:     sourceStamp(ctx.source),
		Stamps:        enums.ReportStampCatalog,
		DecidedAt:     v.DecidedAt,
		DecidedByName: v.DecidedByName,
		Comment:       v.Comment,
		Log:           []responses.SmrValidationLogEntry{},
	}
	if ctx.source != nil {
		out.LastReport = &responses.SmrLastReport{
			ID:           ctx.source.ReportID.Hex(),
			MultibandaID: ctx.source.MultibandaID.Hex(),
			Type:         ctx.source.Type,
			ReportDate:   ctx.source.ReportDate,
			StampCode:    ctx.source.StampCode,
		}
	}
	for _, e := range v.Log {
		item := responses.SmrValidationLogEntry{
			Action: e.Action, At: e.At, ByName: e.ByName, Comment: e.Comment,
			FMRadioOriginal: e.FMRadioOriginal, FMRadioFinal: e.FMRadioFinal,
			StampOriginal: e.StampOriginal, StampFinal: e.StampFinal,
		}
		if !e.SourceReportID.IsZero() {
			item.SourceReportID = e.SourceReportID.Hex()
		}
		out.Log = append(out.Log, item)
	}
	return out
}
