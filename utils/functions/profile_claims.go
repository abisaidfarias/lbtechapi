package functions

import "github.com/abisaidfarias/lbtechapi/models"

// HasProfileClaim reports whether a claim is present and allowed.
// Missing claims are treated as denied.
func HasProfileClaim(claims []models.Claim, name string) bool {
	for _, claim := range claims {
		if claim.Name == name {
			return claim.Allow
		}
	}
	return false
}
