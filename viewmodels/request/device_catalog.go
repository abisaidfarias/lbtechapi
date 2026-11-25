package request

// DeviceCatalogValueCreate is POST /device-catalogs/:category/values.
type DeviceCatalogValueCreate struct {
	Value string `json:"value"`
	Unit  string `json:"unit"`
}

// DeviceCatalogValueUpdate is PUT /device-catalogs/values/:id; absent fields
// are left unchanged.
type DeviceCatalogValueUpdate struct {
	Value  *string `json:"value"`
	Unit   *string `json:"unit"`
	Active *bool   `json:"active"`
}

// DeviceCatalogOrder is PUT /device-catalogs/:category/order.
type DeviceCatalogOrder struct {
	IDs []string `json:"ids"`
}
