package services

import (
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IPersonService is the person service
type IPersonService interface {
	Create(*request.Person) error
	Get() ([]*responses.Person, error)
	Delete(string) (bool, error)
	Update(string, *request.Person) error
}

type personService struct {
	personRepository         repositories.IPersonRepository
	deviceTrackingRepository repositories.IDeviceTrackingRepository
}

// NewPersonService is a constructor
func NewPersonService(personRepository repositories.IPersonRepository,
	deviceTrackingRepository repositories.IDeviceTrackingRepository) IPersonService {
	return &personService{
		personRepository:         personRepository,
		deviceTrackingRepository: deviceTrackingRepository,
	}
}

// Create creates a new cateogry
func (s *personService) Create(personRequest *request.Person) error {

	person, err := mapping.PersonRequestToPerson(personRequest)
	if err != nil {
		return err
	}
	err = s.personRepository.Create(person)

	if err != nil {
		return err
	}

	return nil
}

// Get gets a list of all categories
func (s *personService) Get() ([]*responses.Person, error) {
	result, err := s.personRepository.Get()

	if err != nil {
		return nil, err
	}

	return result, nil
}
func (s *personService) Update(id string, personRequest *request.Person) error {
	personId, _ := primitive.ObjectIDFromHex(id)
	person, err := mapping.PersonRequestToPerson(personRequest)

	if err != nil {
		return err
	}
	err = s.personRepository.Update(personId, person)
	if err != nil {
		return err
	}
	return nil
}
func (s *personService) Delete(id string) (bool, error) {
	personId, _ := primitive.ObjectIDFromHex(id)

	deviceTacking, err := s.deviceTrackingRepository.GetByPerson(personId)
	if err != nil {
		return false, err
	}
	if deviceTacking != nil {
		return true, err
	}
	err = s.personRepository.Delete(personId)
	if err != nil {
		return false, err
	}
	return false, nil
}
