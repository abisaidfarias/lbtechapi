package mapping

import (
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

func VolumeCountriesToVolumeCountriesCharts(chartVolumeCountries []*responses.ChartVolumeCountry) *responses.VolumeCountryBrandChart {

	brandMap := make(map[string]responses.StackedSerieChart)
	countryMap := make(map[string]bool)
	var countries []string
	turns := -1
	lastTime := ""
	var brands []string

	for _, tc := range chartVolumeCountries {

		if !countryMap[tc.Keys.Country[0:3]] {
			countryMap[tc.Keys.Country[0:3]] = true
			countries = append(countries, tc.Keys.Country[0:3])
		}
		if lastTime != tc.Keys.Country[0:3] {
			turns++
			lastTime = tc.Keys.Country[0:3]
		}
		stackedSerie := brandMap[tc.Keys.Brand]
		if len(stackedSerie.Group) == 0 {
			brands = append(brands, tc.Keys.Brand)
		}

		stackedSerie.Group = tc.Keys.Brand
		stackedSerie.XAxis = InsertSerieInPosition(turns, tc.Count, stackedSerie.XAxis)
		brandMap[tc.Keys.Brand] = stackedSerie

	}
	series := make([]responses.StackedSerieChart, 0, len(brandMap))
	for _, key := range brands {
		series = append(series, brandMap[key])
	}
	volumeChart := new(responses.VolumeCountryBrandChart)
	volumeChart.Countries = countries
	volumeChart.Brands = brands
	volumeChart.StackedSerieChart = series

	return volumeChart
}
func VolumeBrandToVolumeBrandCharts(chartVolumeTypes []*responses.ChartVolumeType) *responses.VolumeTypeBrandChart {

	typeMap := make(map[string]responses.StackedSerieChart)
	brandMap := make(map[string]bool)
	var brands []string
	turns := -1
	lastTime := ""
	var types []string

	for _, tc := range chartVolumeTypes {

		if !brandMap[tc.Keys.Brand] {
			brandMap[tc.Keys.Brand] = true
			brands = append(brands, tc.Keys.Brand)
		}
		if lastTime != tc.Keys.Brand {
			turns++
			lastTime = tc.Keys.Brand
		}
		typeName := enums.HomologationType_key[tc.Keys.Type]
		stackedSerie := typeMap[typeName]
		if len(stackedSerie.Group) == 0 {
			types = append(types, typeName)
		}

		stackedSerie.Group = typeName
		stackedSerie.XAxis = InsertSerieInPosition(turns, tc.Count, stackedSerie.XAxis)
		typeMap[typeName] = stackedSerie

	}
	series := make([]responses.StackedSerieChart, 0, len(typeMap))
	for _, key := range types {
		series = append(series, typeMap[key])
	}
	volumeChart := new(responses.VolumeTypeBrandChart)
	volumeChart.Types = types
	volumeChart.Brands = brands
	volumeChart.StackedSerieChart = series

	return volumeChart
}

// func InsertSerieInPosition(turns int, data int, series []int) []int {
// 	if len(series) == turns {
// 		series = append(series, data)
// 	} else {
// 		localTurns := turns - len(series)
// 		for i := 0; i < localTurns; i++ {
// 			series = append(series, 0)
// 		}
// 		series = append(series, data)
// 	}
// 	return series
// }
// func InsertCertificationChart(certificationSerie map[string][3]int, key string,
// 	approvalType int, count int) map[string][3]int {

// 	countCertificationSerie := certificationSerie[string(key)]
// 	if approvalType == enums.HomologationType_value["INITIAL"] {
// 		countCertificationSerie[0] += count

// 	} else if approvalType == enums.HomologationType_value["MAINTENANCE"] {
// 		countCertificationSerie[1] += count
// 	} else {
// 		countCertificationSerie[2] += count
// 	}
// 	certificationSerie[string(key)] = countCertificationSerie
// 	return certificationSerie
// }
