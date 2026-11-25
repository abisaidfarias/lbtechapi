package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IDeviceCatalogService interface {
	List() ([]responses.DeviceCatalog, error)
	CreateValue(category string, req *request.DeviceCatalogValueCreate) (*responses.DeviceCatalogValue, error)
	UpdateValue(id string, req *request.DeviceCatalogValueUpdate) (*responses.DeviceCatalogValue, error)
	SetOrder(category string, req *request.DeviceCatalogOrder) ([]responses.DeviceCatalogValue, error)
	DeleteValue(id string) error
	// ActiveValues returns, per category, the active values in order (the
	// legacy configuration arrays). Categories with no catalog values are
	// absent, so callers can fall back to the stored configuration.
	ActiveValues() (map[string][]string, error)
}

type deviceCatalogService struct {
	repo repositories.IDeviceCatalogRepository
	now  func() time.Time
}

func NewDeviceCatalogService(repo repositories.IDeviceCatalogRepository) IDeviceCatalogService {
	return &deviceCatalogService{repo: repo, now: time.Now}
}

func (s *deviceCatalogService) List() ([]responses.DeviceCatalog, error) {
	values, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	byCategory := map[string][]models.DeviceCatalogValue{}
	for _, v := range values {
		byCategory[v.Category] = append(byCategory[v.Category], v)
	}

	out := make([]responses.DeviceCatalog, 0, len(enums.DeviceCatalogCategories))
	for _, c := range enums.DeviceCatalogCategories {
		used, err := s.repo.UsedCounts(c.DeviceField)
		if err != nil {
			return nil, err
		}
		items := make([]responses.DeviceCatalogValue, 0, len(byCategory[c.Key]))
		for _, v := range byCategory[c.Key] {
			items = append(items, toDeviceCatalogValue(v, used[v.Value]))
		}
		out = append(out, responses.DeviceCatalog{Category: c.Key, Label: c.Label, Group: c.Group, Values: items})
	}
	return out, nil
}

func (s *deviceCatalogService) CreateValue(category string, req *request.DeviceCatalogValueCreate) (*responses.DeviceCatalogValue, error) {
	cat, ok := enums.DeviceCatalogCategoryByKey(category)
	if !ok {
		return nil, utils.NewValidationError("unknown catalog category")
	}
	value := strings.TrimSpace(req.Value)
	if value == "" {
		return nil, utils.NewValidationError("value is required")
	}
	if err := s.ensureUnique(cat, value, primitive.NilObjectID); err != nil {
		return nil, err
	}
	existing, err := s.repo.ListByCategory(cat.Key)
	if err != nil {
		return nil, err
	}
	order := 0
	for _, v := range existing {
		if v.Order >= order {
			order = v.Order + 1
		}
	}
	v := &models.DeviceCatalogValue{
		Category:  cat.Key,
		Value:     value,
		ValueKey:  enums.DeviceCatalogValueKey(value),
		Unit:      strings.TrimSpace(req.Unit),
		Active:    true,
		Order:     order,
		CreatedAt: s.now().UTC(),
	}
	if err := s.repo.Insert(v); err != nil {
		return nil, err
	}
	out := toDeviceCatalogValue(*v, 0)
	return &out, nil
}

// UpdateValue changes value/unit/active. Renaming a value in use does not
// rewrite devices: they keep the old text.
func (s *deviceCatalogService) UpdateValue(id string, req *request.DeviceCatalogValueUpdate) (*responses.DeviceCatalogValue, error) {
	v, cat, err := s.load(id)
	if err != nil {
		return nil, err
	}
	set := primitive.M{}
	if req.Value != nil {
		value := strings.TrimSpace(*req.Value)
		if value == "" {
			return nil, utils.NewValidationError("value is required")
		}
		if err := s.ensureUnique(cat, value, v.ID); err != nil {
			return nil, err
		}
		set["value"], set["value_key"] = value, enums.DeviceCatalogValueKey(value)
		v.Value = value
	}
	if req.Unit != nil {
		set["unit"] = strings.TrimSpace(*req.Unit)
		v.Unit = strings.TrimSpace(*req.Unit)
	}
	if req.Active != nil {
		set["active"] = *req.Active
		v.Active = *req.Active
	}
	if len(set) > 0 {
		if err := s.repo.UpdateFields(v.ID, set); err != nil {
			return nil, err
		}
	}
	used, err := s.repo.UsedCounts(cat.DeviceField)
	if err != nil {
		return nil, err
	}
	out := toDeviceCatalogValue(*v, used[v.Value])
	return &out, nil
}

// SetOrder requires exactly the ids of the category, each once.
func (s *deviceCatalogService) SetOrder(category string, req *request.DeviceCatalogOrder) ([]responses.DeviceCatalogValue, error) {
	cat, ok := enums.DeviceCatalogCategoryByKey(category)
	if !ok {
		return nil, utils.NewValidationError("unknown catalog category")
	}
	existing, err := s.repo.ListByCategory(cat.Key)
	if err != nil {
		return nil, err
	}
	known := make(map[primitive.ObjectID]bool, len(existing))
	for _, v := range existing {
		known[v.ID] = true
	}
	ids := make([]primitive.ObjectID, 0, len(req.IDs))
	seen := map[primitive.ObjectID]bool{}
	for _, raw := range req.IDs {
		id, err := primitive.ObjectIDFromHex(strings.TrimSpace(raw))
		if err != nil || !known[id] || seen[id] {
			return nil, utils.NewValidationError(fmt.Sprintf("invalid id in order: %s", raw))
		}
		seen[id] = true
		ids = append(ids, id)
	}
	if len(ids) != len(existing) {
		return nil, utils.NewValidationError("ids must list every value of the category exactly once")
	}
	if err := s.repo.SetOrder(cat.Key, ids); err != nil {
		return nil, err
	}
	all, err := s.List()
	if err != nil {
		return nil, err
	}
	for _, c := range all {
		if c.Category == cat.Key {
			return c.Values, nil
		}
	}
	return []responses.DeviceCatalogValue{}, nil
}

// DeleteValue only removes unused values; used ones are deactivated instead.
func (s *deviceCatalogService) DeleteValue(id string) error {
	v, cat, err := s.load(id)
	if err != nil {
		return err
	}
	used, err := s.repo.UsedCounts(cat.DeviceField)
	if err != nil {
		return err
	}
	if n := used[v.Value]; n > 0 {
		return utils.NewValidationError(DeviceCatalogInUseMessage(n))
	}
	return s.repo.Delete(v.ID)
}

func (s *deviceCatalogService) ActiveValues() (map[string][]string, error) {
	values, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	out := map[string][]string{}
	for _, v := range values {
		if _, ok := out[v.Category]; !ok {
			out[v.Category] = []string{}
		}
		if v.Active {
			out[v.Category] = append(out[v.Category], v.Value)
		}
	}
	return out, nil
}

// DeviceCatalogInUseMessage is the delete error for a used value.
func DeviceCatalogInUseMessage(n int) string {
	noun := "devices"
	if n == 1 {
		noun = "device"
	}
	return fmt.Sprintf("This value is used by %d %s: deactivate it instead", n, noun)
}

func (s *deviceCatalogService) ensureUnique(cat enums.DeviceCatalogCategory, value string, exclude primitive.ObjectID) error {
	other, err := s.repo.FindByKey(cat.Key, enums.DeviceCatalogValueKey(value), exclude)
	if err != nil {
		return err
	}
	if other != nil {
		return utils.NewValidationError(fmt.Sprintf("%q already exists in %s", other.Value, cat.Label))
	}
	return nil
}

func (s *deviceCatalogService) load(id string) (*models.DeviceCatalogValue, enums.DeviceCatalogCategory, error) {
	oid, err := primitive.ObjectIDFromHex(strings.TrimSpace(id))
	if err != nil {
		return nil, enums.DeviceCatalogCategory{}, utils.NewValidationError("invalid catalog value id")
	}
	v, err := s.repo.Get(oid)
	if err != nil {
		return nil, enums.DeviceCatalogCategory{}, err
	}
	if v == nil {
		return nil, enums.DeviceCatalogCategory{}, utils.NewValidationError("catalog value not found")
	}
	cat, _ := enums.DeviceCatalogCategoryByKey(v.Category)
	return v, cat, nil
}

func toDeviceCatalogValue(v models.DeviceCatalogValue, used int) responses.DeviceCatalogValue {
	return responses.DeviceCatalogValue{
		ID: v.ID.Hex(), Value: v.Value, Unit: v.Unit, Active: v.Active, Order: v.Order, UsedCount: used,
	}
}
