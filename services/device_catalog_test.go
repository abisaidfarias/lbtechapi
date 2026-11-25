package services

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// memCatalogRepo is an in-memory IDeviceCatalogRepository.
type memCatalogRepo struct {
	values map[primitive.ObjectID]*models.DeviceCatalogValue
	used   map[string]map[string]int // device field → value → count
}

func newMemCatalogRepo() *memCatalogRepo {
	return &memCatalogRepo{values: map[primitive.ObjectID]*models.DeviceCatalogValue{}, used: map[string]map[string]int{}}
}

func (m *memCatalogRepo) sorted(filter func(*models.DeviceCatalogValue) bool) []models.DeviceCatalogValue {
	out := []models.DeviceCatalogValue{}
	for _, v := range m.values {
		if filter(v) {
			out = append(out, *v)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Category != out[j].Category {
			return out[i].Category < out[j].Category
		}
		return out[i].Order < out[j].Order
	})
	return out
}
func (m *memCatalogRepo) List() ([]models.DeviceCatalogValue, error) {
	return m.sorted(func(*models.DeviceCatalogValue) bool { return true }), nil
}
func (m *memCatalogRepo) ListByCategory(c string) ([]models.DeviceCatalogValue, error) {
	return m.sorted(func(v *models.DeviceCatalogValue) bool { return v.Category == c }), nil
}
func (m *memCatalogRepo) Get(id primitive.ObjectID) (*models.DeviceCatalogValue, error) {
	if v, ok := m.values[id]; ok {
		c := *v
		return &c, nil
	}
	return nil, nil
}
func (m *memCatalogRepo) FindByKey(c, key string, exclude primitive.ObjectID) (*models.DeviceCatalogValue, error) {
	for _, v := range m.values {
		if v.Category == c && v.ValueKey == key && v.ID != exclude {
			return v, nil
		}
	}
	return nil, nil
}
func (m *memCatalogRepo) Insert(v *models.DeviceCatalogValue) error {
	v.ID = primitive.NewObjectID()
	c := *v
	m.values[v.ID] = &c
	return nil
}
func (m *memCatalogRepo) UpdateFields(id primitive.ObjectID, set primitive.M) error {
	v := m.values[id]
	for k, val := range set {
		switch k {
		case "value":
			v.Value = val.(string)
		case "value_key":
			v.ValueKey = val.(string)
		case "unit":
			v.Unit = val.(string)
		case "active":
			v.Active = val.(bool)
		}
	}
	return nil
}
func (m *memCatalogRepo) SetOrder(_ string, ids []primitive.ObjectID) error {
	for i, id := range ids {
		m.values[id].Order = i
	}
	return nil
}
func (m *memCatalogRepo) Delete(id primitive.ObjectID) error { delete(m.values, id); return nil }
func (m *memCatalogRepo) UsedCounts(field string) (map[string]int, error) {
	if m.used[field] == nil {
		return map[string]int{}, nil
	}
	return m.used[field], nil
}

func newCatalogService() (*deviceCatalogService, *memCatalogRepo) {
	repo := newMemCatalogRepo()
	return &deviceCatalogService{repo: repo, now: func() time.Time { return time.Unix(0, 0) }}, repo
}

func mustCreate(t *testing.T, s *deviceCatalogService, cat, value string) *responses.DeviceCatalogValue {
	t.Helper()
	v, err := s.CreateValue(cat, &request.DeviceCatalogValueCreate{Value: value})
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestDeviceCatalogCreateAppendsAndRejectsDuplicates(t *testing.T) {
	s, _ := newCatalogService()
	a := mustCreate(t, s, "rom", "64GB")
	b, err := s.CreateValue("rom", &request.DeviceCatalogValueCreate{Value: " 128 ", Unit: " GB "})
	if err != nil {
		t.Fatal(err)
	}
	if a.Order != 0 || b.Order != 1 || !b.Active || b.Value != "128" || b.Unit != "GB" {
		t.Fatalf("a=%+v b=%+v", a, b)
	}
	_, err = s.CreateValue("rom", &request.DeviceCatalogValueCreate{Value: "128"})
	if !utils.IsValidationError(err) || err.Error() != `"128" already exists in ROM` {
		t.Fatalf("duplicate: %v", err)
	}
	if _, err := s.CreateValue("ram", &request.DeviceCatalogValueCreate{Value: "128"}); err != nil {
		t.Fatalf("same value in another category must be allowed: %v", err)
	}
	if _, err := s.CreateValue("bands", &request.DeviceCatalogValueCreate{Value: "B1"}); !utils.IsValidationError(err) {
		t.Fatalf("unknown category: %v", err)
	}
	if _, err := s.CreateValue("usb", &request.DeviceCatalogValueCreate{Value: "  "}); !utils.IsValidationError(err) {
		t.Fatalf("empty value: %v", err)
	}
}

func TestDeviceCatalogUpdateRenameDeactivate(t *testing.T) {
	s, repo := newCatalogService()
	a := mustCreate(t, s, "usb", "USB-C")
	mustCreate(t, s, "usb", "MicroUSB")
	repo.used["communication_usb"] = map[string]int{"USB Type-C": 4}

	name, off := "USB Type-C", false
	got, err := s.UpdateValue(a.ID, &request.DeviceCatalogValueUpdate{Value: &name, Active: &off})
	if err != nil {
		t.Fatal(err)
	}
	if got.Value != "USB Type-C" || got.Active || got.UsedCount != 4 {
		t.Fatalf("updated = %+v", got)
	}
	dup := "microusb"
	if _, err := s.UpdateValue(a.ID, &request.DeviceCatalogValueUpdate{Value: &dup}); !utils.IsValidationError(err) {
		t.Fatalf("rename onto another value: %v", err)
	}
	same := "usb type-c" // case change of itself is fine
	if _, err := s.UpdateValue(a.ID, &request.DeviceCatalogValueUpdate{Value: &same}); err != nil {
		t.Fatalf("case change: %v", err)
	}
}

func TestDeviceCatalogDeleteOnlyUnused(t *testing.T) {
	s, repo := newCatalogService()
	used := mustCreate(t, s, "ram", "8GB")
	unused := mustCreate(t, s, "ram", "yyy")
	repo.used["memory_ram"] = map[string]int{"8GB": 3}

	err := s.DeleteValue(used.ID)
	if !utils.IsValidationError(err) || err.Error() != "This value is used by 3 devices: deactivate it instead" {
		t.Fatalf("used: %v", err)
	}
	if err := s.DeleteValue(unused.ID); err != nil {
		t.Fatal(err)
	}
	if len(repo.values) != 1 {
		t.Fatalf("values left = %d", len(repo.values))
	}
	if DeviceCatalogInUseMessage(1) != "This value is used by 1 device: deactivate it instead" {
		t.Error("singular message")
	}
}

func TestDeviceCatalogSetOrder(t *testing.T) {
	s, _ := newCatalogService()
	a := mustCreate(t, s, "device_type", "Smartphone")
	b := mustCreate(t, s, "device_type", "Tablet")
	c := mustCreate(t, s, "device_type", "IoT")

	got, err := s.SetOrder("device_type", &request.DeviceCatalogOrder{IDs: []string{c.ID, a.ID, b.ID}})
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, v := range got {
		names = append(names, v.Value)
	}
	if !reflect.DeepEqual(names, []string{"IoT", "Smartphone", "Tablet"}) {
		t.Fatalf("order = %v", names)
	}
	if _, err := s.SetOrder("device_type", &request.DeviceCatalogOrder{IDs: []string{a.ID, b.ID}}); !utils.IsValidationError(err) {
		t.Fatalf("missing id: %v", err)
	}
	if _, err := s.SetOrder("device_type", &request.DeviceCatalogOrder{IDs: []string{a.ID, a.ID, b.ID}}); !utils.IsValidationError(err) {
		t.Fatalf("repeated id: %v", err)
	}
}

func TestDeviceCatalogListHasEveryCategoryWithUsage(t *testing.T) {
	s, repo := newCatalogService()
	mustCreate(t, s, "rom", "128GB")
	repo.used["memory_rom"] = map[string]int{"128GB": 2}

	list, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	for _, c := range list {
		keys = append(keys, c.Category)
		if c.Values == nil {
			t.Errorf("%s values must be [] not null", c.Category)
		}
	}
	if !reflect.DeepEqual(keys, []string{"device_type", "rom", "ram", "memory_capability", "usb"}) {
		t.Fatalf("categories = %v", keys)
	}
	if list[1].Label != "ROM" || list[1].Group != "Memory" || list[1].Values[0].UsedCount != 2 {
		t.Fatalf("rom = %+v", list[1])
	}
}

// GET /configuration: the five lists are the ACTIVE catalog values in order;
// a category with no catalog values keeps the stored list.
func TestConfigurationGetUsesActiveCatalogValues(t *testing.T) {
	catalogs, _ := newCatalogService()
	mustCreate(t, catalogs, "rom", "64GB")
	off := mustCreate(t, catalogs, "rom", "old")
	mustCreate(t, catalogs, "rom", "128GB")
	inactive := false
	if _, err := catalogs.UpdateValue(off.ID, &request.DeviceCatalogValueUpdate{Active: &inactive}); err != nil {
		t.Fatal(err)
	}

	cfgRepo := &memConfigRepo{stored: &responses.Configuration{
		Roms: []string{"legacy"}, UsbCommunications: []string{"USB-C"}, LteBands: []string{"B28"},
	}}
	s := &configurationService{configurationRepository: cfgRepo, deviceCatalogs: catalogs}
	got, err := s.Get()
	if err != nil {
		t.Fatal(err)
	}
	cfg := got[0]
	if !reflect.DeepEqual(cfg.Roms, []string{"64GB", "128GB"}) {
		t.Errorf("Roms = %v", cfg.Roms)
	}
	if !reflect.DeepEqual(cfg.UsbCommunications, []string{"USB-C"}) || !reflect.DeepEqual(cfg.LteBands, []string{"B28"}) {
		t.Errorf("fallback/other keys changed: %+v", cfg)
	}
}

// POST /configuration ignores the five catalog-backed lists.
func TestConfigurationCreateIgnoresCatalogLists(t *testing.T) {
	cfgRepo := &memConfigRepo{stored: &responses.Configuration{Roms: []string{"64GB"}, Rams: []string{"8GB"}}}
	s := &configurationService{configurationRepository: cfgRepo}
	_, err := s.Create(&request.Configuration{
		Roms: []string{"8GB"}, Rams: []string{"64GB"}, LteBands: []string{"B7"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfgRepo.saved.Roms, []string{"64GB"}) || !reflect.DeepEqual(cfgRepo.saved.Rams, []string{"8GB"}) ||
		!reflect.DeepEqual(cfgRepo.saved.LteBands, []string{"B7"}) {
		t.Fatalf("saved = %+v", cfgRepo.saved)
	}
}

type memConfigRepo struct {
	stored *responses.Configuration
	saved  *models.Configuration
}

func (m *memConfigRepo) Create(c *models.Configuration) (*primitive.ObjectID, error) {
	m.saved = c
	id := primitive.NewObjectID()
	return &id, nil
}
func (m *memConfigRepo) Get() ([]*responses.Configuration, error) {
	if m.stored == nil {
		return []*responses.Configuration{}, nil
	}
	c := *m.stored
	return []*responses.Configuration{&c}, nil
}
