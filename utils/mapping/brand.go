package mapping

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

func BrandRequestToBrand(brand *request.Brand) (*models.Brand, error) {

	return &models.Brand{
		Name: brand.Name,
	}, nil
}
