package services

import (
	"reflect"
	"strings"
	"testing"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// memCountryRepo is an in-memory ICountryRepository.
type memCountryRepo struct {
	countries map[primitive.ObjectID]*responses.Country
	usage     map[primitive.ObjectID][]string
	users     map[primitive.ObjectID]int
	created   *models.Country
	updates   []primitive.M
	deleted   []primitive.ObjectID
}

func newMemCountryRepo() *memCountryRepo {
	return &memCountryRepo{
		countries: map[primitive.ObjectID]*responses.Country{},
		usage:     map[primitive.ObjectID][]string{},
		users:     map[primitive.ObjectID]int{},
	}
}

func (m *memCountryRepo) add(name string) *responses.Country {
	c := &responses.Country{ID: primitive.NewObjectID(), Name: name, BandLte: []string{"B28"}}
	m.countries[c.ID] = c
	return c
}

func (m *memCountryRepo) Create(c *models.Country) error { m.created = c; return nil }
func (m *memCountryRepo) Get() ([]*responses.Country, error) {
	out := []*responses.Country{}
	for _, c := range m.countries {
		out = append(out, c)
	}
	return out, nil
}
func (m *memCountryRepo) Update(primitive.ObjectID, *models.Country) error { return nil }
func (m *memCountryRepo) Delete(id primitive.ObjectID) error {
	m.deleted = append(m.deleted, id)
	return nil
}
func (m *memCountryRepo) GetById(id string) (*responses.Country, error) {
	oid, _ := primitive.ObjectIDFromHex(id)
	return m.countries[oid], nil
}
func (m *memCountryRepo) GetByName(string) (*responses.Country, error)            { return nil, nil }
func (m *memCountryRepo) GetCountriesById([]primitive.ObjectID) ([]string, error) { return nil, nil }
func (m *memCountryRepo) UpdateFields(_ primitive.ObjectID, set primitive.M) error {
	m.updates = append(m.updates, set)
	return nil
}
func (m *memCountryRepo) FindByNameExcluding(name string, exclude primitive.ObjectID) (*responses.Country, error) {
	for _, c := range m.countries {
		if strings.EqualFold(strings.TrimSpace(c.Name), strings.TrimSpace(name)) && c.ID != exclude {
			return c, nil
		}
	}
	return nil, nil
}
func (m *memCountryRepo) CountryUsage() (map[primitive.ObjectID][]string, error) { return m.usage, nil }
func (m *memCountryRepo) UsersCountByCountry() (map[primitive.ObjectID]int, error) {
	return m.users, nil
}

func TestCountryCreateRejectsDuplicateNameIgnoringCase(t *testing.T) {
	repo := newMemCountryRepo()
	repo.add("Chile")
	s := &countryService{countryRepository: repo}

	err := s.Create(&request.Country{Name: "  chile "})
	if !utils.IsValidationError(err) || err.Error() != `Country "Chile" already exists` {
		t.Fatalf("got %v", err)
	}
	if err := s.Create(&request.Country{Name: " "}); !utils.IsValidationError(err) || err.Error() != "Name is required" {
		t.Fatalf("empty name: %v", err)
	}
	if err := s.Create(&request.Country{Name: "Peru"}); err != nil {
		t.Fatal(err)
	}
	if repo.created.Name != "Peru" || !*repo.created.Active || repo.created.BandLte == nil {
		t.Errorf("created = %+v", repo.created)
	}
}

func TestCountryUpdateIsPartialAndAllowsOwnName(t *testing.T) {
	repo := newMemCountryRepo()
	bolivia := repo.add("Bolivia")
	repo.add("Peru")
	s := &countryService{countryRepository: repo}

	inactive := false
	if err := s.Update(bolivia.ID.Hex(), &request.Country{Name: "BOLIVIA", Active: &inactive}); err != nil {
		t.Fatalf("renaming to its own name: %v", err)
	}
	want := primitive.M{"name": "BOLIVIA", "active": false}
	if !reflect.DeepEqual(repo.updates[0], want) {
		t.Errorf("update = %v, want only name+active (bands untouched)", repo.updates[0])
	}
	if err := s.Update(bolivia.ID.Hex(), &request.Country{Name: "peru"}); !utils.IsValidationError(err) {
		t.Fatalf("taking Peru's name: %v", err)
	}
}

func TestCountryGetAddsAdminFields(t *testing.T) {
	repo := newMemCountryRepo()
	chile := repo.add("CHILE")
	peru := repo.add("Peru")
	off := false
	peru.StoredActive = &off
	repo.usage[chile.ID] = []string{repositories.CountryModuleHomologation, repositories.CountryModuleTracking}
	repo.users[chile.ID] = 4
	s := &countryService{countryRepository: repo}

	items, err := s.Get()
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]*responses.CountryListItem{}
	for _, it := range items {
		byName[it.Name] = it
	}
	c := byName["CHILE"]
	if !c.Active || c.UsersCount != 4 ||
		!reflect.DeepEqual(c.UsedIn, []string{"homologation", "tracking"}) ||
		!reflect.DeepEqual(c.DefaultFor, []string{"multibanda", "shipment_control"}) ||
		!reflect.DeepEqual(c.BandLte, []string{"B28"}) {
		t.Errorf("chile = %+v", c)
	}
	p := byName["Peru"]
	if p.Active || p.UsedIn == nil || len(p.UsedIn) != 0 || len(p.DefaultFor) != 0 {
		t.Errorf("peru = %+v", p)
	}
}

func TestCountryDeleteRefusedWhenUsed(t *testing.T) {
	repo := newMemCountryRepo()
	chile := repo.add("Chile")
	unused := repo.add("Peru")
	repo.usage[chile.ID] = []string{repositories.CountryModuleHomologation, repositories.CountryModuleTracking}
	s := &countryService{countryRepository: repo}

	err := s.Delete(chile.ID.Hex())
	want := "This country is used in Homologation/Tracking and cannot be deleted. Deactivate it instead."
	if !utils.IsValidationError(err) || err.Error() != want {
		t.Fatalf("got %v", err)
	}
	if err := s.Delete(unused.ID.Hex()); err != nil || len(repo.deleted) != 1 || repo.deleted[0] != unused.ID {
		t.Fatalf("deleting unused country: %v %v", err, repo.deleted)
	}
}

// Chile is resolved by name by Shipment Control: it cannot be renamed or
// deactivated, but other edits (and a letter-case change) still work.
func TestCountryUpdateGuardsChile(t *testing.T) {
	repo := newMemCountryRepo()
	chile := repo.add("Chile")
	s := &countryService{countryRepository: repo}
	off, on := false, true

	err := s.Update(chile.ID.Hex(), &request.Country{Name: "Chile CL"})
	if !utils.IsValidationError(err) || err.Error() != CountryFixedRenameMessage {
		t.Fatalf("rename: %v", err)
	}
	err = s.Update(chile.ID.Hex(), &request.Country{Name: "Chile", Active: &off})
	if !utils.IsValidationError(err) || err.Error() != CountryFixedDeactivateMessage {
		t.Fatalf("deactivate: %v", err)
	}
	if len(repo.updates) != 0 {
		t.Fatalf("blocked updates were written: %v", repo.updates)
	}
	for _, req := range []*request.Country{
		{Name: "Chile"},
		{Name: " chile ", Active: &on},
		{Name: "Chile", BandLte: []string{"B28", "B7"}},
	} {
		if err := s.Update(chile.ID.Hex(), req); err != nil {
			t.Errorf("allowed edit %+v: %v", req, err)
		}
	}
}
