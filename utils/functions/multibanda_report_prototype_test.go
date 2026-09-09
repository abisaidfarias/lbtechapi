package functions_test

import (
	"os"
	"testing"
	"time"

	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
)

// TestBuildMultibandaReportPrototype renders a fully populated report so the
// layout can be eyeballed. It only writes a file when MBR_PROTOTYPE_OUT is set,
// so a normal `go test ./...` just exercises the renderer end to end.
func TestBuildMultibandaReportPrototype(t *testing.T) {
	stampCode := utils.EnvOrDefault("MBR_PROTOTYPE_STAMP", "5g_and_below")
	stamp, ok := enums.StampByCode(stampCode)
	if !ok {
		t.Fatalf("unknown stamp code %q", stampCode)
	}
	stampImage := utils.MultibandaStampImage(stamp.ImageKey)
	if len(stampImage) == 0 {
		t.Fatalf("no artwork embedded for stamp %q (image key %q)", stampCode, stamp.ImageKey)
	}

	screenshot := samplePhoneScreenshot(t)

	data := functions.MultibandaReportPDFData{
		ProcessTypeLabel:       "Initial Process",
		ReportDate:             "06-sep-2026",
		Manufacturer:           "Samsung",
		CommercialName:         "Galaxy A16 5G",
		TechnicalModel:         "SM-A166M/DSN",
		HardwareVersion:        "REV 0.3",
		SoftwareVersion:        "A166MUBU1AXH8",
		SARValue:               "0.812 W/kg",
		CBSPackage:             "com.samsung.cmas 14.0.02",
		GooglePlaySystemUpdate: "01-may-2026",
		OperativeSystemType:    "Android 14",
		PreferredNetwork:       "5G/4G/3G/2G",
		FMRadio:                enums.ReportFMRadioSupported,
		TestDate:               "28-ago-2026",
		IMEI:                   "862542070112758",
		SerialNumber:           "P16BB/W5UT00079",

		StampImage:     stampImage,
		StampLabel:     stamp.Label,
		SWVersionPhoto: screenshot,

		CarriersTested:    prototypeCarrierNames(),
		IncludesSimlock:   true,
		IncludesMultiband: true,

		SimlockRows: prototypeSimlockRows(),
		BandGroups:  prototypeBandGroups(),
		SAEScenario: "Laboratory and SAE Room",
		SAEBlocks:   prototypeSAEBlocks(),
		SAEEvidence: []functions.MultibandaReportEvidenceItem{
			{Label: "Pop-Up", ScenarioLabel: "Laboratory", Operator: "Entel", Image: screenshot},
			{Label: "CB History", ScenarioLabel: "SAE Room", Operator: "Movistar", Image: screenshot},
		},

		FMRadioTestApplies: true,
		FMRadioTestResult:  enums.ReportResultPass,
		GeneratedAt:        time.Date(2026, 9, 6, 10, 30, 0, 0, time.UTC),
		Year:               2026,
	}

	pdf, err := functions.BuildMultibandaReportPDF(data)
	if err != nil {
		t.Fatalf("build report: %v", err)
	}
	if len(pdf) == 0 {
		t.Fatal("renderer produced no bytes")
	}

	out := os.Getenv("MBR_PROTOTYPE_OUT")
	if out == "" {
		return
	}
	if err := os.WriteFile(out, pdf, 0o644); err != nil {
		t.Fatalf("write %s: %v", out, err)
	}
	t.Logf("wrote %s (%d KB)", out, len(pdf)/1024)
}

// samplePhoneScreenshot reuses one of the stamp assets as stand-in artwork so
// the prototype does not depend on files outside the repository.
func samplePhoneScreenshot(t *testing.T) []byte {
	t.Helper()
	if data := utils.MultibandaStampImage("2G_3G"); len(data) > 0 {
		return data
	}
	t.Fatal("no embedded image available to stand in for a screenshot")
	return nil
}

func prototypeCarrierNames() []string {
	names := make([]string, 0, len(enums.ReportSimlockCarriers))
	for _, c := range enums.ReportSimlockCarriers {
		names = append(names, c.Name)
	}
	return names
}

func prototypeSimlockRows() []functions.MultibandaReportSimlockRow {
	var rows []functions.MultibandaReportSimlockRow
	i := 0
	for _, test := range enums.ReportSimlockTests {
		for _, carrier := range enums.ReportSimlockCarriers {
			row := functions.MultibandaReportSimlockRow{
				TestID: test.ID, Name: test.Name,
				Carrier: carrier.Name, MNC: carrier.MNC,
				Result: enums.ReportResultPass,
			}
			if i == 5 {
				row.Result = enums.ReportResultFail
				row.Comment = "The handset kept the carrier lock after the reset; escalated to the manufacturer."
			}
			rows = append(rows, row)
			i++
		}
	}
	return rows
}

func prototypeBandGroups() []functions.MultibandaReportBandGroup {
	var groups []functions.MultibandaReportBandGroup
	for _, tech := range enums.ReportBandCatalog {
		group := functions.MultibandaReportBandGroup{Technology: tech.Label}
		for i, band := range tech.Bands {
			row := functions.MultibandaReportBandRow{Band: band, Result: enums.ReportResultOK}
			switch {
			case tech.Code == "nr" && i == 3:
				row.Result = enums.ReportResultNotSupported
			case tech.Code == "lte" && i == 4:
				row.Result = enums.ReportResultNoOK
				row.Comment = "No attach on B7; the device falls back to B4."
			}
			group.Bands = append(group.Bands, row)
		}
		groups = append(groups, group)
	}
	return groups
}

func prototypeSAEBlocks() []functions.MultibandaReportSAEBlock {
	var blocks []functions.MultibandaReportSAEBlock
	for _, scenario := range enums.SAEScenariosFor(enums.SAEScenarioBoth) {
		for _, channel := range enums.SAEChannels {
			block := functions.MultibandaReportSAEBlock{
				ScenarioLabel: enums.SAEScenarioLabels[scenario],
				Channel:       channel,
				Operators:     enums.SAEOperators,
			}
			for i, test := range enums.ReportSAETests {
				row := functions.MultibandaReportSAERow{TestID: test.ID, Name: test.Name}
				for range enums.SAEOperators {
					row.Results = append(row.Results, enums.ReportResultPass)
					row.Comments = append(row.Comments, "")
				}
				if test.RequiresTechnology == enums.SAETechUMTS {
					for j := range row.Results {
						row.Results[j] = enums.ReportResultNA
					}
				}
				if i == 6 && channel == "4370" {
					row.Results[1] = enums.ReportResultFail
					row.Comments[1] = "No vibration on channel 4370 with Movistar."
				}
				block.Rows = append(block.Rows, row)
			}
			blocks = append(blocks, block)
		}
	}
	return blocks
}
