package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Country model
type Country struct {
	ID         primitive.ObjectID `bson:"_id"`
	Name       string             `bson:"name"`
	BandGsm    []string           `bson:"band_gsm" json:"band_gsm"`
	BandWcdma  []string           `bson:"band_wcdma" json:"band_wcdma"`
	BandLte    []string           `bson:"band_lte" json:"band_lte"`
	Band5gNsa  []string           `bson:"band_5g_nsa" json:"band_5g_nsa"`
	Band5gSa   []string           `bson:"band_5g_sa" json:"band_5g_sa"`
	CarrierAgg []string           `bson:"carrier_agg" json:"carrier_agg"`
	// StoredActive is the raw field (nil on old countries); use the
	// "active" value of CountryListItem.
	StoredActive *bool `bson:"active" json:"-"`
}

// CountryListItem is one row of GET /country: the country (same fields as
// before) plus the Countries screen data.
type CountryListItem struct {
	Country
	// Active is false only when explicitly deactivated.
	Active bool `json:"active"`
	// UsedIn lists the modules with records of this country:
	// "homologation", "tracking", "shipment_control".
	UsedIn []string `json:"used_in"`
	// UsersCount counts users whose countries list includes it.
	UsersCount int `json:"users_count"`
	// DefaultFor lists the modules fixed to this country.
	DefaultFor []string `json:"default_for"`
}
