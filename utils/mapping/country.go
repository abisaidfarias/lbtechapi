package mapping

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

func CountryRequestToCountry(country *request.Country) (*models.Country, error) {

	return &models.Country{
		Name: country.Name,
	}, nil
}
