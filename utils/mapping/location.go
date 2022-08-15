package mapping

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

func LocationRequestToLocation(location *request.Location) (*models.Location, error) {

	return &models.Location{
		Name: location.Name,
	}, nil
}
