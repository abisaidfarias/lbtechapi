package responses

type ShipmentControlAvailableResponse struct {
	Company Company                          `json:"company"`
	Devices []ShipmentControlAvailableDevice `json:"devices"`
}

type ShipmentControlAvailableDevice struct {
	Device  Device                           `json:"device"`
	Brand   Brand                            `json:"brand"`
	Options []ShipmentControlAvailableOption `json:"options"`
}

type ShipmentControlAvailableOption struct {
	MultibandaID      string `json:"multibanda_id"`
	SubtelCertificateNumber string `json:"subtel_certificate_number"`
	SoftwareVersion   string `json:"software_version"`
	HardwareVersion   string `json:"hardware_version"`
	OsVersion         string `json:"os_version"`
	OsVersionView     string `json:"os_version_view"`
}
