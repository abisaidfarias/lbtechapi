package mapping

import (
	"reflect"
	"testing"

	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

// The Countries screen sends only name and active: the update must not touch
// any band list.
func TestCountryRequestToUpdateKeepsAbsentBands(t *testing.T) {
	inactive := false
	set := CountryRequestToUpdate(&request.Country{Name: " Chile ", Active: &inactive})
	want := map[string]interface{}{"name": "Chile", "active": false}
	if !reflect.DeepEqual(map[string]interface{}(set), want) {
		t.Fatalf("set = %v, want %v", set, want)
	}
}

func TestCountryRequestToUpdateWithoutActiveKeepsIt(t *testing.T) {
	set := CountryRequestToUpdate(&request.Country{Name: "Peru"})
	if _, ok := set["active"]; ok {
		t.Fatalf("active must not be set when absent: %v", set)
	}
}

// Bands sent explicitly are still written, including [] to clear one.
func TestCountryRequestToUpdateWritesSentBands(t *testing.T) {
	set := CountryRequestToUpdate(&request.Country{
		Name: "Chile", BandLte: []string{"B28"}, Band5gSa: []string{},
	})
	if !reflect.DeepEqual(set["band_lte"], []string{"B28"}) {
		t.Errorf("band_lte = %v", set["band_lte"])
	}
	if v, ok := set["band_5g_sa"]; !ok || len(v.([]string)) != 0 {
		t.Errorf("band_5g_sa = %v (sent [] must clear)", v)
	}
	for _, f := range []string{"band_gsm", "band_wcdma", "band_5g_nsa", "carrier_agg"} {
		if _, ok := set[f]; ok {
			t.Errorf("%s set although absent", f)
		}
	}
}

func TestCountryRequestToCountryDefaults(t *testing.T) {
	c, _ := CountryRequestToCountry(&request.Country{Name: " Peru "})
	if c.Name != "Peru" || c.Active == nil || !*c.Active {
		t.Fatalf("country = %+v", c)
	}
	for name, bands := range map[string][]string{
		"band_gsm": c.BandGsm, "band_wcdma": c.BandWcdma, "band_lte": c.BandLte,
		"band_5g_nsa": c.Band5gNsa, "band_5g_sa": c.Band5gSa, "carrier_agg": c.CarrierAgg,
	} {
		if bands == nil || len(bands) != 0 {
			t.Errorf("%s = %v, want empty list", name, bands)
		}
	}
}
