package mapping

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

func CompanyRequestToCompany(company *request.Company) (*models.Company, error) {

	return &models.Company{
		Name:    company.Name,
		Email:   company.Email,
		Address: company.Address,
	}, nil
}
