package mapping

import (
	"fmt"
	"strconv"
	"time"

	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

func TypeCountriesToTypeCountriesCharts(chartTypeCountries []*responses.ChartTypeCountry) *responses.DashboardChart {

	countriesMap := make(map[string]responses.StackedSerieChart)
	timeMap := make(map[string]bool)
	var times []string
	turns := -1
	lastTime := ""
	pieSerie := make(map[string]int)
	certificationSerie := make(map[string][3]int)
	var countries []string
	for _, tc := range chartTypeCountries {

		certificationSerie = InsertCertificationChart(certificationSerie, tc.Keys.Country[0:3], tc.Keys.Type, tc.Count)
		countPieSerie := pieSerie[enums.HomologationType_key[tc.Keys.Type]]
		countPieSerie++
		pieSerie[enums.HomologationType_key[tc.Keys.Type]] = countPieSerie
		yearStr := strconv.Itoa(tc.Keys.Year)
		timeKey := fmt.Sprintf("%s %s", time.Month(tc.Keys.Month), yearStr[len(yearStr)-2:])
		time := timeMap[timeKey]
		if !time {
			timeMap[timeKey] = true
			times = append(times, timeKey)
		}
		if lastTime != timeKey {
			turns++
			lastTime = timeKey
		}
		typeName := enums.HomologationType_key[tc.Keys.Type]
		countryKey := fmt.Sprintf("%s %s", tc.Keys.Country, typeName)
		stackedSerie := countriesMap[countryKey]
		if len(stackedSerie.Group) == 0 {
			countries = append(countries, countryKey)
		}

		stackedSerie.Group = countryKey
		stackedSerie.XAxis = InsertSerieInPosition(turns, tc.Count, stackedSerie.XAxis)
		stackedSerie.Descripcion = string(typeName[0])
		countriesMap[countryKey] = stackedSerie

	}
	series := make([]responses.StackedSerieChart, 0, len(countriesMap))
	for _, key := range countries {
		series = append(series, countriesMap[key])
	}
	dashboardChart := new(responses.DashboardChart)
	dashboardChart.Countries = countries
	dashboardChart.Times = times
	dashboardChart.StackedSerieChart = series
	dashboardChart.PieSerieChart = pieSerie
	dashboardChart.CertificationSerieChart = certificationSerie

	return dashboardChart
}
func InsertSerieInPosition(turns int, data int, series []int) []int {
	if len(series) == turns {
		series = append(series, data)
	} else {
		localTurns := turns - len(series)
		for i := 0; i < localTurns; i++ {
			series = append(series, 0)
		}
		series = append(series, data)
	}
	return series
}
func InsertCertificationChart(certificationSerie map[string][3]int, key string,
	approvalType int, count int) map[string][3]int {

	countCertificationSerie := certificationSerie[string(key)]
	if approvalType == enums.HomologationType_value["INITIAL"] {
		countCertificationSerie[0] += count

	} else if approvalType == enums.HomologationType_value["MAINTENANCE"] {
		countCertificationSerie[1] += count
	} else {
		countCertificationSerie[2] += count
	}
	certificationSerie[string(key)] = countCertificationSerie
	return certificationSerie
}
