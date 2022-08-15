package mapping

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

func PersonRequestToPerson(person *request.Person) (*models.Person, error) {

	return &models.Person{
		Name: person.Name,
	}, nil
}
