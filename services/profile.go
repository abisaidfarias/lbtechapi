package services

import (
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"gopkg.in/mgo.v2/bson"
)

// IProfileService is the test case service interface
type IProfileService interface {
	Create(*request.Profile) error
	Get(string) ([]*bson.M, error)
	GetById(string) (*responses.Profile, error)
	Update(string, *request.Profile) error
	Delete(string) (bool, error)
}

type profileService struct {
	profileRepository repositories.IProfileRepository
	userRepository    repositories.IUserRepository
}

// NewProfileService is a constructor
func NewProfileService(profileRepository repositories.IProfileRepository,
	userRepository repositories.IUserRepository) IProfileService {
	return &profileService{
		profileRepository: profileRepository,
		userRepository:    userRepository,
	}
}

// Create creates a new test case
func (s *profileService) Create(profileRequest *request.Profile) error {

	user, err := s.userRepository.GetByID(profileRequest.UserID)

	if err != nil {
		return err
	}
	profile, err := mapping.ProfileRequestToProfile(profileRequest)
	profile.Company = user.Company
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
func (s *profileService) Get(userID string) ([]*bson.M, error) {
	user, err := s.userRepository.GetByID(userID)

	if err != nil {
		return nil, err
	}
	if user.IsInternal {
		profiles, err := s.profileRepository.Get()
		if err != nil {
			return nil, err
		}
		return profiles, nil
	}

	profiles, err := s.profileRepository.GetByCompany(user.Company)

	if err != nil {
		return nil, err
	}

	return profiles, nil
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
func (s *profileService) Delete(id string) (bool, error) {
	canDelete, err := s.profileRepository.Delete(id)

	if err != nil {
		return canDelete, err
	}
	return canDelete, nil
}
