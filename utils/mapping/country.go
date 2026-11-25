package mapping

import (
	"strings"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"go.mongodb.org/mongo-driver/bson"
)

// CountryRequestToCountry maps a create request. Absent band lists start
// empty ([] rather than null) and a missing active means active.
func CountryRequestToCountry(country *request.Country) (*models.Country, error) {
	active := true
	if country.Active != nil {
		active = *country.Active
	}
	return &models.Country{
		Name:       strings.TrimSpace(country.Name),
		BandGsm:    emptyIfNil(country.BandGsm),
		BandWcdma:  emptyIfNil(country.BandWcdma),
		BandLte:    emptyIfNil(country.BandLte),
		Band5gNsa:  emptyIfNil(country.Band5gNsa),
		Band5gSa:   emptyIfNil(country.Band5gSa),
		CarrierAgg: emptyIfNil(country.CarrierAgg),
		Active:     &active,
	}, nil
}

// CountryRequestToUpdate builds a partial $set: name always, active and each
// band list only when present in the body, so hidden bands are never wiped.
func CountryRequestToUpdate(country *request.Country) bson.M {
	set := bson.M{"name": strings.TrimSpace(country.Name)}
	if country.Active != nil {
		set["active"] = *country.Active
	}
	bands := map[string][]string{
		"band_gsm":    country.BandGsm,
		"band_wcdma":  country.BandWcdma,
		"band_lte":    country.BandLte,
		"band_5g_nsa": country.Band5gNsa,
		"band_5g_sa":  country.Band5gSa,
		"carrier_agg": country.CarrierAgg,
	}
	for field, values := range bands {
		if values != nil {
			set[field] = values
		}
	}
	return set
}

func emptyIfNil(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
