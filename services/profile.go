package services

import (
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

// IProfileService is the test case service interface
type IProfileService interface {
	Create(*request.Profile) error
	Get() ([]*responses.Profile, error)
	GetById(string) (*responses.Profile, error)
	Update(string, *request.Profile) error
	Delete(string) (error, bool)
}

type profileService struct {
	profileRepository repositories.IProfileRepository
}

// NewProfileService is a constructor
func NewProfileService(profileRepository repositories.IProfileRepository) IProfileService {
	return &profileService{
		profileRepository: profileRepository,
	}
}

// Create creates a new test case
func (s *profileService) Create(profileRequest *request.Profile) error {

	profile, err := mapping.ProfileRequestToProfile(profileRequest)

	if err != nil {
		return err
	}

	err = s.profileRepository.Create(profile)

	if err != nil {
		return err
	}

	return nil
}

// Get gets a list of test cases
func (s *profileService) Get() ([]*responses.Profile, error) {
	result, err := s.profileRepository.Get()

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetById gets a case by id
func (s *profileService) GetById(id string) (*responses.Profile, error) {

	profile, err := s.profileRepository.GetById(id)

	if err != nil {
		return nil, err
	}
	return profile, nil
}

// Update updates a test case
func (s *profileService) Update(id string, profileRequest *request.Profile) error {

	profile, err := mapping.ProfileRequestToProfile(profileRequest)

	if err != nil {
		return err
	}
	err = s.profileRepository.Update(id, profile)
	if err != nil {
		return err
	}
	return nil
}
func (s *profileService) Delete(id string) (error, bool) {
	err, canDelete := s.profileRepository.Delete(id)

	if err != nil {
		return err, canDelete
	}
	return nil, canDelete
}
