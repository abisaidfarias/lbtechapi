package services

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ---- in-memory repository ----

type memImeiRepo struct {
	seq     int
	batches map[primitive.ObjectID]models.ImeiBatch
	records map[string]models.ImeiRecord
}

func newMemImeiRepo() *memImeiRepo {
	return &memImeiRepo{batches: map[primitive.ObjectID]models.ImeiBatch{}, records: map[string]models.ImeiRecord{}}
}

func (m *memImeiRepo) NextBatchCode(day time.Time) (string, error) {
	m.seq++
	return fmt.Sprintf("LOT-%s-%03d", day.Format("20060102"), m.seq), nil
}
func (m *memImeiRepo) CreateBatch(b *models.ImeiBatch) error { m.batches[b.ID] = *b; return nil }
func (m *memImeiRepo) GetBatch(id primitive.ObjectID) (*models.ImeiBatch, error) {
	if b, ok := m.batches[id]; ok {
		return &b, nil
	}
	return nil, nil
}
func (m *memImeiRepo) GetBatchesByIDs(ids []primitive.ObjectID) (map[primitive.ObjectID]models.ImeiBatch, error) {
	out := map[primitive.ObjectID]models.ImeiBatch{}
	for _, id := range ids {
		if b, ok := m.batches[id]; ok {
			out[id] = b
		}
	}
	return out, nil
}
func (m *memImeiRepo) ListBatches(f repositories.ImeiBatchListFilter) ([]models.ImeiBatch, int64, error) {
	var out []models.ImeiBatch
	for _, b := range m.batches {
		if f.Status != "" && b.Status != f.Status {
			continue
		}
		if f.CodeSearch != "" {
			match := strings.Contains(strings.ToLower(b.Code), strings.ToLower(f.CodeSearch))
			for _, id := range f.BatchIDs {
				match = match || id == b.ID
			}
			if !match {
				continue
			}
		}
		out = append(out, b)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code > out[j].Code })
	return out, int64(len(out)), nil
}
func (m *memImeiRepo) SetBatchTotal(id primitive.ObjectID, total int) error {
	b := m.batches[id]
	b.Total = total
	m.batches[id] = b
	return nil
}
func (m *memImeiRepo) AddBatchTotal(id primitive.ObjectID, delta int) (*models.ImeiBatch, error) {
	b, ok := m.batches[id]
	if !ok {
		return nil, nil
	}
	b.Total += delta
	m.batches[id] = b
	return &b, nil
}
func (m *memImeiRepo) RegisterBatch(id primitive.ObjectID, at time.Time, userID primitive.ObjectID, name string) error {
	for k, r := range m.records {
		if r.BatchID == id {
			r.Status = models.ImeiValidatorStatusRegistered
			r.RegisteredAt = &at
			r.RegisteredBy = userID
			m.records[k] = r
		}
	}
	b := m.batches[id]
	b.Status = models.ImeiValidatorStatusRegistered
	b.RegisteredAt = &at
	b.RegisteredBy = userID
	b.RegisteredByName = name
	m.batches[id] = b
	return nil
}
func (m *memImeiRepo) DeleteBatch(id primitive.ObjectID) error {
	for k, r := range m.records {
		if r.BatchID == id {
			delete(m.records, k)
		}
	}
	delete(m.batches, id)
	return nil
}
func (m *memImeiRepo) FindExisting(imeis []string) (map[string]models.ImeiRecord, error) {
	out := map[string]models.ImeiRecord{}
	for _, i := range imeis {
		if r, ok := m.records[i]; ok {
			out[i] = r
		}
	}
	return out, nil
}
func (m *memImeiRepo) InsertRecords(records []models.ImeiRecord) (int, []string, error) {
	n := 0
	var dups []string
	for _, r := range records {
		if _, ok := m.records[r.IMEI]; ok {
			dups = append(dups, r.IMEI)
			continue
		}
		m.records[r.IMEI] = r
		n++
	}
	return n, dups, nil
}
func (m *memImeiRepo) GetRecord(imei string) (*models.ImeiRecord, error) {
	if r, ok := m.records[imei]; ok {
		return &r, nil
	}
	return nil, nil
}
func (m *memImeiRepo) ListBatchRecords(id primitive.ObjectID, search string, page, limit int) ([]models.ImeiRecord, int64, error) {
	var out []models.ImeiRecord
	for _, r := range m.records {
		if r.BatchID == id && strings.HasPrefix(r.IMEI, search) {
			out = append(out, r)
		}
	}
	return out, int64(len(out)), nil
}
func (m *memImeiRepo) AllBatchRecords(id primitive.ObjectID) ([]models.ImeiRecord, error) {
	out, _, _ := m.ListBatchRecords(id, "", 1, 0)
	sort.Slice(out, func(i, j int) bool { return out[i].IMEI < out[j].IMEI })
	return out, nil
}
func (m *memImeiRepo) DeleteRecord(imei string) (bool, error) {
	_, ok := m.records[imei]
	delete(m.records, imei)
	return ok, nil
}
func (m *memImeiRepo) DeleteDevice(batchID primitive.ObjectID, serial string) ([]string, error) {
	removed := []string{}
	for k, r := range m.records {
		if r.BatchID == batchID && r.Serial == serial && r.Status == models.ImeiValidatorStatusReady {
			removed = append(removed, k)
			delete(m.records, k)
		}
	}
	sort.Strings(removed)
	return removed, nil
}
func (m *memImeiRepo) BatchIDsWithIMEIPrefix(prefix string, max int) ([]primitive.ObjectID, error) {
	seen := map[primitive.ObjectID]bool{}
	var ids []primitive.ObjectID
	for _, r := range m.records {
		if strings.HasPrefix(r.IMEI, prefix) && !seen[r.BatchID] {
			seen[r.BatchID] = true
			ids = append(ids, r.BatchID)
		}
	}
	return ids, nil
}

// ---- fixtures ----

func luhnIMEI(body string) string {
	sum := 0
	for i := 0; i < 14; i++ {
		n := int(body[i] - '0')
		if i%2 == 1 {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
	}
	return fmt.Sprintf("%s%d", body, (10-sum%10)%10)
}

// uploadXlsx builds an OABI-format CSV (no header) where every IMEI is its
// own Single SIM device (S/N "SN-<imei>", quantity 1). Row n = nth IMEI.
func uploadXlsx(t *testing.T, imeis ...string) *bytes.Reader {
	t.Helper()
	var b strings.Builder
	for _, imei := range imeis {
		fmt.Fprintf(&b, "%s;SN-%s;1\r\n", imei, imei)
	}
	return bytes.NewReader([]byte(b.String()))
}

// uploadCSV builds an OABI-format CSV from IMEI;S/N;quantity lines.
func uploadCSV(lines ...string) *bytes.Reader {
	return bytes.NewReader([]byte(strings.Join(lines, "\r\n") + "\r\n"))
}

func newValidatorForTest(internal bool) (*imeiValidatorService, *memImeiRepo) {
	repo := newMemImeiRepo()
	user := &responses.User{ID: primitive.NewObjectID(), Name: "Ana", LastName: "Soto", IsInternal: internal}
	s := &imeiValidatorService{
		repo:           repo,
		userRepository: stubUserRepo{user: user},
		now:            func() time.Time { return time.Date(2026, 11, 24, 15, 0, 0, 0, time.UTC) },
	}
	return s, repo
}

var (
	imeiA = luhnIMEI("35693803564380")
	imeiB = luhnIMEI("35693803564381")
	imeiC = luhnIMEI("35693803564382")
)

// ---- tests ----

func TestUploadCreatesBatchWithNewIMEIs(t *testing.T) {
	s, repo := newValidatorForTest(true)
	res, err := s.UploadBatch("u", "lote1.xlsx", uploadXlsx(t, imeiA, imeiB, imeiA, "123"))
	if err != nil {
		t.Fatal(err)
	}
	if res.Batch == nil || res.Batch.Code != "LOT-20261124-001" || res.Batch.Total != 2 ||
		res.Batch.Status != "ready" || res.Batch.CreatedByName != "Ana Soto" || res.Batch.FileName != "lote1.xlsx" {
		t.Fatalf("batch = %+v", res.Batch)
	}
	want := responses.ImeiBatchUploadSummary{TotalRows: 4, New: 2, DuplicatedInFile: 1, Invalid: 1}
	if res.Summary != want {
		t.Fatalf("summary = %+v, want %+v", res.Summary, want)
	}
	if len(res.DuplicatedInFile) != 1 || res.DuplicatedInFile[0].IMEI != imeiA ||
		fmt.Sprint(res.DuplicatedInFile[0].Rows) != "[1 3]" {
		t.Errorf("duplicated = %+v", res.DuplicatedInFile)
	}
	if len(res.Invalid) != 1 || res.Invalid[0].Row != 4 || res.Invalid[0].Reason != "format" {
		t.Errorf("invalid = %+v", res.Invalid)
	}
	if len(repo.records) != 2 {
		t.Errorf("stored %d records", len(repo.records))
	}
}

func TestUploadReportsExistingAndCreatesNoBatchWhenNothingIsNew(t *testing.T) {
	s, repo := newValidatorForTest(true)
	first, err := s.UploadBatch("u", "a.xlsx", uploadXlsx(t, imeiA, imeiB))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.RegisterBatch("u", first.Batch.ID); err != nil {
		t.Fatal(err)
	}
	second, err := s.UploadBatch("u", "b.xlsx", uploadXlsx(t, imeiC))
	if err != nil {
		t.Fatal(err)
	}

	res, err := s.UploadBatch("u", "c.xlsx", uploadXlsx(t, imeiA, imeiC))
	if err != nil {
		t.Fatal(err)
	}
	if res.Batch != nil {
		t.Fatalf("no batch expected, got %+v", res.Batch)
	}
	if len(repo.batches) != 2 {
		t.Fatalf("batches = %d, want 2", len(repo.batches))
	}
	if res.Summary.AlreadyRegistered != 1 || res.Summary.AlreadyReady != 1 || res.Summary.New != 0 {
		t.Fatalf("summary = %+v", res.Summary)
	}
	reg := res.AlreadyRegistered[0]
	if reg.IMEI != imeiA || reg.BatchCode != first.Batch.Code || reg.RegisteredAt == nil {
		t.Errorf("already_registered = %+v", reg)
	}
	rdy := res.AlreadyReady[0]
	if rdy.IMEI != imeiC || rdy.BatchID != second.Batch.ID || rdy.BatchCode != second.Batch.Code || rdy.CreatedAt == "" {
		t.Errorf("already_ready = %+v", rdy)
	}
}

func TestUploadRejectsEmptyFile(t *testing.T) {
	s, _ := newValidatorForTest(true)
	_, err := s.UploadBatch("u", "x.xlsx", uploadXlsx(t))
	if !utils.IsValidationError(err) {
		t.Fatalf("want validation error, got %v", err)
	}
}

func TestValidatorRequiresInternalUser(t *testing.T) {
	s, _ := newValidatorForTest(false)
	if _, err := s.UploadBatch("u", "x.xlsx", uploadXlsx(t, imeiA)); !errors.Is(err, utils.ErrorForbidden) {
		t.Fatalf("want forbidden, got %v", err)
	}
}

func TestRegisterBatchIsIdempotent(t *testing.T) {
	s, repo := newValidatorForTest(true)
	up, _ := s.UploadBatch("u", "a.xlsx", uploadXlsx(t, imeiA, imeiB))

	b1, err := s.RegisterBatch("u", up.Batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if b1.Status != "registered" || b1.RegisteredAt == nil || b1.RegisteredByName != "Ana Soto" {
		t.Fatalf("batch = %+v", b1)
	}
	for _, r := range repo.records {
		if r.Status != "registered" || r.RegisteredAt == nil {
			t.Fatalf("record not registered: %+v", r)
		}
	}
	b2, err := s.RegisterBatch("u", up.Batch.ID)
	if err != nil || *b2.RegisteredAt != *b1.RegisteredAt {
		t.Fatalf("second register changed the batch: %+v / %v", b2, err)
	}
}

func TestDeleteIMEIUpdatesTotalAndDeletesEmptyBatch(t *testing.T) {
	s, repo := newValidatorForTest(true)
	up, _ := s.UploadBatch("u", "a.xlsx", uploadXlsx(t, imeiA, imeiB))

	res, err := s.DeleteIMEI("u", up.Batch.ID, imeiA)
	if err != nil {
		t.Fatal(err)
	}
	if res.BatchDeleted || res.Batch == nil || res.Batch.Total != 1 {
		t.Fatalf("after first delete: %+v", res)
	}
	res, err = s.DeleteIMEI("u", up.Batch.ID, imeiB)
	if err != nil {
		t.Fatal(err)
	}
	if !res.BatchDeleted || res.Batch != nil {
		t.Fatalf("after last delete: %+v", res)
	}
	if len(repo.batches) != 0 || len(repo.records) != 0 {
		t.Fatalf("leftovers: %d batches, %d records", len(repo.batches), len(repo.records))
	}
}

func TestDeleteRegisteredIMEIOrBatchIsRejected(t *testing.T) {
	s, _ := newValidatorForTest(true)
	up, _ := s.UploadBatch("u", "a.xlsx", uploadXlsx(t, imeiA))
	if _, err := s.RegisterBatch("u", up.Batch.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DeleteIMEI("u", up.Batch.ID, imeiA); !utils.IsValidationError(err) || err.Error() != ImeiValidatorErrImeiRegistered {
		t.Errorf("delete registered IMEI: %v", err)
	}
	if err := s.DeleteBatch("u", up.Batch.ID); !utils.IsValidationError(err) || err.Error() != ImeiValidatorErrBatchRegistered {
		t.Errorf("delete registered batch: %v", err)
	}
}

func TestDeleteReadyBatchRemovesIMEIs(t *testing.T) {
	s, repo := newValidatorForTest(true)
	up, _ := s.UploadBatch("u", "a.xlsx", uploadXlsx(t, imeiA, imeiB))
	if err := s.DeleteBatch("u", up.Batch.ID); err != nil {
		t.Fatal(err)
	}
	if len(repo.batches) != 0 || len(repo.records) != 0 {
		t.Fatalf("leftovers: %d batches, %d records", len(repo.batches), len(repo.records))
	}
}

func TestDeleteIMEIFromAnotherBatchIsNotFound(t *testing.T) {
	s, _ := newValidatorForTest(true)
	a, _ := s.UploadBatch("u", "a.xlsx", uploadXlsx(t, imeiA))
	b, _ := s.UploadBatch("u", "b.xlsx", uploadXlsx(t, imeiB))
	if _, err := s.DeleteIMEI("u", b.Batch.ID, imeiA); !errors.Is(err, ErrImeiValidatorNotFound) {
		t.Fatalf("want not found, got %v (batch a %s)", err, a.Batch.ID)
	}
}

func TestLookupAndSearchByIMEI(t *testing.T) {
	s, _ := newValidatorForTest(true)
	a, _ := s.UploadBatch("u", "a.xlsx", uploadXlsx(t, imeiA))
	_, _ = s.UploadBatch("u", "b.xlsx", uploadXlsx(t, imeiB))

	found, err := s.LookupIMEI("u", imeiA)
	if err != nil || !found.Found || found.Status != "ready" || found.Batch.Code != a.Batch.Code {
		t.Fatalf("lookup = %+v, %v", found, err)
	}
	missing, err := s.LookupIMEI("u", imeiC)
	if err != nil || missing.Found {
		t.Fatalf("lookup missing = %+v, %v", missing, err)
	}

	page, err := s.ListBatches("u", "", imeiA, 1, 20)
	if err != nil || page.Total != 1 || page.Items[0].ID != a.Batch.ID {
		t.Fatalf("search by IMEI = %+v, %v", page, err)
	}
	page, err = s.ListBatches("u", "", "lot-20261124-002", 1, 20)
	if err != nil || page.Total != 1 {
		t.Fatalf("search by code = %+v, %v", page, err)
	}
	if _, err := s.ListBatches("u", "bogus", "", 1, 20); !utils.IsValidationError(err) {
		t.Fatalf("invalid status: %v", err)
	}
}

// A Dual SIM device (rows not contiguous) plus a Single SIM device: the OABI
// file has no header and keeps each device's IMEIs together.
func TestOabiFileGroupsDevices(t *testing.T) {
	s, _ := newValidatorForTest(true)
	up, err := s.UploadBatch("u", "real.csv", uploadCSV(
		imeiA+";83904/Y6UQ01392;2",
		imeiC+";83904/Y6UQ09999;1",
		imeiB+";83904/Y6UQ01392;2",
	))
	if err != nil || up.Batch == nil || up.Batch.Total != 3 {
		t.Fatalf("upload = %+v, %v", up, err)
	}
	file, name, err := s.OabiFile("u", up.Batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if name != up.Batch.Code+"-OABI.csv" {
		t.Errorf("name = %q", name)
	}
	want := imeiA + ";83904/Y6UQ01392;2\r\n" + imeiB + ";83904/Y6UQ01392;2\r\n" + imeiC + ";83904/Y6UQ09999;1\r\n"
	if string(file) != want {
		t.Fatalf("file = %q\nwant  %q", file, want)
	}
}

// A device with one IMEI already stored is not saved: its new IMEI is
// device_incomplete and no batch is created for it.
func TestUploadMixedDeviceIsNotSaved(t *testing.T) {
	s, repo := newValidatorForTest(true)
	if _, err := s.UploadBatch("u", "a.csv", uploadXlsx(t, imeiA)); err != nil {
		t.Fatal(err)
	}
	res, err := s.UploadBatch("u", "b.csv", uploadCSV(imeiA+";DUAL-1;2", imeiB+";DUAL-1;2"))
	if err != nil {
		t.Fatal(err)
	}
	if res.Batch != nil {
		t.Fatalf("no batch expected: %+v", res.Batch)
	}
	if res.Summary.AlreadyReady != 1 || res.Summary.Invalid != 1 || res.Summary.DeviceIncomplete != 1 {
		t.Fatalf("summary = %+v", res.Summary)
	}
	inv := res.Invalid[0]
	if inv.Value != imeiB || inv.Reason != "device_incomplete" || inv.Serial != "DUAL-1" || inv.Detail == "" {
		t.Fatalf("invalid = %+v", inv)
	}
	if _, stored := repo.records[imeiB]; stored {
		t.Fatal("the new IMEI of an incomplete device was stored")
	}
}

// Deleting one IMEI of a Dual SIM device deletes the whole device.
func TestDeleteIMEIDeletesWholeDevice(t *testing.T) {
	s, repo := newValidatorForTest(true)
	up, err := s.UploadBatch("u", "a.csv", uploadCSV(imeiA+";DUAL-1;2", imeiB+";DUAL-1;2", imeiC+";SINGLE-1;1"))
	if err != nil {
		t.Fatal(err)
	}
	res, err := s.DeleteIMEI("u", up.Batch.ID, imeiA)
	if err != nil {
		t.Fatal(err)
	}
	if res.BatchDeleted || res.Batch.Total != 1 || fmt.Sprint(res.DeletedIMEIs) != fmt.Sprint([]string{imeiA, imeiB}) {
		t.Fatalf("result = %+v", res)
	}
	if len(repo.records) != 1 || repo.records[imeiC].ImeiCount != 1 {
		t.Fatalf("records left = %+v", repo.records)
	}
}
