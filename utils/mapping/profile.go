package mapping

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

func ProfileRequestToProfile(profile *request.Profile) (*models.Profile, error) {

	var claims []models.Claim
	for _, claimRequest := range profile.Claims {
		var claim models.Claim
		claim.Name = claimRequest.Name
		claim.Allow = claimRequest.Allow
		claims = append(claims, claim)
	}
	return &models.Profile{
		Claims:     claims,
		Name:       profile.Name,
		IsInternal: profile.IsInternal,
	}, nil
}
