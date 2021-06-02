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
	countryUnderMap := make(map[string][3]float64)
	countryUnderCountMap := make(map[string][3]int)
	brandTestingMap := make(map[string][3]float64)
	brandTestingCountMap := make(map[string][3]int)
	brandUnderMap := make(map[string][3]float64)
	brandUnderCountMap := make(map[string][3]int)
	for _, tc := range homologations {
		if tc.Type == enums.HomologationType_value["INITIAL"] {
			if tc.TestStartDate != nil && tc.TestEndDate != nil {
				countryTotalTime := countryTestingMap[tc.Country.Name[0:3]]
				countryTotalTime[0] += tc.TestEndDate.Sub(*tc.TestStartDate).Hours() / 24
				countryTestingMap[tc.Country.Name[0:3]] = countryTotalTime
				countryCount := countryTestingCountMap[tc.Country.Name[0:3]]
				countryCount[0]++
				countryTestingCountMap[tc.Country.Name[0:3]] = countryCount

				brandTotalTime := brandTestingMap[tc.Brand.Name]
				brandTotalTime[0] += tc.TestEndDate.Sub(*tc.TestStartDate).Hours() / 24
				brandTestingMap[tc.Brand.Name] = brandTotalTime
				brandCount := brandTestingCountMap[tc.Brand.Name]
				brandCount[0]++
				brandTestingCountMap[tc.Brand.Name] = brandCount
			}
			if tc.UnderStartDate != nil && tc.UnderEndDate != nil {
				countryTotalTime := countryUnderMap[tc.Country.Name[0:3]]
				countryTotalTime[0] += tc.UnderEndDate.Sub(*tc.UnderStartDate).Hours() / 24
				countryUnderMap[tc.Country.Name[0:3]] = countryTotalTime
				countryCount := countryUnderCountMap[tc.Country.Name[0:3]]
				countryCount[0]++
				countryUnderCountMap[tc.Country.Name[0:3]] = countryCount

				brandTotalTime := brandUnderMap[tc.Brand.Name]
				brandTotalTime[0] += tc.UnderEndDate.Sub(*tc.UnderStartDate).Hours() / 24
				brandUnderMap[tc.Brand.Name] = brandTotalTime
				brandCount := brandUnderCountMap[tc.Brand.Name]
				brandCount[0]++
				brandUnderCountMap[tc.Brand.Name] = brandCount
			}
		} else if tc.Type == enums.HomologationType_value["MAINTENANCE"] {
			if tc.TestStartDate != nil && tc.TestEndDate != nil {
				countryTotalTime := countryTestingMap[tc.Country.Name[0:3]]
				countryTotalTime[1] += tc.TestEndDate.Sub(*tc.TestStartDate).Hours() / 24
				countryTestingMap[tc.Country.Name[0:3]] = countryTotalTime
				countryCount := countryTestingCountMap[tc.Country.Name[0:3]]
				countryCount[1]++
				countryTestingCountMap[tc.Country.Name[0:3]] = countryCount

				brandTotalTime := brandTestingMap[tc.Brand.Name]
				brandTotalTime[1] += tc.TestEndDate.Sub(*tc.TestStartDate).Hours() / 24
				brandTestingMap[tc.Brand.Name] = brandTotalTime
				brandCount := brandTestingCountMap[tc.Brand.Name]
				brandCount[1]++
				brandTestingCountMap[tc.Brand.Name] = brandCount
			}
			if tc.UnderStartDate != nil && tc.UnderEndDate != nil {
				countryTotalTime := countryUnderMap[tc.Country.Name[0:3]]
				countryTotalTime[1] += tc.UnderEndDate.Sub(*tc.UnderStartDate).Hours() / 24
				countryUnderMap[tc.Country.Name[0:3]] = countryTotalTime
				countryCount := countryUnderCountMap[tc.Country.Name[0:3]]
				countryCount[1]++
				countryUnderCountMap[tc.Country.Name[0:3]] = countryCount

				brandTotalTime := brandUnderMap[tc.Brand.Name]
				brandTotalTime[1] += tc.UnderEndDate.Sub(*tc.UnderStartDate).Hours() / 24
				brandUnderMap[tc.Brand.Name] = brandTotalTime
				brandCount := brandUnderCountMap[tc.Brand.Name]
				brandCount[1]++
				brandUnderCountMap[tc.Brand.Name] = brandCount
			}
		} else {
			if tc.TestStartDate != nil && tc.TestEndDate != nil {
				countryTotalTime := countryTestingMap[tc.Country.Name[0:3]]
				countryTotalTime[2] += tc.TestEndDate.Sub(*tc.TestStartDate).Hours() / 24
				countryTestingMap[tc.Country.Name[0:3]] = countryTotalTime
				countryCount := countryTestingCountMap[tc.Country.Name[0:3]]
				countryCount[2]++
				countryTestingCountMap[tc.Country.Name[0:3]] = countryCount

				brandTotalTime := brandTestingMap[tc.Brand.Name]
				brandTotalTime[2] += tc.TestEndDate.Sub(*tc.TestStartDate).Hours() / 24
				brandTestingMap[tc.Brand.Name] = brandTotalTime
				brandCount := brandTestingCountMap[tc.Brand.Name]
				brandCount[2]++
				brandTestingCountMap[tc.Brand.Name] = brandCount
			}
			if tc.UnderStartDate != nil && tc.UnderEndDate != nil {
				countryTotalTime := countryUnderMap[tc.Country.Name[0:3]]
				countryTotalTime[2] += tc.UnderEndDate.Sub(*tc.UnderStartDate).Hours() / 24
				countryUnderMap[tc.Country.Name[0:3]] = countryTotalTime
				countryCount := countryUnderCountMap[tc.Country.Name[0:3]]
				countryCount[2]++
				countryUnderCountMap[tc.Country.Name[0:3]] = countryCount

				brandTotalTime := brandUnderMap[tc.Brand.Name]
				brandTotalTime[2] += tc.UnderEndDate.Sub(*tc.UnderStartDate).Hours() / 24
				brandUnderMap[tc.Brand.Name] = brandTotalTime
				brandCount := brandUnderCountMap[tc.Brand.Name]
				brandCount[2]++
				brandUnderCountMap[tc.Brand.Name] = brandCount
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
	countryUnderFinalMap := make(map[string][3]float64)
	for key, value := range countryUnderMap {
		averageTime := countryUnderFinalMap[key]
		if value[0] > 0 {
			averageTime[0] = value[0] / float64(countryUnderCountMap[key][0])
		}
		if value[1] > 0 {
			averageTime[1] = value[1] / float64(countryUnderCountMap[key][1])
		}
		if value[2] > 0 {
			averageTime[2] = value[2] / float64(countryUnderCountMap[key][2])
		}
		countryUnderFinalMap[key] = averageTime
	}

	brandTestingFinalMap := make(map[string][3]float64)
	for key, value := range brandTestingMap {
		averageTime := brandTestingFinalMap[key]
		if value[0] > 0 {
			averageTime[0] = value[0] / float64(brandTestingCountMap[key][0])
		}
		if value[1] > 0 {
			averageTime[1] = value[1] / float64(brandTestingCountMap[key][1])
		}
		if value[2] > 0 {
			averageTime[2] = value[2] / float64(brandTestingCountMap[key][2])
		}

		brandTestingFinalMap[key] = averageTime
	}
	brandUnderFinalMap := make(map[string][3]float64)
	for key, value := range brandUnderMap {
		averageTime := brandUnderFinalMap[key]
		if value[0] > 0 {
			averageTime[0] = value[0] / float64(brandUnderCountMap[key][0])
		}
		if value[1] > 0 {
			averageTime[1] = value[1] / float64(brandUnderCountMap[key][1])
		}
		if value[2] > 0 {
			averageTime[2] = value[2] / float64(brandUnderCountMap[key][2])
		}
		brandUnderFinalMap[key] = averageTime
	}
	timeChart := new(responses.TimeChart)
	timeChart.CountryTestingChart = countryTestingFinalMap
	timeChart.CountryUnderChart = countryUnderFinalMap
	timeChart.BrandTestingChart = brandTestingFinalMap
	timeChart.BrandUnderChart = brandUnderFinalMap

	return timeChart, nil
}
