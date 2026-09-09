package services

import (
	"testing"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ---- stubs ----

type stubReportRepo struct {
	report      *models.MultibandaReport
	markedURL   string
	upsertCalls int
}

func (s *stubReportRepo) GetByMultibanda(primitive.ObjectID) (*models.MultibandaReport, error) {
	return s.report, nil
}

func (s *stubReportRepo) Upsert(r *models.MultibandaReport) error {
	s.upsertCalls++
	s.report = r
	return nil
}

func (s *stubReportRepo) MarkGenerated(_ primitive.ObjectID, url, _ string, _ time.Time) error {
	s.markedURL = url
	return nil
}

type stubMultibandaRepo struct {
	multibanda *responses.MultibandaExpanded
	// linkedURL records what Generate wrote back onto the process.
	linkedURL string
	linkCalls int
	linkErr   error
}

func (s *stubMultibandaRepo) GetByIdExpanded(primitive.ObjectID) (*responses.MultibandaExpanded, error) {
	return s.multibanda, nil
}

func (s *stubMultibandaRepo) SetTestReportURL(_ primitive.ObjectID, url string) error {
	s.linkCalls++
	if s.linkErr != nil {
		return s.linkErr
	}
	s.linkedURL = url
	return nil
}

type stubDeviceRepo struct{}

func (stubDeviceRepo) GetById(string) (*responses.Device, error) { return &responses.Device{}, nil }

type stubUserRepo struct{ user *responses.User }

func (s stubUserRepo) GetByID(string) (*responses.User, error) { return s.user, nil }

// stubStorage returns a fixed URL so the test can assert the exact value that
// travels from the upload to the multibanda.
type stubStorage struct{ url string }

func (s stubStorage) UploadFileWithKeyAndName([]byte, string, string) (string, error) {
	return s.url, nil
}

// The report only uploads through UploadFileWithKeyAndName; the rest of the
// interface is here to satisfy it.
func (s stubStorage) UploadFile([]byte) (string, error) { return s.url, nil }

func (s stubStorage) UploadUserFile([]byte, string) (responses.UploadFileResponse, error) {
	return responses.UploadFileResponse{}, nil
}

func (s stubStorage) UploadFileWithKey([]byte, string) (string, error) { return s.url, nil }

// ---- fixtures ----

const testMultibandaHex = "652f1c2b8f1a4b0012a3d4e5"

// buildGeneratableReport fills every cell the generation validator demands for
// an SMR process: that type carries Device Information, SAE and FM Radio but
// not SIMLOCK or Multi-band, which keeps the fixture to the essentials.
func buildGeneratableReport() *request.MultibandaReportSave {
	testDate := time.Date(2026, 8, 28, 0, 0, 0, 0, time.UTC)

	req := &request.MultibandaReportSave{
		DeviceInfo: request.MultibandaReportDeviceInfo{
			CBSPackage:             "com.samsung.cmas 14.0.02",
			GooglePlaySystemUpdate: "01-may-2026",
			PreferredNetwork:       enums.ReportPreferredNetworkOptions[0],
			FMRadio:                enums.ReportFMRadioNotSupported,
			TestDate:               &testDate,
			IMEI:                   "862542070112758",
			SerialNumber:           "P16BB/W5UT00079",
			StampCode:              "5g_and_below",
		},
		CarriersTested: []string{"Entel"},
		SAEScenario:    enums.SAEScenarioLaboratory,
		Evidence: []request.MultibandaReportEvidence{
			{EvidenceType: enums.EvidenceSWVersion, URL: "http://127.0.0.1:1/sw.png"},
			{EvidenceType: enums.SAEEvidencePopUp, URL: "http://127.0.0.1:1/popup.png",
				Scenario: enums.SAEScenarioLaboratory, Operator: "Entel"},
			{EvidenceType: enums.SAEEvidenceCBHistory, URL: "http://127.0.0.1:1/cb.png",
				Scenario: enums.SAEScenarioLaboratory, Operator: "Entel"},
		},
	}

	for _, scenario := range enums.SAEScenariosFor(enums.SAEScenarioLaboratory) {
		for _, channel := range enums.SAEChannels {
			for _, test := range enums.ReportSAETests {
				for _, operator := range enums.SAEOperators {
					result := enums.ReportResultPass
					// Technology rows are N/A: the stub device declares no radios.
					if test.RequiresTechnology != "" {
						result = enums.ReportResultNA
					}
					req.SAEResults = append(req.SAEResults, request.MultibandaReportSAEResult{
						TestID: test.ID, Scenario: scenario, Channel: channel,
						Operator: operator, Result: result,
					})
				}
			}
		}
	}

	return req
}

func newReportServiceForTest(storageURL string) (*multibandaReportService, *stubMultibandaRepo, *stubReportRepo) {
	companyID := primitive.NewObjectID()

	multibandaRepo := &stubMultibandaRepo{
		multibanda: &responses.MultibandaExpanded{
			Type:    enums.MultibandaTypeSMR,
			Company: responses.Company{ID: companyID},
		},
	}
	reportRepo := &stubReportRepo{}

	service := &multibandaReportService{
		reportRepository:     reportRepo,
		multibandaRepository: multibandaRepo,
		deviceRepository:     stubDeviceRepo{},
		userRepository:       stubUserRepo{user: &responses.User{IsInternal: true, Company: companyID}},
		storageService:       stubStorage{url: storageURL},
	}
	return service, multibandaRepo, reportRepo
}

// ---- tests ----

// Generating the report must point the process at the PDF: the automatic
// report is the multibanda test report, so everything already reading
// test_report_url picks it up without knowing this module exists.
func TestGenerateLinksReportToMultibanda(t *testing.T) {
	service, multibandaRepo, _ := newReportServiceForTest("https://s3/report.pdf")

	generated, err := service.Generate(testMultibandaHex, buildGeneratableReport(), "user-1")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if multibandaRepo.linkCalls != 1 {
		t.Fatalf("expected the process to be linked once, got %d calls", multibandaRepo.linkCalls)
	}
	if multibandaRepo.linkedURL != generated.ReportURL {
		t.Errorf("linked url %q, want the generated url %q",
			multibandaRepo.linkedURL, generated.ReportURL)
	}
	if multibandaRepo.linkedURL == "" {
		t.Error("the process was linked to an empty url")
	}
}

// Regeneration cache-busts the URL, so the link has to be refreshed every time
// rather than only on the first generation.
func TestGenerateRelinksOnRegeneration(t *testing.T) {
	service, multibandaRepo, _ := newReportServiceForTest("https://s3/report.pdf")

	if _, err := service.Generate(testMultibandaHex, buildGeneratableReport(), "user-1"); err != nil {
		t.Fatalf("first generate: %v", err)
	}

	second, err := service.Generate(testMultibandaHex, buildGeneratableReport(), "user-1")
	if err != nil {
		t.Fatalf("regenerate: %v", err)
	}

	// Whether the two URLs differ is down to CacheBustCertificateURL, which
	// stamps whole seconds; what matters here is that the process is re-pointed
	// at whatever the latest generation produced.
	if multibandaRepo.linkCalls != 2 {
		t.Errorf("expected a link on each generation, got %d calls", multibandaRepo.linkCalls)
	}
	if multibandaRepo.linkedURL != second.ReportURL {
		t.Errorf("after regeneration the process points at %q, want %q",
			multibandaRepo.linkedURL, second.ReportURL)
	}
}

// A failure to link must surface: silently returning the URL would leave the
// process pointing at nothing while the caller believes it succeeded.
func TestGenerateReportsLinkFailure(t *testing.T) {
	service, multibandaRepo, _ := newReportServiceForTest("https://s3/report.pdf")
	multibandaRepo.linkErr = errNotLinked

	if _, err := service.Generate(testMultibandaHex, buildGeneratableReport(), "user-1"); err == nil {
		t.Fatal("expected an error when the process could not be linked")
	}
}

var errNotLinked = &linkError{}

type linkError struct{}

func (*linkError) Error() string { return "could not link" }
