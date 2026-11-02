package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	imeiutil "github.com/abisaidfarias/lbtechapi/utils/imei"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/sync/singleflight"
)

type IImeiValidationService interface {
	Lookup(userID string, rawIMEI string) (*responses.ImeiLookupResult, error)
	ListHistory(userID string, filter repositories.ImeiValidationListFilter) (*responses.ImeiQueryHistoryPage, error)
	ExportHistory(userID string, filter repositories.ImeiValidationListFilter) ([]byte, error)
}

type imeiValidationService struct {
	repo           repositories.IImeiValidationRepository
	userRepository repositoriesUser
	imeiInfoClient IImeiInfoClient
	imeiInfoLimits utils.ImeiInfoLimits
	tacLookupGroup singleflight.Group
}

func NewImeiValidationService(
	repo repositories.IImeiValidationRepository,
	userRepository repositoriesUser,
	imeiInfoClient IImeiInfoClient,
) IImeiValidationService {
	return &imeiValidationService{
		repo:           repo,
		userRepository: userRepository,
		imeiInfoClient: imeiInfoClient,
		imeiInfoLimits: utils.LoadImeiInfoLimits(),
	}
}

func (s *imeiValidationService) Lookup(userID string, rawIMEI string) (*responses.ImeiLookupResult, error) {
	user, err := s.requireInternalUser(userID)
	if err != nil {
		return nil, err
	}

	imei := imeiutil.Normalize(rawIMEI)
	if len(imei) != 15 {
		return nil, utils.NewValidationError("IMEI inválido: debe tener 15 dígitos")
	}
	if !imeiutil.IsValidLuhn(imei) {
		return nil, utils.NewValidationError("IMEI inválido: dígito verificador incorrecto")
	}

	tac := imeiutil.TAC(imei)
	now := time.Now().UTC()

	cached, err := s.repo.GetTacByID(tac)
	if err != nil {
		return nil, err
	}

	result := &responses.ImeiLookupResult{
		IMEI:      imei,
		TAC:       tac,
		QueriedAt: now.Format(time.RFC3339),
	}

	query := &models.ImeiQuery{
		IMEI:      imei,
		TAC:       tac,
		UserID:    user.ID,
		QueriedAt: now,
	}

	if cached != nil {
		result.Brand = cached.Brand
		result.TechnicalModel = cached.TechnicalModel
		result.CommercialName = cached.CommercialName
		result.Source = "BASE LOCAL"

		query.Brand = cached.Brand
		query.TechnicalModel = cached.TechnicalModel
		query.CommercialName = cached.CommercialName
		query.Source = models.ImeiQuerySourceLocal
	} else {
		catalog, lookupErr := s.resolveTacFromImeiInfo(user.ID, imei, tac, now)
		if lookupErr != nil {
			if errors.Is(lookupErr, utils.ErrorImeiInfoRateLimited) {
				return nil, lookupErr
			}
			result.Source = "ERROR"
			result.ErrorMessage = lookupErr.Error()
			query.Source = models.ImeiQuerySourceError
			query.ErrorMessage = lookupErr.Error()
			if err := s.repo.InsertQuery(query); err != nil {
				return nil, err
			}
			return result, nil
		}

		result.Brand = catalog.Brand
		result.TechnicalModel = catalog.TechnicalModel
		result.CommercialName = catalog.CommercialName
		result.Source = "IMEI.INFO"

		query.Brand = catalog.Brand
		query.TechnicalModel = catalog.TechnicalModel
		query.CommercialName = catalog.CommercialName
		query.Source = models.ImeiQuerySourceImeiInfo
	}

	if err := s.repo.InsertQuery(query); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *imeiValidationService) ListHistory(
	userID string,
	filter repositories.ImeiValidationListFilter,
) (*responses.ImeiQueryHistoryPage, error) {
	if _, err := s.requireInternalUser(userID); err != nil {
		return nil, err
	}

	rows, total, err := s.repo.ListQueries(filter)
	if err != nil {
		return nil, err
	}

	userNames := map[string]string{}
	items := make([]responses.ImeiQueryHistoryItem, 0, len(rows))
	for _, row := range rows {
		name := userNames[row.UserID.Hex()]
		if name == "" {
			u, err := s.userRepository.GetByID(row.UserID.Hex())
			if err == nil && u != nil {
				name = strings.TrimSpace(u.Name + " " + u.LastName)
				userNames[row.UserID.Hex()] = name
			}
		}
		items = append(items, responses.ImeiQueryHistoryItem{
			ID:             row.ID.Hex(),
			IMEI:           row.IMEI,
			TAC:            row.TAC,
			Brand:          row.Brand,
			TechnicalModel: row.TechnicalModel,
			CommercialName: row.CommercialName,
			Source:         mapImeiSourceLabel(row.Source),
			ErrorMessage:   row.ErrorMessage,
			UserID:         row.UserID.Hex(),
			UserName:       name,
			QueriedAt:      row.QueriedAt.UTC().Format(time.RFC3339),
		})
	}

	return &responses.ImeiQueryHistoryPage{
		Items:    items,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (s *imeiValidationService) ExportHistory(
	userID string,
	filter repositories.ImeiValidationListFilter,
) ([]byte, error) {
	filter.Page = 1
	filter.PageSize = 10000
	page, err := s.ListHistory(userID, filter)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	headers := []string{"IMEI", "Marca", "Modelo técnico", "Modelo comercial", "Fecha", "Usuario", "Origen"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	for i, row := range page.Items {
		r := i + 2
		_ = f.SetCellValue(sheet, cellName(1, r), row.IMEI)
		_ = f.SetCellValue(sheet, cellName(2, r), row.Brand)
		_ = f.SetCellValue(sheet, cellName(3, r), row.TechnicalModel)
		_ = f.SetCellValue(sheet, cellName(4, r), row.CommercialName)
		qt := parseRFC3339(row.QueriedAt)
		_ = f.SetCellValue(sheet, cellName(5, r), functions.ExcelDate(&qt))
		_ = f.SetCellValue(sheet, cellName(6, r), row.UserName)
		_ = f.SetCellValue(sheet, cellName(7, r), row.Source)
	}

	buf := &bytes.Buffer{}
	if err := f.Write(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *imeiValidationService) resolveTacFromImeiInfo(
	userID primitive.ObjectID,
	imei, tac string,
	now time.Time,
) (*models.TacCatalog, error) {
	v, err, _ := s.tacLookupGroup.Do(tac, func() (interface{}, error) {
		cached, err := s.repo.GetTacByID(tac)
		if err != nil {
			return nil, err
		}
		if cached != nil {
			return cached, nil
		}

		if err := s.assertImeiInfoRateLimits(userID, imei, now); err != nil {
			return nil, err
		}

		ctx := context.Background()
		if err := s.repo.InsertImeiInfoUsage(userID, imei, tac, now); err != nil {
			return nil, err
		}

		lookup, lookupErr := s.imeiInfoClient.Lookup(ctx, imei)
		if lookupErr != nil {
			return nil, lookupErr
		}

		catalog := &models.TacCatalog{
			TAC:            tac,
			Brand:          lookup.Brand,
			TechnicalModel: lookup.TechnicalModel,
			CommercialName: lookup.CommercialName,
			SourceIMEI:     imei,
			RawResponse:    primitiveMFromMap(lookup.Raw),
			CreatedAt:      now,
		}
		if err := s.repo.UpsertTac(catalog); err != nil {
			return nil, err
		}
		return catalog, nil
	})
	if err != nil {
		return nil, err
	}
	catalog, ok := v.(*models.TacCatalog)
	if !ok || catalog == nil {
		return nil, fmt.Errorf("imei.info lookup failed")
	}
	return catalog, nil
}

func (s *imeiValidationService) assertImeiInfoRateLimits(
	userID primitive.ObjectID,
	imei string,
	now time.Time,
) error {
	dayStart := utils.StartOfUTCDay(now)
	cooldownSince := now.Add(-time.Duration(s.imeiInfoLimits.CooldownSameIMEISecs) * time.Second)

	hasRecent, err := s.repo.HasImeiInfoUsageSince(userID, imei, cooldownSince)
	if err != nil {
		return err
	}
	if hasRecent {
		return utils.ErrorImeiInfoRateLimited
	}

	userCount, err := s.repo.CountImeiInfoUsageSince(&userID, dayStart)
	if err != nil {
		return err
	}
	if int(userCount) >= s.imeiInfoLimits.MaxPerUserPerDay {
		return utils.ErrorImeiInfoRateLimited
	}

	globalCount, err := s.repo.CountImeiInfoUsageSince(nil, dayStart)
	if err != nil {
		return err
	}
	if int(globalCount) >= s.imeiInfoLimits.MaxGlobalPerDay {
		return utils.ErrorImeiInfoRateLimited
	}
	return nil
}

func (s *imeiValidationService) requireInternalUser(userID string) (*responses.User, error) {
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsInternal {
		return nil, fmt.Errorf("%w", utils.ErrorForbidden)
	}
	return user, nil
}

func mapImeiSourceLabel(source string) string {
	switch source {
	case models.ImeiQuerySourceLocal:
		return "BASE LOCAL"
	case models.ImeiQuerySourceImeiInfo:
		return "IMEI.INFO"
	case models.ImeiQuerySourceError:
		return "ERROR"
	default:
		return source
	}
}

func primitiveMFromMap(raw map[string]interface{}) primitive.M {
	if raw == nil {
		return nil
	}
	out := primitive.M{}
	for k, v := range raw {
		out[k] = v
	}
	return out
}

func cellName(col, row int) string {
	c, _ := excelize.CoordinatesToCellName(col, row)
	return c
}

func parseRFC3339(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return t
}

