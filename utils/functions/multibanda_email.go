package functions

import (
	"fmt"
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

func FormatMultibandaTypeLabel(value string) string {
	if label, ok := enums.MultibandaTypeLabels[value]; ok {
		return label
	}
	return value
}

func FormatMultibandaEvaluationTypes(values []string) string {
	if len(values) == 0 {
		return "—"
	}

	labels := make([]string, 0, len(values))
	for _, value := range values {
		if label, ok := enums.MultibandaEvaluationTypeLabels[value]; ok {
			labels = append(labels, label)
			continue
		}
		labels = append(labels, value)
	}
	return strings.Join(labels, ", ")
}

func MultibandaPhaseLabel(phase int) string {
	if label, ok := enums.MultibandaPhaseLabels[phase]; ok {
		return label
	}
	return fmt.Sprintf("Phase %d", phase)
}

func FormatMultibandaEmailDate(date time.Time) string {
	if date.IsZero() {
		return "N/A"
	}
	return fmt.Sprintf("%02d/%02d/%d", date.Day(), date.Month(), date.Year())
}

func FormatMultibandaOsVersion(osVersionView, platformOs, osVersion string) string {
	platformOs = strings.TrimSpace(platformOs)
	osVersionView = strings.TrimSpace(osVersionView)
	osVersion = strings.TrimSpace(osVersion)

	version := osVersionView
	if version == "" {
		version = osVersion
	}

	if version == "" {
		return "—"
	}

	if platformOs == "" {
		return version
	}

	if strings.HasPrefix(strings.ToLower(version), strings.ToLower(platformOs)) {
		return version
	}

	return platformOs + " " + version
}

func FormatMultibandaSarValue(value float64) string {
	if value == 0 {
		return "—"
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), ".")
}

func BuildMultibandaPhaseEmailData(
	multibanda *request.MultibandaNotify,
	brandName string,
	technicalModel string,
	commercialModel string,
	platformOs string,
	sarValue float64,
	userName string,
	mainMessage string,
	finished bool,
	decision string,
) MultibandaPhaseEmailData {
	projectType := "External"
	if multibanda.IsInternalProject {
		projectType = "Internal"
	}

	now := time.Now()
	clientName := strings.TrimSpace(multibanda.CompanyName)
	if clientName == "" {
		clientName = "Client"
	}

	return MultibandaPhaseEmailData{
		ClientName:              clientName,
		MainMessage:             mainMessage,
		NotificationDate:        FormatMultibandaEmailDate(now),
		CurrentPhase:            MultibandaPhaseLabel(multibanda.CurrentPhase),
		ProjectType:             projectType,
		ProcessType:             FormatMultibandaTypeLabel(multibanda.Type),
		EvaluationTypes:         FormatMultibandaEvaluationTypes(multibanda.EvaluationType),
		Brand:                   brandName,
		TechnicalModel:          technicalModel,
		CommercialModel:         commercialModel,
		SoftwareVersion:         emptyAsDash(multibanda.SoftwareVersion),
		HardwareVersion:         emptyAsDash(multibanda.HardwareVersion),
		OsVersion:               FormatMultibandaOsVersion(multibanda.OsVersionView, platformOs, multibanda.OsVersion),
		SARValue:                FormatMultibandaSarValue(sarValue),
		UpdatedBy:               emptyAsDash(userName),
		PlanningDate:            FormatMultibandaEmailDate(multibanda.PlanningDate),
		SampleStartDate:         FormatMultibandaEmailDate(multibanda.SampleStartDate),
		SampleEndDate:           FormatMultibandaEmailDate(multibanda.SampleEndDate),
		TestStartDate:           FormatMultibandaEmailDate(multibanda.TestStartDate),
		TestEndDate:             FormatMultibandaEmailDate(multibanda.TestEndDate),
		UnderStartDate:          FormatMultibandaEmailDate(multibanda.UnderStartDate),
		UnderEndDate:            FormatMultibandaEmailDate(multibanda.UnderEndDate),
		ResultDate:              FormatMultibandaEmailDate(multibanda.CompletedDate),
		Finished:                finished,
		Decision:                emptyAsDash(decision),
		TestReportURL:           strings.TrimSpace(multibanda.TestReportUrl),
		MultibandCertificateURL: strings.TrimSpace(multibanda.MultibandCertificateUrl),
		Year:                    now.Year(),
	}
}

func emptyAsDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return strings.TrimSpace(value)
}
