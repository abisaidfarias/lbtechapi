package services

import (
	"time"

	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

// IDashboardService is the dashboard service
type IDashboardService interface {
	Get(string) (*responses.DashboardChart, error)
	GetGeneralInfo(string) (*responses.DashboardInfo, error)
}

type dashboardService struct {
	homologationRepository repositories.IHomologationRepository
	userRepository         repositories.IUserRepository
}

// NewDashboardService is a constructor
func NewDashboardService(homologationRepository repositories.IHomologationRepository,
	userRepository repositories.IUserRepository) IDashboardService {
	return &dashboardService{
		homologationRepository: homologationRepository,
		userRepository:         userRepository,
	}
}
func (s *dashboardService) Get(userID string) (*responses.DashboardChart, error) {
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	chartTypeCountry, err := s.homologationRepository.GetGroupedByTypeCountry(user.Clients,
		user.Brands, user.Countries,user.Company,user.IsInternal)
	if err != nil {
		return nil, err
	}
	return mapping.TypeCountriesToTypeCountriesCharts(chartTypeCountry), nil
}
func (s *dashboardService) GetGeneralInfo(userID string) (*responses.DashboardInfo, error) {
	user, err := s.userRepository.GetByIDExpanded(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	t := time.Now()
	dashboardTotals, err := s.homologationRepository.GetByCompanyStatusGrouped(user.Company.ID)
	if err != nil {
		return nil, nil
	}
	response := new(responses.DashboardInfo)
	response.CompanyName = user.Company.Name
	response.LogoImage = user.Company.LogoUrl
	response.Month = t.Month().String()
	for _, dashboardTotal := range dashboardTotals {
		if dashboardTotal.CurrentPhase == enums.HomologationPhase_value["PLANNING"] {
			response.TotalPlanning = dashboardTotal.Count
			response.TotalOngoing += dashboardTotal.Count
		} else if dashboardTotal.CurrentPhase == enums.HomologationPhase_value["SAMPLE_RECEPTION"] {
			response.TotalSampleReception = dashboardTotal.Count
			response.TotalOngoing += dashboardTotal.Count
		} else if dashboardTotal.CurrentPhase == enums.HomologationPhase_value["COMPLETE"] {
			response.TotalFinished = dashboardTotal.Count
		} else {
			response.TotalOngoing += dashboardTotal.Count
		}
	}
	return response, nil
}
