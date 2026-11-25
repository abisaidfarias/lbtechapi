package services

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

// ---- stubs ----

type smrRepoStub struct {
	m       *responses.MultibandaExpanded
	source  *repositories.SmrSourceReport
	lastSet primitive.M
}

func (s *smrRepoStub) GetByIdExpanded(primitive.ObjectID) (*responses.MultibandaExpanded, error) {
	c := *s.m
	if s.m.SmrValidation != nil {
		v := *s.m.SmrValidation
		c.SmrValidation = &v
	}
	return &c, nil
}

func (s *smrRepoStub) ApplySmrDecision(_ primitive.ObjectID, set primitive.M, e models.SmrValidationEntry) (bool, error) {
	if s.m.SmrValidation.Status != models.SmrValidationPending {
		return false, nil
	}
	s.lastSet = set
	v := s.m.SmrValidation
	v.Status = set["smr_validation.status"].(string)
	at := set["smr_validation.decided_at"].(time.Time)
	v.DecidedAt = &at
	v.DecidedByName = set["smr_validation.decided_by_name"].(string)
	v.Comment = set["smr_validation.comment"].(string)
	v.Log = append(v.Log, e)
	s.m.CurrentPhase = set["current_phase"].(int)
	if st, ok := set["status"].(int); ok {
		s.m.Status = st
	}
	return true, nil
}

func (s *smrRepoStub) LastGeneratedReportForTechnicalModel(string, primitive.ObjectID) (*repositories.SmrSourceReport, error) {
	return s.source, nil
}

type smrDeviceStub struct{ set []bool }

func (d *smrDeviceStub) SetCommunicationRadio(_ primitive.ObjectID, v bool) error {
	d.set = append(d.set, v)
	return nil
}

type smrUserStub struct {
	user   *responses.User
	claims []responses.Claim
}

func (u smrUserStub) GetByID(string) (*responses.User, error) { return u.user, nil }
func (u smrUserStub) GetProfileByID(primitive.ObjectID) (*responses.Profile, error) {
	return &responses.Profile{Claims: u.claims}, nil
}

type smrNotifierStub struct {
	calls   int
	details functions.MultibandaReportEmailDetails
}

func (n *smrNotifierStub) NotifyReportGenerated(_ primitive.ObjectID, _ string, d functions.MultibandaReportEmailDetails) {
	n.calls++
	n.details = d
}

// ---- fixture ----

type smrFixture struct {
	s        *smrValidationService
	repo     *smrRepoStub
	devices  *smrDeviceStub
	notifier *smrNotifierStub
	id       string
}

func newSmrFixture(t *testing.T, deviceFM *bool, sourceStamp string) *smrFixture {
	t.Helper()
	off := false
	m := &responses.MultibandaExpanded{
		ID:              primitive.NewObjectID(),
		Type:            enums.MultibandaTypeSMR,
		CurrentPhase:    enums.HomologationPhase_value["PLANNING"],
		RequiresTesting: &off,
		SmrValidation:   &responses.MultibandaSmrValidation{Status: models.SmrValidationPending},
		Device:          responses.Device{ID: primitive.NewObjectID(), TechnicalModel: "2510ERA8BG", CommunicationRadio: deviceFM},
	}
	repo := &smrRepoStub{m: m}
	if sourceStamp != "" {
		date := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
		repo.source = &repositories.SmrSourceReport{
			ReportID: primitive.NewObjectID(), MultibandaID: primitive.NewObjectID(),
			Type: "initial", ReportDate: &date, StampCode: sourceStamp,
		}
	}
	f := &smrFixture{repo: repo, devices: &smrDeviceStub{}, notifier: &smrNotifierStub{}, id: m.ID.Hex()}
	f.s = &smrValidationService{
		multibandas: repo, devices: f.devices, notifier: f.notifier,
		users: smrUserStub{
			user:   &responses.User{ID: primitive.NewObjectID(), Name: "Ana", LastName: "Soto", IsInternal: true},
			claims: []responses.Claim{{Name: enums.CanReadMultibanda, Allow: true}, {Name: enums.CanWriteMultibanda, Allow: true}},
		},
		now: func() time.Time { return time.Date(2026, 11, 25, 15, 0, 0, 0, time.UTC) },
	}
	return f
}

func wantErr(t *testing.T, err error, msg string) {
	t.Helper()
	if !utils.IsValidationError(err) || err.Error() != msg {
		t.Fatalf("want %q, got %v", msg, err)
	}
}

// ---- tests ----

func TestSmrGetShowsReferenceData(t *testing.T) {
	on := true
	f := newSmrFixture(t, &on, "5g_and_below")
	got, err := f.s.Get(f.id, "u")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "pending" || got.FMRadio == nil || !*got.FMRadio || got.StampCode != "5g_and_below" ||
		got.LastReport == nil || got.LastReport.Type != "initial" || len(got.Stamps) != 8 || got.Log == nil {
		t.Fatalf("got %+v", got)
	}
}

func TestSmrApproveConfirmingValues(t *testing.T) {
	on := true
	f := newSmrFixture(t, &on, "5g_and_below")
	got, err := f.s.Approve(f.id, "u", &request.SmrValidationApprove{FMRadio: &on, StampCode: "5g_and_below"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "approved" || got.DecidedByName != "Ana Soto" || got.DecidedAt == nil || len(got.Log) != 1 {
		t.Fatalf("got %+v", got)
	}
	set := f.repo.lastSet
	if set["current_phase"] != enums.HomologationPhase_value["UNDER_EVALUATION"] ||
		!set["under_start_date"].(time.Time).Equal(*got.DecidedAt) {
		t.Fatalf("set = %v", set)
	}
	ps := set["phase_states"].(*models.MultibandaPhaseStates)
	if ps.Planning != "done" || ps.SampleReception != "skipped" || ps.Testing != "skipped" || ps.UnderEvaluation != "pending" {
		t.Fatalf("phase_states = %+v", ps)
	}
	if len(f.devices.set) != 0 {
		t.Errorf("device updated although FM Radio did not change")
	}
	if f.notifier.calls != 1 || !f.notifier.details.WithoutTesting || f.notifier.details.FMRadio != "Supported" ||
		f.notifier.details.StampCode != "5g_and_below" {
		t.Errorf("notifier = %+v", f.notifier)
	}
	log := got.Log[0]
	if log.Action != "approve" || log.StampOriginal != "5g_and_below" || log.StampFinal != "5g_and_below" ||
		log.SourceReportID == "" || *log.FMRadioOriginal != true || *log.FMRadioFinal != true {
		t.Errorf("log = %+v", log)
	}
}

func TestSmrApproveWithCorrections(t *testing.T) {
	on, off := true, false
	f := newSmrFixture(t, &on, "5g_and_below")
	_, err := f.s.Approve(f.id, "u", &request.SmrValidationApprove{FMRadio: &off, StampCode: "4g_and_below"})
	wantErr(t, err, functions.SmrErrCommentOnCorrection)

	got, err := f.s.Approve(f.id, "u", &request.SmrValidationApprove{
		FMRadio: &off, StampCode: "4g_and_below", Comment: "FM app removed in this build",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(f.devices.set) != 1 || f.devices.set[0] != false {
		t.Errorf("device FM Radio not updated: %v", f.devices.set)
	}
	if got.Comment != "FM app removed in this build" || got.Log[0].StampFinal != "4g_and_below" || *got.Log[0].FMRadioFinal {
		t.Errorf("got %+v", got)
	}
}

// Filling empty values (no previous report, FM Radio never set) is not a
// correction: no comment needed, and the device gets the value.
func TestSmrApproveFillingEmptyValues(t *testing.T) {
	f := newSmrFixture(t, nil, "")
	off := false
	got, err := f.s.Approve(f.id, "u", &request.SmrValidationApprove{FMRadio: &off, StampCode: "2g_3g"})
	if err != nil {
		t.Fatal(err)
	}
	if got.LastReport != nil || len(f.devices.set) != 1 || got.Log[0].FMRadioOriginal != nil {
		t.Fatalf("got %+v devices=%v", got, f.devices.set)
	}
}

func TestSmrApproveValidations(t *testing.T) {
	on := true
	f := newSmrFixture(t, &on, "5g_and_below")
	_, err := f.s.Approve(f.id, "u", &request.SmrValidationApprove{FMRadio: &on})
	wantErr(t, err, functions.SmrErrStampRequired)
	_, err = f.s.Approve(f.id, "u", &request.SmrValidationApprove{StampCode: "5g_and_below"})
	wantErr(t, err, functions.SmrErrFMRadioRequired)
}

func TestSmrRequiresInternalUserWithClaim(t *testing.T) {
	on := true
	f := newSmrFixture(t, &on, "")
	f.s.users = smrUserStub{user: &responses.User{ID: primitive.NewObjectID(), IsInternal: false},
		claims: []responses.Claim{{Name: enums.CanWriteMultibanda, Allow: true}}}
	_, err := f.s.Approve(f.id, "u", &request.SmrValidationApprove{FMRadio: &on, StampCode: "2g"})
	if !errors.Is(err, utils.ErrorForbidden) {
		t.Fatalf("external user: %v", err)
	}
	f.s.users = smrUserStub{user: &responses.User{ID: primitive.NewObjectID(), IsInternal: true},
		claims: []responses.Claim{{Name: enums.CanReadMultibanda, Allow: true}}}
	if _, err := f.s.Reject(f.id, "u", &request.SmrValidationReject{Comment: "x"}); !errors.Is(err, utils.ErrorForbidden) {
		t.Fatalf("internal without canWriteMultibanda: %v", err)
	}
}

func TestSmrRejectClosesTheRequest(t *testing.T) {
	on := true
	f := newSmrFixture(t, &on, "5g_and_below")
	_, err := f.s.Reject(f.id, "u", &request.SmrValidationReject{Comment: "  "})
	wantErr(t, err, functions.SmrErrCommentOnReject)

	got, err := f.s.Reject(f.id, "u", &request.SmrValidationReject{Comment: "Not only a security patch"})
	if err != nil {
		t.Fatal(err)
	}
	set := f.repo.lastSet
	if got.Status != "rejected" || set["current_phase"] != enums.HomologationPhase_value["COMPLETE"] ||
		set["status"] != enums.HomologationStatus_value["REJECTED"] ||
		!set["completed_date"].(time.Time).Equal(*got.DecidedAt) {
		t.Fatalf("got %+v set=%v", got, set)
	}
	ps := set["phase_states"].(*models.MultibandaPhaseStates)
	if *ps != *functions.SmrRejectedStates() || ps.Complete != "done" || ps.UnderEvaluation != "skipped" {
		t.Fatalf("phase_states = %+v", ps)
	}
	if f.notifier.calls != 0 {
		t.Error("reject must not send the report email")
	}
	// Closed: no second decision, and the normal flow cannot reopen it.
	_, err = f.s.Approve(f.id, "u", &request.SmrValidationApprove{FMRadio: &on, StampCode: "2g"})
	wantErr(t, err, functions.SmrErrNotPending)
	if err := functions.GuardSmrChange(got.Status, 4, 2, 3, 0); err == nil {
		t.Fatal("PhaseChange after rejection must fail")
	}
}

func TestSmrOnlyForRequestsWithoutTesting(t *testing.T) {
	on := true
	f := newSmrFixture(t, &on, "")
	f.repo.m.RequiresTesting = &on
	_, err := f.s.Get(f.id, "u")
	wantErr(t, err, functions.SmrErrNotWithoutTesting)
}

// The pending email goes to every internal user, links to the request and
// prints the skipped dates as N/A.
func TestSmrPendingEmailGoesToInternalUsers(t *testing.T) {
	var gotTo []string
	var gotSubject string
	var gotData functions.MultibandaPhaseEmailData
	origRecipients, origSend := smrInternalRecipients, sendSmrEmail
	defer func() { smrInternalRecipients, sendSmrEmail = origRecipients, origSend }()
	smrInternalRecipients = func(primitive.ObjectID) ([]string, bool) {
		return []string{"ana@lb.test", "camilo@lb.test"}, false
	}
	sendSmrEmail = func(to []string, subject string, data functions.MultibandaPhaseEmailData) error {
		gotTo, gotSubject, gotData = to, subject, data
		return nil
	}

	m := &responses.MultibandaExpanded{ID: primitive.NewObjectID(), Brand: responses.Brand{Name: "VIVO"},
		Device: responses.Device{CommercialModel: "Y29"}}
	(&multibandaService{}).NotifySmrValidationPending(m)

	if len(gotTo) != 2 || !strings.Contains(gotSubject, "SMR without testing pending validation") ||
		!strings.HasSuffix(gotData.ActionURL, "/multibanda?validate="+m.ID.Hex()) || gotData.TestStartDate != "N/A" ||
		gotData.SampleStartDate != "N/A" {
		t.Fatalf("to=%v subject=%q data=%+v", gotTo, gotSubject, gotData)
	}
	html, err := functions.RenderMultibandaPhaseEmailHTML(gotData, templatePathForServices(t))
	if err != nil || !strings.Contains(string(html), "Validate request") {
		t.Fatalf("render: %v", err)
	}
}

// templatePathForServices resolves the multibanda email template from the
// repository root (go test runs in the package directory).
func templatePathForServices(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, filepath.FromSlash(utils.TEMPLATE_MULTIBANDA_PHASE_PATH))
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repository root not found")
		}
		dir = parent
	}
}
