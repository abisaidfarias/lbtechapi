package enums

const (
	CanReadMultibanda   = "canReadMultibanda"
	CanWriteMultibanda  = "canWriteMultibanda"
	CanDeleteMultibanda = "canDeleteMultibanda"

	CanReadShipmentControl   = "canReadShipmentControl"
	CanWriteShipmentControl  = "canWriteShipmentControl"
	CanDeleteShipmentControl = "canDeleteShipmentControl"
)

// MultibandaProfileClaims lists the profile permissions for the Multibanda module.
var MultibandaProfileClaims = []string{
	CanReadMultibanda,
	CanWriteMultibanda,
	CanDeleteMultibanda,
}

// ShipmentControlProfileClaims lists the profile permissions for the Shipment Control module.
var ShipmentControlProfileClaims = []string{
	CanReadShipmentControl,
	CanWriteShipmentControl,
	CanDeleteShipmentControl,
}
