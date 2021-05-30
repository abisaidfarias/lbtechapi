package services

import (
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

// IKpiService is the kpi service
type IKpiService interface {
	GetVolumeChart(string) (*responses.VolumeChart, error)
}

type kpiService struct {
	homologationRepository repositories.IHomologationRepository
	userRepository         repositories.IUserRepository
}

// NewKpiService is a constructor
func NewKpiService(homologationRepository repositories.IHomologationRepository,
	userRepository repositories.IUserRepository) IKpiService {
	return &kpiService{
		homologationRepository: homologationRepository,
		userRepository:         userRepository,
	}
}
func (s *kpiService) GetVolumeChart(userID string) (*responses.VolumeChart, error) {
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	chartBrandCountry, err := s.homologationRepository.GetGroupedByBrandCountry(user.Clients,
		user.Brands, user.Countries)
	if err != nil {
		return nil, err
	}
	chartTypeBrand, err := s.homologationRepository.GetGroupedByBrandType(user.Clients,
		user.Brands, user.Countries)
	if err != nil {
		return nil, err
	}
	volumeCountriesBrand := mapping.VolumeCountriesToVolumeCountriesCharts(chartBrandCountry)
	volumeTypeBrand := mapping.VolumeBrandToVolumeBrandCharts(chartTypeBrand)
	var volumeChart responses.VolumeChart
	volumeChart.CountryBrandChart = volumeCountriesBrand
	volumeChart.TypeBrandChart = volumeTypeBrand

	return &volumeChart, nil
}
