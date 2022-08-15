package responses

type VolumeChart struct {
	CountryBrandChart *VolumeCountryBrandChart `json:"country_brand_chart" bson:"country_brand_chart"`
	TypeBrandChart    *VolumeTypeBrandChart    `json:"type_brand_chart" bson:"type_brand_chart"`
}
