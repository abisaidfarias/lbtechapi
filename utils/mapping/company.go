package mapping

import (
	"strings"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

// CompanyRequestToCompany maps a validated request; rut must already be the
// normalized value returned by utils.ValidateCompanyRequest.
func CompanyRequestToCompany(company *request.Company) (*models.Company, error) {
	return &models.Company{
		Name:                   strings.TrimSpace(company.Name),
		Address:                strings.TrimSpace(company.Address),
		ClientID:               strings.TrimSpace(company.ClientID),
		Rut:                    strings.TrimSpace(company.Rut),
		RazonSocial:            strings.TrimSpace(company.RazonSocial),
		AllowSmrWithoutTesting: company.AllowSmrWithoutTesting,
	}, nil
}
