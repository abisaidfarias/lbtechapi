package services

import (
	"fmt"
	"strings"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ICountryService is the country service
type ICountryService interface {
	Create(*request.Country) error
	Get() ([]*responses.CountryListItem, error)
	Delete(string) error
	Update(string, *request.Country) error
}

type countryService struct {
	countryRepository      repositories.ICountryRepository
	homologationRepository repositories.IHomologationRepository
}

// NewCountryService is a constructor
func NewCountryService(countryRepository repositories.ICountryRepository,
	homologationRepository repositories.IHomologationRepository) ICountryService {
	return &countryService{
		countryRepository:      countryRepository,
		homologationRepository: homologationRepository,
	}
}

// Modules fixed to Chile. Shipment Control resolves the country by this name
// for external users; Multi-band records have no country at all.
var countryDefaultModules = []string{"multibanda", "shipment_control"}

var countryModuleLabels = map[string]string{
	repositories.CountryModuleHomologation:    "Homologation",
	repositories.CountryModuleTracking:        "Tracking",
	repositories.CountryModuleShipmentControl: "Shipment Control",
}

// Create validates and stores a new country with empty bands (unless sent).
func (s *countryService) Create(countryRequest *request.Country) error {
	if err := s.validateName(countryRequest, primitive.NilObjectID); err != nil {
		return err
	}
	country, err := mapping.CountryRequestToCountry(countryRequest)
	if err != nil {
		return err
	}
	return s.countryRepository.Create(country)
}

// Get lists the countries with the Countries screen data.
func (s *countryService) Get() ([]*responses.CountryListItem, error) {
	countries, err := s.countryRepository.Get()
	if err != nil {
		return nil, err
	}
	usage, err := s.countryRepository.CountryUsage()
	if err != nil {
		return nil, err
	}
	users, err := s.countryRepository.UsersCountByCountry()
	if err != nil {
		return nil, err
	}

	items := make([]*responses.CountryListItem, 0, len(countries))
	for _, c := range countries {
		usedIn := usage[c.ID]
		if usedIn == nil {
			usedIn = []string{}
		}
		defaultFor := []string{}
		if strings.EqualFold(strings.TrimSpace(c.Name), enums.ShipmentControlExternalCountryName) {
			defaultFor = append(defaultFor, countryDefaultModules...)
		}
		items = append(items, &responses.CountryListItem{
			Country:    *c,
			Active:     models.IsCountryActive(c.StoredActive),
			UsedIn:     usedIn,
			UsersCount: users[c.ID],
			DefaultFor: defaultFor,
		})
	}
	return items, nil
}

// Update changes name/active and only the band lists present in the body.
func (s *countryService) Update(id string, countryRequest *request.Country) error {
	countryID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return utils.NewValidationError("invalid country id")
	}
	if err := s.validateName(countryRequest, countryID); err != nil {
		return err
	}
	existing, err := s.countryRepository.GetById(id)
	if err != nil || existing == nil {
		return utils.NewValidationError("country not found")
	}
	if err := guardFixedCountry(existing.Name, countryRequest); err != nil {
		return err
	}
	return s.countryRepository.UpdateFields(countryID, mapping.CountryRequestToUpdate(countryRequest))
}

const (
	CountryFixedRenameMessage     = "Chile is fixed for Multi-band and Shipment Control and cannot be renamed"
	CountryFixedDeactivateMessage = "Chile is fixed for Multi-band and Shipment Control and cannot be deactivated"
)

// guardFixedCountry protects the country Shipment Control resolves by name
// (enums.ShipmentControlExternalCountryName): renaming or deactivating it
// would break Shipment Control for external users. A change of letter case
// only is allowed, since the lookup is case-insensitive.
func guardFixedCountry(storedName string, req *request.Country) error {
	fixed := enums.ShipmentControlExternalCountryName
	if !strings.EqualFold(strings.TrimSpace(storedName), fixed) {
		return nil
	}
	if !strings.EqualFold(strings.TrimSpace(req.Name), fixed) {
		return utils.NewValidationError(CountryFixedRenameMessage)
	}
	if req.Active != nil && !*req.Active {
		return utils.NewValidationError(CountryFixedDeactivateMessage)
	}
	return nil
}

func (s *countryService) validateName(req *request.Country, exclude primitive.ObjectID) error {
	if req == nil || strings.TrimSpace(req.Name) == "" {
		return utils.NewValidationError("Name is required")
	}
	name := strings.TrimSpace(req.Name)
	other, err := s.countryRepository.FindByNameExcluding(name, exclude)
	if err != nil {
		return err
	}
	if other != nil {
		return utils.NewValidationError(fmt.Sprintf("Country %q already exists", strings.TrimSpace(other.Name)))
	}
	return nil
}

// Delete refuses countries referenced by any module; they can be
// deactivated instead.
func (s *countryService) Delete(id string) error {
	countryID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return utils.NewValidationError("invalid country id")
	}
	usage, err := s.countryRepository.CountryUsage()
	if err != nil {
		return err
	}
	if modules := usage[countryID]; len(modules) > 0 {
		return utils.NewValidationError(CountryInUseMessage(modules))
	}
	return s.countryRepository.Delete(countryID)
}

// CountryInUseMessage is the delete error, e.g. "This country is used in
// Homologation/Tracking and cannot be deleted. Deactivate it instead."
func CountryInUseMessage(modules []string) string {
	labels := make([]string, 0, len(modules))
	for _, m := range modules {
		labels = append(labels, countryModuleLabels[m])
	}
	return fmt.Sprintf("This country is used in %s and cannot be deleted. Deactivate it instead.",
		strings.Join(labels, "/"))
}
