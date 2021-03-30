package mapping

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

func ProfileRequestToProfile(profile *request.Profile) (*models.Profile, error) {

	var claims []models.Claim
	for _, claimRequest := range profile.Claims {
		var claim models.Claim
		claim.Name = claimRequest.Name
		claim.Allow = claimRequest.Allow
		claims = append(claims, claim)
	}
	// isInternal, _ := strconv.ParseBool()
	return &models.Profile{
		Claims:     claims,
		Name:       profile.Name,
		IsInternal: profile.IsInternal,
	}, nil
}

func ProfileToProfileResponse(profile *models.Profile) (*responses.Profile, error) {

	var claims []responses.Claim
	for _, claimRequest := range profile.Claims {
		var claim responses.Claim
		claim.Name = claimRequest.Name
		claim.Allow = claimRequest.Allow
		claims = append(claims, claim)
	}
	return &responses.Profile{
		ID:         profile.ID,
		Claims:     claims,
		Name:       profile.Name,
		IsInternal: profile.IsInternal,
	}, nil
}
