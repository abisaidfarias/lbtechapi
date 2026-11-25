package responses

// DeviceCatalog is one category of GET /device-catalogs.
type DeviceCatalog struct {
	Category string               `json:"category"`
	Label    string               `json:"label"`
	Group    string               `json:"group"`
	Values   []DeviceCatalogValue `json:"values"`
}

// DeviceCatalogValue is one value of a catalog, in order.
type DeviceCatalogValue struct {
	ID     string `json:"id"`
	Value  string `json:"value"`
	Unit   string `json:"unit"`
	Active bool   `json:"active"`
	Order  int    `json:"order"`
	// UsedCount counts devices whose mapped field equals this value exactly.
	UsedCount int `json:"used_count"`
}
