package services

import (
	"testing"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

func TestBuildReportPrefilledFMRadioFromDevice(t *testing.T) {
	yes, no := true, false
	cases := []struct {
		name  string
		radio *bool
		want  string
	}{
		{"undeclared", nil, ""},
		{"supported", &yes, enums.ReportFMRadioSupported},
		{"not supported", &no, enums.ReportFMRadioNotSupported},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mb := &responses.MultibandaExpanded{}
			mb.Device.CommunicationRadio = tc.radio
			if got := buildReportPrefilled(mb).FMRadio; got != tc.want {
				t.Fatalf("FMRadio = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFMRadioOptionsOnlySupportedAndNotSupported(t *testing.T) {
	want := []string{enums.ReportFMRadioSupported, enums.ReportFMRadioNotSupported}
	got := enums.ReportFMRadioOptions
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("fm_radio_options = %v, want %v", got, want)
	}
}

// The client's fm_radio is ignored before generation: the device value wins.
func TestSaveDraftStoresDeviceFMRadio(t *testing.T) {
	service, multibandaRepo, reportRepo := newReportServiceForTest("https://s3/report.pdf")
	yes := true
	multibandaRepo.multibanda.Device.CommunicationRadio = &yes

	req := buildGeneratableReport()
	req.DeviceInfo.FMRadio = enums.ReportFMRadioNotSupported
	if _, err := service.SaveDraft(testMultibandaHex, req, "user-1"); err != nil {
		t.Fatalf("save draft: %v", err)
	}
	if got := reportRepo.report.DeviceInfo.FMRadio; got != enums.ReportFMRadioSupported {
		t.Fatalf("stored fm_radio = %q, want %q", got, enums.ReportFMRadioSupported)
	}
}

func TestGenerateStoresDeviceFMRadioIgnoringClient(t *testing.T) {
	service, _, reportRepo := newReportServiceForTest("https://s3/report.pdf")

	req := buildGeneratableReport()
	req.DeviceInfo.FMRadio = enums.ReportFMRadioNA
	if _, err := service.Generate(testMultibandaHex, req, "user-1"); err != nil {
		t.Fatalf("generate: %v", err)
	}
	if got := reportRepo.report.DeviceInfo.FMRadio; got != enums.ReportFMRadioNotSupported {
		t.Fatalf("stored fm_radio = %q, want %q", got, enums.ReportFMRadioNotSupported)
	}
}

// The report email gets the stored FM Radio and stamp, i.e. what the PDF prints.
func TestGenerateNotifiesWithStoredFMRadioAndStamp(t *testing.T) {
	service, _, _ := newReportServiceForTest("https://s3/report.pdf")

	req := buildGeneratableReport()
	req.DeviceInfo.FMRadio = enums.ReportFMRadioSupported // ignored: device says false
	if _, err := service.Generate(testMultibandaHex, req, "user-1"); err != nil {
		t.Fatalf("generate: %v", err)
	}
	got := notifierOf(service).details
	if got.FMRadio != enums.ReportFMRadioNotSupported || got.StampCode != req.DeviceInfo.StampCode {
		t.Fatalf("notified details = %+v", got)
	}
}

func TestGenerateRejectsUndeclaredFMRadio(t *testing.T) {
	service, multibandaRepo, reportRepo := newReportServiceForTest("https://s3/report.pdf")
	multibandaRepo.multibanda.Device.CommunicationRadio = nil

	_, err := service.Generate(testMultibandaHex, buildGeneratableReport(), "user-1")
	if !utils.IsValidationError(err) || err.Error() != ReportFMRadioNotSetMessage {
		t.Fatalf("expected %q validation error, got %v", ReportFMRadioNotSetMessage, err)
	}
	if reportRepo.upsertCalls != 0 {
		t.Errorf("the report was saved %d times", reportRepo.upsertCalls)
	}
}

// Issued reports keep their stored value, even a legacy N/A, and even if the
// device changed or no longer declares FM Radio.
func TestGeneratedReportKeepsStoredFMRadio(t *testing.T) {
	service, multibandaRepo, reportRepo := newReportServiceForTest("https://s3/report.pdf")
	multibandaRepo.multibanda.Device.CommunicationRadio = nil
	reportRepo.report = &models.MultibandaReport{
		Status:     enums.MultibandaReportStatusGenerated,
		DeviceInfo: models.MultibandaReportDeviceInfo{FMRadio: enums.ReportFMRadioNA},
	}

	req := buildGeneratableReport()
	req.DeviceInfo.FMRadio = enums.ReportFMRadioSupported
	if _, err := service.Generate(testMultibandaHex, req, "user-1"); err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	if got := reportRepo.report.DeviceInfo.FMRadio; got != enums.ReportFMRadioNA {
		t.Fatalf("stored fm_radio = %q, want the issued %q", got, enums.ReportFMRadioNA)
	}
}

func TestGetFormShowsDeviceFMRadioForDraft(t *testing.T) {
	service, multibandaRepo, reportRepo := newReportServiceForTest("https://s3/report.pdf")
	yes := true
	multibandaRepo.multibanda.Device.CommunicationRadio = &yes
	reportRepo.report = &models.MultibandaReport{
		Status:     enums.MultibandaReportStatusDraft,
		DeviceInfo: models.MultibandaReportDeviceInfo{FMRadio: enums.ReportFMRadioNA},
	}

	form, err := service.GetForm(testMultibandaHex, "user-1")
	if err != nil {
		t.Fatalf("get form: %v", err)
	}
	if form.Prefilled.FMRadio != enums.ReportFMRadioSupported {
		t.Errorf("prefilled.fm_radio = %q", form.Prefilled.FMRadio)
	}
	if form.Saved.DeviceInfo.FMRadio != enums.ReportFMRadioSupported {
		t.Errorf("saved.device_info.fm_radio = %q", form.Saved.DeviceInfo.FMRadio)
	}
}
