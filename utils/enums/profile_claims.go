package enums

const (
	CanReadMultibanda   = "canReadMultibanda"
	CanWriteMultibanda  = "canWriteMultibanda"
	CanDeleteMultibanda = "canDeleteMultibanda"
)

// MultibandaProfileClaims lists the profile permissions for the Multibanda module.
var MultibandaProfileClaims = []string{
	CanReadMultibanda,
	CanWriteMultibanda,
	CanDeleteMultibanda,
}
