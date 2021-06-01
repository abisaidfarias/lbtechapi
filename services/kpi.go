package services

import (
	"time"

	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

// IKpiService is the kpi service
type IKpiService interface {
	GetVolumeChart(string, time.Time, time.Time) (*responses.VolumeChart, error)
	GetTimeChart(string, time.Time, time.Time) (*responses.TimeChart, error)
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
func (s *kpiService) GetVolumeChart(userID string, startDate time.Time,
	endDate time.Time) (*responses.VolumeChart, error) {
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
func (s *kpiService) GetTimeChart(userID string, startDate time.Time,
	endDate time.Time) (*responses.TimeChart, error) {
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	var homologations []*responses.HomologationExpanded = []*responses.HomologationExpanded{}

	if user.IsInternal {
		homologations, err = s.homologationRepository.GetByInternal(user.Clients,
			user.Brands, user.Countries)
		if err != nil {
			return nil, err
		}
	} else {
		homologations, err = s.homologationRepository.GetByExternal(user.Company,
			user.Brands, user.Countries)
		if err != nil {
			return nil, err
		}
	}
	countryTestingMap := make(map[string][3]float64)
	countryTestingCountMap := make(map[string][3]int)
	for _, tc := range homologations {
		if tc.Type == enums.HomologationType_value["INITIAL"] {
			if tc.TestStartDate != nil && tc.TestEndDate != nil {
				totalTime := countryTestingMap[tc.Country.Name[0:3]]
				totalTime[0] += tc.TestEndDate.Sub(*tc.TestStartDate).Hours() / 24
				countryTestingMap[tc.Country.Name[0:3]] = totalTime
				count := countryTestingCountMap[tc.Country.Name[0:3]]
				count[0]++
				countryTestingCountMap[tc.Country.Name[0:3]] = count
			}
		} else if tc.Type == enums.HomologationType_value["MAINTENANCE"] {
			if tc.TestStartDate != nil && tc.TestEndDate != nil {
				totalTime := countryTestingMap[tc.Country.Name[0:3]]
				totalTime[1] += tc.TestEndDate.Sub(*tc.TestStartDate).Hours() / 24
				countryTestingMap[tc.Country.Name[0:3]] = totalTime
				count := countryTestingCountMap[tc.Country.Name[0:3]]
				count[1]++
				countryTestingCountMap[tc.Country.Name[0:3]] = count
			}
		} else {
			if tc.TestStartDate != nil && tc.TestEndDate != nil {
				totalTime := countryTestingMap[tc.Country.Name[0:3]]
				totalTime[2] += tc.TestEndDate.Sub(*tc.TestStartDate).Hours() / 24
				countryTestingMap[tc.Country.Name[0:3]] = totalTime
				count := countryTestingCountMap[tc.Country.Name[0:3]]
				count[2]++
				countryTestingCountMap[tc.Country.Name[0:3]] = count
			}
		}

	}
	countryTestingFinalMap := make(map[string][3]float64)
	for key, value := range countryTestingMap {
		averageTime := countryTestingFinalMap[key]
		if value[0] > 0 {
			averageTime[0] = value[0] / float64(countryTestingCountMap[key][0])
		}
		if value[1] > 0 {
			averageTime[1] = value[1] / float64(countryTestingCountMap[key][1])
		}
		if value[2] > 0 {
			averageTime[2] = value[2] / float64(countryTestingCountMap[key][2])
		}

		countryTestingFinalMap[key] = averageTime
	}
	timeChart := new(responses.TimeChart)
	timeChart.CountryTestingChart = countryTestingFinalMap

	return timeChart, nil
}
