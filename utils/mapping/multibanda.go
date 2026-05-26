package mapping

import (
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func MultibandaRequestToMultibanda(
	multibanda *request.Multibanda,
	companyID primitive.ObjectID,
	deviceID primitive.ObjectID,
	brandID primitive.ObjectID,
) *models.Multibanda {
	return &models.Multibanda{
		Company:           companyID,
		Device:            deviceID,
		Brand:             brandID,
		SoftwareVersion:   multibanda.SoftwareVersion,
		HardwareVersion:   multibanda.HardwareVersion,
		OsVersion:         multibanda.OsVersion,
		OsVersionView:     multibanda.OsVersionView,
		Type:              multibanda.Type,
		EvaluationType:    append([]string(nil), multibanda.EvaluationType...),
		CurrentPhase:      enums.MultibandaPhasePlanning,
		PlanningDate:            multibanda.PlanningDate,
		SampleStartDate:         nil,
		SampleEndDate:           nil,
		TestReportUrl:           multibanda.TestReportUrl,
		MultibandCertificateUrl: multibanda.MultibandCertificateUrl,
		Status:                  enums.HomologationStatus_value["IN_PROGRESS"],
		DashBoardPhase:    enums.MultibandaPhasePlanning,
		IsInternalProject: false,
		CreatedDate:       time.Now(),
	}
}

func MultibandaRequestToMultibandaResume(multibanda *request.MultibandaResume) *models.Multibanda {
	return &models.Multibanda{
		SoftwareVersion:   multibanda.SoftwareVersion,
		HardwareVersion:   multibanda.HardwareVersion,
		OsVersion:         multibanda.OsVersion,
		OsVersionView:     multibanda.OsVersionView,
		CurrentPhase:      multibanda.CurrentPhase,
		PlanningDate:      multibanda.PlanningDate,
		SampleStartDate:   timeToOptionalPtr(multibanda.SampleStartDate),
		SampleEndDate:     timeToOptionalPtr(multibanda.SampleEndDate),
		TestStartDate:     multibanda.TestStartDate,
		TestEndDate:       multibanda.TestEndDate,
		UnderStartDate:    multibanda.UnderStartDate,
		UnderEndDate:      multibanda.UnderEndDate,
		CompletedDate:           multibanda.CompletedDate,
		Status:                  multibanda.Status,
		TestReportUrl:           multibanda.TestReportUrl,
		MultibandCertificateUrl: multibanda.MultibandCertificateUrl,
		IsInternalProject:       multibanda.IsInternalProject,
	}
}

func MultibandaResponseToMultibandaNotify(multibanda responses.MultibandaExpanded) request.MultibandaNotify {
	return request.MultibandaNotify{
		Company:                 multibanda.Company.ID.Hex(),
		CompanyName:             multibanda.Company.Name,
		Device:                  multibanda.Device.ID.Hex(),
		Brand:                   multibanda.Brand.ID.Hex(),
		SoftwareVersion:         multibanda.SoftwareVersion,
		HardwareVersion:         multibanda.HardwareVersion,
		OsVersion:               multibanda.OsVersion,
		OsVersionView:           multibanda.OsVersionView,
		Type:                    multibanda.Type,
		EvaluationType:          append([]string(nil), multibanda.EvaluationType...),
		CurrentPhase:            multibanda.CurrentPhase,
		PlanningDate:            derefTime(multibanda.PlanningDate),
		SampleStartDate:         derefTime(multibanda.SampleStartDate),
		SampleEndDate:           derefTime(multibanda.SampleEndDate),
		TestStartDate:           derefTime(multibanda.TestStartDate),
		TestEndDate:             derefTime(multibanda.TestEndDate),
		UnderStartDate:          derefTime(multibanda.UnderStartDate),
		UnderEndDate:            derefTime(multibanda.UnderEndDate),
		CompletedDate:           derefTime(multibanda.CompletedDate),
		Status:                  multibanda.Status,
		TestReportUrl:           multibanda.TestReportUrl,
		MultibandCertificateUrl: multibanda.MultibandCertificateUrl,
		IsInternalProject:       multibanda.IsInternalProject,
	}
}

func timeToOptionalPtr(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}

func derefTime(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}
