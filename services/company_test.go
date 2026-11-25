package services

import (
	"errors"
	"testing"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// memCompanyRepo is an in-memory ICompanyRepository.
type memCompanyRepo struct {
	companies    map[primitive.ObjectID]*responses.Company
	certificates map[primitive.ObjectID]int
	users        map[primitive.ObjectID]int
	updated      *models.Company
	created      *models.Company
}

func newMemCompanyRepo() *memCompanyRepo {
	return &memCompanyRepo{
		companies:    map[primitive.ObjectID]*responses.Company{},
		certificates: map[primitive.ObjectID]int{},
		users:        map[primitive.ObjectID]int{},
	}
}

func (m *memCompanyRepo) add(name, clientID string) *responses.Company {
	c := &responses.Company{ID: primitive.NewObjectID(), Name: name, ClientID: clientID,
		Rut: "12.345.678-5", RazonSocial: name + " SpA", Address: "Calle 1"}
	m.companies[c.ID] = c
	return c
}

func (m *memCompanyRepo) Create(c *models.Company) error { m.created = c; return nil }
func (m *memCompanyRepo) Get() ([]*responses.Company, error) {
	out := []*responses.Company{}
	for _, c := range m.companies {
		out = append(out, c)
	}
	return out, nil
}
func (m *memCompanyRepo) Update(_ primitive.ObjectID, c *models.Company) error {
	m.updated = c
	return nil
}
func (m *memCompanyRepo) Delete(primitive.ObjectID) error { return nil }
func (m *memCompanyRepo) GetById(id string) (*responses.Company, error) {
	oid, _ := primitive.ObjectIDFromHex(id)
	if c, ok := m.companies[oid]; ok {
		return c, nil
	}
	return nil, errors.New("mongo: no documents in result")
}
func (m *memCompanyRepo) Count() (int64, error) { return int64(len(m.companies)), nil }
func (m *memCompanyRepo) GetByClientID(clientID string, exclude primitive.ObjectID) (*responses.Company, error) {
	for _, c := range m.companies {
		if c.ClientID == clientID && c.ID != exclude {
			return c, nil
		}
	}
	return nil, nil
}
func (m *memCompanyRepo) UsersCountByCompany() (map[primitive.ObjectID]int, error) {
	return m.users, nil
}
func (m *memCompanyRepo) IssuedCertificatesByCompany() (map[primitive.ObjectID]int, error) {
	return m.certificates, nil
}
func (m *memCompanyRepo) CountIssuedCertificates(id primitive.ObjectID) (int64, error) {
	return int64(m.certificates[id]), nil
}
func (m *memCompanyRepo) GetCompanyUsers(primitive.ObjectID) ([]responses.CompanyUser, error) {
	return []responses.CompanyUser{}, nil
}

func companyReq(name, clientID string) *request.Company {
	return &request.Company{Name: name, ClientID: clientID, RazonSocial: name + " SpA",
		Address: "Calle 1", Rut: "12345678-5", AllowSmrWithoutTesting: true}
}

func TestCompanyCreateRejectsDuplicateClientID(t *testing.T) {
	repo := newMemCompanyRepo()
	repo.add("Xiaomi Multibanda", "110")
	s := &companyService{companyRepository: repo}

	err := s.Create(companyReq("Other", " 110 "))
	want := "Client ID 110 is already used by \"Xiaomi Multibanda\""
	if !utils.IsValidationError(err) || err.Error() != want {
		t.Fatalf("want %q, got %v", want, err)
	}
	if err := s.Create(companyReq("Other", "111")); err != nil {
		t.Fatal(err)
	}
	if repo.created.Rut != "12.345.678-5" || repo.created.ClientID != "111" || !repo.created.AllowSmrWithoutTesting {
		t.Errorf("created = %+v", repo.created)
	}
}

func TestCompanyUpdateKeepsOwnClientIDAndChecksOthers(t *testing.T) {
	repo := newMemCompanyRepo()
	c := repo.add("Xiaomi", "110")
	repo.add("Samsung", "120")
	s := &companyService{companyRepository: repo}

	if err := s.Update(c.ID.Hex(), companyReq("Xiaomi", "110")); err != nil {
		t.Fatalf("saving with its own client id: %v", err)
	}
	if err := s.Update(c.ID.Hex(), companyReq("Xiaomi", "120")); !utils.IsValidationError(err) {
		t.Fatalf("taking Samsung's client id: %v", err)
	}
}

func TestCompanyClientIDLockedAfterCertificates(t *testing.T) {
	repo := newMemCompanyRepo()
	c := repo.add("Xiaomi", "110")
	repo.certificates[c.ID] = 2
	s := &companyService{companyRepository: repo}

	err := s.Update(c.ID.Hex(), companyReq("Xiaomi", "999"))
	if !utils.IsValidationError(err) || err.Error() != utils.CompanyClientIDLockedMessage {
		t.Fatalf("want lock error, got %v", err)
	}
	// Other fields stay editable while the client id is unchanged.
	req := companyReq("Xiaomi Chile", "110")
	req.Address = "Nueva direccion 99"
	if err := s.Update(c.ID.Hex(), req); err != nil {
		t.Fatalf("editing other fields: %v", err)
	}
	if repo.updated.Address != "Nueva direccion 99" {
		t.Errorf("updated = %+v", repo.updated)
	}
}

func TestCompanyUpdateClientIDFreeWithoutCertificates(t *testing.T) {
	repo := newMemCompanyRepo()
	c := repo.add("Xiaomi", "110")
	s := &companyService{companyRepository: repo}
	if err := s.Update(c.ID.Hex(), companyReq("Xiaomi", "999")); err != nil {
		t.Fatal(err)
	}
}

func TestCompanyGetAddsAdminFields(t *testing.T) {
	repo := newMemCompanyRepo()
	complete := repo.add("Complete", "110")
	legacy := &responses.Company{ID: primitive.NewObjectID(), Name: "Legacy"} // incomplete data stays readable
	repo.companies[legacy.ID] = legacy
	repo.certificates[complete.ID] = 3
	repo.users[complete.ID] = 5
	s := &companyService{companyRepository: repo}

	items, err := s.Get()
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]*responses.CompanyListItem{}
	for _, it := range items {
		byName[it.Name] = it
	}
	c := byName["Complete"]
	if c.UsersCount != 5 || c.OabiCertificatesIssued != 3 || !c.ClientIDLocked || len(c.OabiMissing) != 0 {
		t.Errorf("complete = %+v", c)
	}
	l := byName["Legacy"]
	if l.ClientIDLocked || len(l.OabiMissing) != 4 {
		t.Errorf("legacy = %+v", l)
	}
}

func TestCompanyClientIDAvailable(t *testing.T) {
	repo := newMemCompanyRepo()
	c := repo.add("Xiaomi Multibanda", "110")
	s := &companyService{companyRepository: repo}

	res, err := s.ClientIDAvailable("110", "")
	if err != nil || res.Available || res.UsedBy != "Xiaomi Multibanda" {
		t.Fatalf("taken: %+v %v", res, err)
	}
	res, err = s.ClientIDAvailable("110", c.ID.Hex())
	if err != nil || !res.Available || res.UsedBy != "" {
		t.Fatalf("own id excluded: %+v %v", res, err)
	}
	if _, err := s.ClientIDAvailable(" ", ""); !utils.IsValidationError(err) {
		t.Fatalf("empty client id: %v", err)
	}
}
