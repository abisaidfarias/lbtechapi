package services

import (
	"github.com/abisaidfarias/lbtechapi/repositories"
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
		user.Brands, user.Countries)
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
	response := new(responses.DashboardInfo)
	response.CompanyName = user.Company.Name
	response.LogoImage = "https://lbtechfilestorage.blob.core.windows.net/images/5708700001266629581.jpg"
	response.TotalFinished = 10
	response.TotalOngoing = 6
	response.TotalPlanning = 8
	response.TotalSampleReception = 9
	response.Month = "September"

	return response, nil
}
