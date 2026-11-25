package services

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/imeivalidator"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Error messages of the IMEI Validator (400 unless noted).
const (
	ImeiValidatorErrBatchNotFound     = "batch not found"                                          // 404
	ImeiValidatorErrImeiNotInBatch    = "IMEI not found in this batch"                             // 404
	ImeiValidatorErrImeiRegistered    = "this IMEI is already registered in OABI and cannot be deleted"
	ImeiValidatorErrBatchRegistered   = "this batch is already registered in OABI and cannot be deleted"
	ImeiValidatorErrInvalidBatchID    = "invalid batch id"
	ImeiValidatorErrInvalidStatus     = "invalid status: use \"ready\" or \"registered\""
	ImeiValidatorErrFileRequired      = "file is required"
	imeiValidatorMaxSearchBatchIDs    = 1000
	imeiValidatorDefaultLimit         = 20
	imeiValidatorMaxLimit             = 200
	imeiValidatorMinIMEISearchDigits  = 4
)

// ErrImeiValidatorNotFound marks the 404 cases.
var ErrImeiValidatorNotFound = errors.New("not found")

type IImeiValidatorService interface {
	Template(userID string) ([]byte, error)
	UploadBatch(userID, fileName string, file io.Reader) (*responses.ImeiBatchUploadResult, error)
	ListBatches(userID, status, search string, page, limit int) (*responses.ImeiBatchPage, error)
	ListBatchIMEIs(userID, batchID, search string, page, limit int) (*responses.ImeiBatchRecordPage, error)
	RegisterBatch(userID, batchID string) (*responses.ImeiBatch, error)
	DeleteIMEI(userID, batchID, imei string) (*responses.ImeiDeleteResult, error)
	DeleteBatch(userID, batchID string) error
	LookupIMEI(userID, imei string) (*responses.ImeiValidatorLookup, error)
	OabiFile(userID, batchID string) ([]byte, string, error)
}

type imeiValidatorService struct {
	repo           repositories.IImeiValidatorRepository
	userRepository repositoriesUser
	now            func() time.Time
}

func NewImeiValidatorService(repo repositories.IImeiValidatorRepository, userRepository repositoriesUser) IImeiValidatorService {
	return &imeiValidatorService{repo: repo, userRepository: userRepository, now: time.Now}
}

// Same rule as the existing imei-validation endpoints: internal users only.
func (s *imeiValidatorService) requireInternal(userID string) (*responses.User, error) {
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsInternal {
		return nil, fmt.Errorf("%w", utils.ErrorForbidden)
	}
	return user, nil
}

func userDisplayName(u *responses.User) string {
	return strings.TrimSpace(u.Name + " " + u.LastName)
}

func (s *imeiValidatorService) Template(userID string) ([]byte, error) {
	if _, err := s.requireInternal(userID); err != nil {
		return nil, err
	}
	return imeivalidator.BuildTemplate()
}

// ---- Upload ----

func (s *imeiValidatorService) UploadBatch(userID, fileName string, file io.Reader) (*responses.ImeiBatchUploadResult, error) {
	user, err := s.requireInternal(userID)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, utils.NewValidationError(ImeiValidatorErrFileRequired)
	}

	rows, err := imeivalidator.ParseUpload(file)
	if err != nil {
		return nil, utils.NewValidationError(err.Error())
	}

	// One bulk existence lookup for the whole file (chunked $in in the repo).
	stored, err := s.repo.FindExisting(imeivalidator.ValidIMEIs(rows))
	if err != nil {
		return nil, err
	}
	existing := make(map[string]imeivalidator.Existing, len(stored))
	for imei, rec := range stored {
		existing[imei] = imeivalidator.Existing{Status: rec.Status}
	}
	c := imeivalidator.Classify(rows, existing)

	result := &responses.ImeiBatchUploadResult{
		AlreadyReady:      []responses.ImeiAlreadyReady{},
		AlreadyRegistered: []responses.ImeiAlreadyRegistered{},
		DuplicatedInFile:  []responses.ImeiDuplicatedInFile{},
		Invalid:           []responses.ImeiInvalidRow{},
	}

	// The batch exists only when at least one IMEI is new.
	if len(c.New) > 0 {
		batch, raced, err := s.createBatch(user, fileName, c.New)
		if err != nil {
			return nil, err
		}
		result.Batch = batch
		if len(raced) > 0 {
			if err := s.reportRacedDevices(&c, raced, stored); err != nil {
				return nil, err
			}
		}
	}

	if err := s.fillExisting(result, c, stored); err != nil {
		return nil, err
	}
	for _, d := range c.Duplicates {
		result.DuplicatedInFile = append(result.DuplicatedInFile, responses.ImeiDuplicatedInFile{IMEI: d.IMEI, Rows: d.Rows})
	}
	deviceIncomplete := 0
	for _, inv := range c.Invalid {
		if inv.Reason == imeivalidator.ReasonDeviceIncomplete {
			deviceIncomplete++
		}
		result.Invalid = append(result.Invalid, responses.ImeiInvalidRow{
			Row: inv.Row, Value: inv.Value, Serial: inv.Serial, Reason: inv.Reason, Detail: inv.Detail,
		})
	}

	newCount := 0
	if result.Batch != nil {
		newCount = result.Batch.Total
	}
	duplicateRows := 0
	for _, d := range c.Duplicates {
		duplicateRows += len(d.Rows) - 1 // the first occurrence is classified normally
	}
	result.Summary = responses.ImeiBatchUploadSummary{
		TotalRows:         c.TotalRows,
		New:               newCount,
		AlreadyReady:      len(result.AlreadyReady),
		AlreadyRegistered: len(result.AlreadyRegistered),
		DuplicatedInFile:  duplicateRows,
		Invalid:           len(result.Invalid),
		DeviceIncomplete:  deviceIncomplete,
	}
	return result, nil
}

// reportRacedDevices handles IMEIs saved by a concurrent upload between the
// lookup and the insert. They are reported as already saved, and the other
// IMEIs of their devices, already inserted in this batch, were removed by
// createBatch and are reported as device_incomplete.
func (s *imeiValidatorService) reportRacedDevices(
	c *imeivalidator.Classification, raced []string, stored map[string]models.ImeiRecord,
) error {
	more, err := s.repo.FindExisting(raced)
	if err != nil {
		return err
	}
	isRaced := map[string]bool{}
	for imei, rec := range more {
		isRaced[imei] = true
		stored[imei] = rec
		if rec.Status == models.ImeiValidatorStatusRegistered {
			c.AlreadyRegistered = append(c.AlreadyRegistered, imei)
		} else {
			c.AlreadyReady = append(c.AlreadyReady, imei)
		}
	}
	brokenSerials := map[string]bool{}
	for _, r := range c.New {
		if isRaced[r.IMEI] {
			brokenSerials[r.Serial] = true
		}
	}
	kept := c.New[:0]
	for _, r := range c.New {
		switch {
		case isRaced[r.IMEI]:
		case brokenSerials[r.Serial]:
			c.Invalid = append(c.Invalid, imeivalidator.InvalidRow{
				Row: r.Row, Value: r.IMEI, Serial: r.Serial, Reason: imeivalidator.ReasonDeviceIncomplete,
				Detail: "S/N " + r.Serial + ": not saved because another IMEI of this device was saved by another upload at the same time",
			})
		default:
			kept = append(kept, r)
		}
	}
	c.New = kept
	return nil
}

// createBatch stores the batch and its new IMEIs. It returns the IMEIs lost
// to a concurrent upload; if all of them were lost the batch is removed.
func (s *imeiValidatorService) createBatch(
	user *responses.User, fileName string, rows []imeivalidator.NewRecord,
) (*responses.ImeiBatch, []string, error) {
	now := s.now().UTC()
	code, err := s.repo.NextBatchCode(now)
	if err != nil {
		return nil, nil, err
	}
	batch := &models.ImeiBatch{
		ID:            primitive.NewObjectID(),
		Code:          code,
		FileName:      strings.TrimSpace(fileName),
		CreatedBy:     user.ID,
		CreatedByName: userDisplayName(user),
		CreatedAt:     now,
		Total:         len(rows),
		Status:        models.ImeiValidatorStatusReady,
	}
	if err := s.repo.CreateBatch(batch); err != nil {
		return nil, nil, err
	}

	records := make([]models.ImeiRecord, len(rows))
	for i, r := range rows {
		records[i] = models.ImeiRecord{
			IMEI:      r.IMEI,
			BatchID:   batch.ID,
			Status:    models.ImeiValidatorStatusReady,
			Serial:    r.Serial,
			ImeiCount: r.ImeiCount,
			CreatedAt: now,
		}
	}
	inserted, raced, err := s.repo.InsertRecords(records)
	if err != nil {
		return nil, nil, err
	}
	if len(raced) > 0 {
		// A device that lost an IMEI to a concurrent upload would be
		// incomplete in the OABI file: drop its other IMEIs from this batch.
		isRaced := map[string]bool{}
		for _, imei := range raced {
			isRaced[imei] = true
		}
		serials := map[string]bool{}
		for _, r := range rows {
			if isRaced[r.IMEI] {
				serials[r.Serial] = true
			}
		}
		for serial := range serials {
			removed, err := s.repo.DeleteDevice(batch.ID, serial)
			if err != nil {
				return nil, nil, err
			}
			inserted -= len(removed)
		}
	}
	if inserted == 0 {
		if err := s.repo.DeleteBatch(batch.ID); err != nil {
			return nil, nil, err
		}
		return nil, raced, nil
	}
	if inserted != batch.Total {
		batch.Total = inserted
		if err := s.repo.SetBatchTotal(batch.ID, inserted); err != nil {
			return nil, nil, err
		}
	}
	out := toImeiBatchResponse(*batch)
	return &out, raced, nil
}

func (s *imeiValidatorService) fillExisting(
	result *responses.ImeiBatchUploadResult,
	c imeivalidator.Classification,
	stored map[string]models.ImeiRecord,
) error {
	ids := map[primitive.ObjectID]struct{}{}
	for _, imei := range append(append([]string{}, c.AlreadyReady...), c.AlreadyRegistered...) {
		ids[stored[imei].BatchID] = struct{}{}
	}
	idList := make([]primitive.ObjectID, 0, len(ids))
	for id := range ids {
		idList = append(idList, id)
	}
	batches, err := s.repo.GetBatchesByIDs(idList)
	if err != nil {
		return err
	}
	for _, imei := range c.AlreadyReady {
		rec := stored[imei]
		result.AlreadyReady = append(result.AlreadyReady, responses.ImeiAlreadyReady{
			IMEI:      imei,
			BatchID:   rec.BatchID.Hex(),
			BatchCode: batches[rec.BatchID].Code,
			CreatedAt: formatTime(rec.CreatedAt),
		})
	}
	for _, imei := range c.AlreadyRegistered {
		rec := stored[imei]
		result.AlreadyRegistered = append(result.AlreadyRegistered, responses.ImeiAlreadyRegistered{
			IMEI:         imei,
			BatchID:      rec.BatchID.Hex(),
			BatchCode:    batches[rec.BatchID].Code,
			RegisteredAt: formatTimePtr(rec.RegisteredAt),
		})
	}
	return nil
}

// ---- Lists ----

func normalizePaging(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = imeiValidatorDefaultLimit
	}
	if limit > imeiValidatorMaxLimit {
		limit = imeiValidatorMaxLimit
	}
	return page, limit
}

func (s *imeiValidatorService) ListBatches(userID, status, search string, page, limit int) (*responses.ImeiBatchPage, error) {
	if _, err := s.requireInternal(userID); err != nil {
		return nil, err
	}
	status = strings.TrimSpace(status)
	if status != "" && status != models.ImeiValidatorStatusReady && status != models.ImeiValidatorStatusRegistered {
		return nil, utils.NewValidationError(ImeiValidatorErrInvalidStatus)
	}
	page, limit = normalizePaging(page, limit)
	filter := repositories.ImeiBatchListFilter{Status: status, Page: page, Limit: limit}

	// search matches the batch code, or a batch holding an IMEI starting with
	// those digits.
	if search = strings.TrimSpace(search); search != "" {
		filter.CodeSearch = search
		if digits, ok := imeivalidator.NormalizeIMEI(search); ok && len(digits) >= imeiValidatorMinIMEISearchDigits {
			ids, err := s.repo.BatchIDsWithIMEIPrefix(digits, imeiValidatorMaxSearchBatchIDs)
			if err != nil {
				return nil, err
			}
			filter.BatchIDs = ids
		}
	}

	batches, total, err := s.repo.ListBatches(filter)
	if err != nil {
		return nil, err
	}
	out := &responses.ImeiBatchPage{Items: make([]responses.ImeiBatch, 0, len(batches)), Total: total}
	for _, b := range batches {
		out.Items = append(out.Items, toImeiBatchResponse(b))
	}
	return out, nil
}

func (s *imeiValidatorService) ListBatchIMEIs(userID, batchID, search string, page, limit int) (*responses.ImeiBatchRecordPage, error) {
	if _, err := s.requireInternal(userID); err != nil {
		return nil, err
	}
	batch, err := s.loadBatch(batchID)
	if err != nil {
		return nil, err
	}
	page, limit = normalizePaging(page, limit)
	digits, _ := imeivalidator.NormalizeIMEI(strings.TrimSpace(search))
	records, total, err := s.repo.ListBatchRecords(batch.ID, digits, page, limit)
	if err != nil {
		return nil, err
	}
	out := &responses.ImeiBatchRecordPage{Items: make([]responses.ImeiBatchRecord, 0, len(records)), Total: total}
	for _, r := range records {
		out.Items = append(out.Items, responses.ImeiBatchRecord{
			IMEI:         r.IMEI,
			Serial:       r.Serial,
			ImeiCount:    r.ImeiCount,
			Status:       r.Status,
			CreatedAt:    formatTime(r.CreatedAt),
			RegisteredAt: formatTimePtr(r.RegisteredAt),
		})
	}
	return out, nil
}

// ---- Register / delete ----

// RegisterBatch marks the batch and all its IMEIs as registered in OABI.
// Registering an already registered batch returns it unchanged.
func (s *imeiValidatorService) RegisterBatch(userID, batchID string) (*responses.ImeiBatch, error) {
	user, err := s.requireInternal(userID)
	if err != nil {
		return nil, err
	}
	batch, err := s.loadBatch(batchID)
	if err != nil {
		return nil, err
	}
	if batch.Status != models.ImeiValidatorStatusRegistered {
		if err := s.repo.RegisterBatch(batch.ID, s.now().UTC(), user.ID, userDisplayName(user)); err != nil {
			return nil, err
		}
		if batch, err = s.loadBatch(batchID); err != nil {
			return nil, err
		}
	}
	out := toImeiBatchResponse(*batch)
	return &out, nil
}

// DeleteIMEI removes a ready IMEI. When it was the last one of its batch the
// batch is removed too.
func (s *imeiValidatorService) DeleteIMEI(userID, batchID, rawIMEI string) (*responses.ImeiDeleteResult, error) {
	if _, err := s.requireInternal(userID); err != nil {
		return nil, err
	}
	batch, err := s.loadBatch(batchID)
	if err != nil {
		return nil, err
	}
	imei, _ := imeivalidator.NormalizeIMEI(rawIMEI)
	rec, err := s.repo.GetRecord(imei)
	if err != nil {
		return nil, err
	}
	if rec == nil || rec.BatchID != batch.ID {
		return nil, fmt.Errorf("%s: %w", ImeiValidatorErrImeiNotInBatch, ErrImeiValidatorNotFound)
	}
	if rec.Status == models.ImeiValidatorStatusRegistered {
		return nil, utils.NewValidationError(ImeiValidatorErrImeiRegistered)
	}
	// A device goes to OABI whole: deleting one of its IMEIs deletes every
	// IMEI of the same S/N in this batch. Records without S/N (none in the
	// real format) are deleted alone.
	var removed []string
	if strings.TrimSpace(rec.Serial) != "" {
		if removed, err = s.repo.DeleteDevice(batch.ID, rec.Serial); err != nil {
			return nil, err
		}
	} else {
		deleted, err := s.repo.DeleteRecord(imei)
		if err != nil {
			return nil, err
		}
		if deleted {
			removed = []string{imei}
		}
	}
	if len(removed) == 0 {
		return nil, fmt.Errorf("%s: %w", ImeiValidatorErrImeiNotInBatch, ErrImeiValidatorNotFound)
	}

	updated, err := s.repo.AddBatchTotal(batch.ID, -len(removed))
	if err != nil {
		return nil, err
	}
	if updated == nil || updated.Total <= 0 {
		if err := s.repo.DeleteBatch(batch.ID); err != nil {
			return nil, err
		}
		return &responses.ImeiDeleteResult{BatchDeleted: true, Batch: nil, DeletedIMEIs: removed}, nil
	}
	out := toImeiBatchResponse(*updated)
	return &responses.ImeiDeleteResult{BatchDeleted: false, Batch: &out, DeletedIMEIs: removed}, nil
}

func (s *imeiValidatorService) DeleteBatch(userID, batchID string) error {
	if _, err := s.requireInternal(userID); err != nil {
		return err
	}
	batch, err := s.loadBatch(batchID)
	if err != nil {
		return err
	}
	if batch.Status != models.ImeiValidatorStatusReady {
		return utils.NewValidationError(ImeiValidatorErrBatchRegistered)
	}
	return s.repo.DeleteBatch(batch.ID)
}

// ---- Lookup / OABI file ----

func (s *imeiValidatorService) LookupIMEI(userID, rawIMEI string) (*responses.ImeiValidatorLookup, error) {
	if _, err := s.requireInternal(userID); err != nil {
		return nil, err
	}
	imei, ok := imeivalidator.NormalizeIMEI(rawIMEI)
	if !ok || imei == "" {
		return nil, utils.NewValidationError("invalid IMEI")
	}
	rec, err := s.repo.GetRecord(imei)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return &responses.ImeiValidatorLookup{Found: false, IMEI: imei}, nil
	}
	ref := &responses.ImeiBatchRef{ID: rec.BatchID.Hex()}
	if batch, err := s.repo.GetBatch(rec.BatchID); err != nil {
		return nil, err
	} else if batch != nil {
		ref.Code = batch.Code
	}
	return &responses.ImeiValidatorLookup{
		Found:        true,
		IMEI:         rec.IMEI,
		Status:       rec.Status,
		Batch:        ref,
		Serial:       rec.Serial,
		ImeiCount:    rec.ImeiCount,
		CreatedAt:    formatTime(rec.CreatedAt),
		RegisteredAt: formatTimePtr(rec.RegisteredAt),
	}, nil
}

func (s *imeiValidatorService) OabiFile(userID, batchID string) ([]byte, string, error) {
	if _, err := s.requireInternal(userID); err != nil {
		return nil, "", err
	}
	batch, err := s.loadBatch(batchID)
	if err != nil {
		return nil, "", err
	}
	records, err := s.repo.AllBatchRecords(batch.ID)
	if err != nil {
		return nil, "", err
	}
	rows := make([]imeivalidator.OabiRecord, len(records))
	for i, r := range records {
		rows[i] = imeivalidator.OabiRecord{IMEI: r.IMEI, Serial: r.Serial, ImeiCount: r.ImeiCount}
	}
	file, err := imeivalidator.BuildOabiFile(rows)
	if err != nil {
		return nil, "", err
	}
	return file, imeivalidator.OabiFileName(batch.Code), nil
}

// ---- helpers ----

func (s *imeiValidatorService) loadBatch(batchID string) (*models.ImeiBatch, error) {
	id, err := primitive.ObjectIDFromHex(strings.TrimSpace(batchID))
	if err != nil {
		return nil, utils.NewValidationError(ImeiValidatorErrInvalidBatchID)
	}
	batch, err := s.repo.GetBatch(id)
	if err != nil {
		return nil, err
	}
	if batch == nil {
		return nil, fmt.Errorf("%s: %w", ImeiValidatorErrBatchNotFound, ErrImeiValidatorNotFound)
	}
	return batch, nil
}

func toImeiBatchResponse(b models.ImeiBatch) responses.ImeiBatch {
	return responses.ImeiBatch{
		ID:               b.ID.Hex(),
		Code:             b.Code,
		FileName:         b.FileName,
		CreatedByName:    b.CreatedByName,
		CreatedAt:        formatTime(b.CreatedAt),
		Total:            b.Total,
		Status:           b.Status,
		RegisteredAt:     formatTimePtr(b.RegisteredAt),
		RegisteredByName: b.RegisteredByName,
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func formatTimePtr(t *time.Time) *string {
	if t == nil || t.IsZero() {
		return nil
	}
	v := formatTime(*t)
	return &v
}
