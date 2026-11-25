package repositories

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// imeiValidatorChunk bounds every $in lookup and InsertMany so 100k+ IMEI
// uploads stay well under the 16 MB command limit.
const imeiValidatorChunk = 5000

// ImeiBatchListFilter filters GET /batches. BatchIDs, when not nil, limits
// the code search to an OR with those batches (the ones holding the IMEI).
type ImeiBatchListFilter struct {
	Status     string
	CodeSearch string
	BatchIDs   []primitive.ObjectID
	Page       int
	Limit      int
}

type IImeiValidatorRepository interface {
	NextBatchCode(day time.Time) (string, error)
	CreateBatch(*models.ImeiBatch) error
	GetBatch(primitive.ObjectID) (*models.ImeiBatch, error)
	GetBatchesByIDs([]primitive.ObjectID) (map[primitive.ObjectID]models.ImeiBatch, error)
	ListBatches(ImeiBatchListFilter) ([]models.ImeiBatch, int64, error)
	SetBatchTotal(primitive.ObjectID, int) error
	AddBatchTotal(primitive.ObjectID, int) (*models.ImeiBatch, error)
	RegisterBatch(id primitive.ObjectID, at time.Time, userID primitive.ObjectID, userName string) error
	DeleteBatch(primitive.ObjectID) error

	FindExisting([]string) (map[string]models.ImeiRecord, error)
	InsertRecords([]models.ImeiRecord) (inserted int, duplicates []string, err error)
	GetRecord(imei string) (*models.ImeiRecord, error)
	ListBatchRecords(batchID primitive.ObjectID, search string, page, limit int) ([]models.ImeiRecord, int64, error)
	AllBatchRecords(primitive.ObjectID) ([]models.ImeiRecord, error)
	DeleteRecord(imei string) (bool, error)
	// DeleteDevice removes every READY IMEI of serial in the batch and returns
	// the removed IMEIs (a device is stored and deleted as a whole).
	DeleteDevice(batchID primitive.ObjectID, serial string) ([]string, error)
	BatchIDsWithIMEIPrefix(prefix string, max int) ([]primitive.ObjectID, error)
}

type imeiValidatorRepository struct{}

func NewImeiValidatorRepository() IImeiValidatorRepository {
	return &imeiValidatorRepository{}
}

var (
	imeiBatchesCollection       = database.GetInstance().Collection("imei_batches")
	imeiRecordsCollection       = database.GetInstance().Collection("imei_records")
	imeiBatchCountersCollection = database.GetInstance().Collection("imei_batch_counters")
)

// NextBatchCode returns LOT-YYYYMMDD-NNN with an atomic per-day counter, so
// two concurrent uploads never get the same code.
func (r *imeiValidatorRepository) NextBatchCode(day time.Time) (string, error) {
	key := day.UTC().Format("20060102")
	var doc struct {
		Seq int `bson:"seq"`
	}
	err := imeiBatchCountersCollection.FindOneAndUpdate(
		context.TODO(),
		bson.M{"_id": key},
		bson.M{"$inc": bson.M{"seq": 1}},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&doc)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("LOT-%s-%03d", key, doc.Seq), nil
}

func (r *imeiValidatorRepository) CreateBatch(b *models.ImeiBatch) error {
	if b.ID.IsZero() {
		b.ID = primitive.NewObjectID()
	}
	_, err := imeiBatchesCollection.InsertOne(context.TODO(), b)
	return err
}

func (r *imeiValidatorRepository) GetBatch(id primitive.ObjectID) (*models.ImeiBatch, error) {
	var b models.ImeiBatch
	err := imeiBatchesCollection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&b)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *imeiValidatorRepository) GetBatchesByIDs(ids []primitive.ObjectID) (map[primitive.ObjectID]models.ImeiBatch, error) {
	out := make(map[primitive.ObjectID]models.ImeiBatch, len(ids))
	for start := 0; start < len(ids); start += imeiValidatorChunk {
		end := min(start+imeiValidatorChunk, len(ids))
		cur, err := imeiBatchesCollection.Find(context.TODO(), bson.M{"_id": bson.M{"$in": ids[start:end]}})
		if err != nil {
			return nil, err
		}
		var batches []models.ImeiBatch
		if err := cur.All(context.TODO(), &batches); err != nil {
			return nil, err
		}
		for _, b := range batches {
			out[b.ID] = b
		}
	}
	return out, nil
}

func (r *imeiValidatorRepository) ListBatches(f ImeiBatchListFilter) ([]models.ImeiBatch, int64, error) {
	match := bson.M{}
	if f.Status != "" {
		match["status"] = f.Status
	}
	if f.CodeSearch != "" {
		or := bson.A{bson.M{"code": bson.M{"$regex": regexp.QuoteMeta(f.CodeSearch), "$options": "i"}}}
		if len(f.BatchIDs) > 0 {
			or = append(or, bson.M{"_id": bson.M{"$in": f.BatchIDs}})
		}
		match["$or"] = or
	}
	total, err := imeiBatchesCollection.CountDocuments(context.TODO(), match)
	if err != nil {
		return nil, 0, err
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}).
		SetSkip(int64((f.Page - 1) * f.Limit)).
		SetLimit(int64(f.Limit))
	cur, err := imeiBatchesCollection.Find(context.TODO(), match, opts)
	if err != nil {
		return nil, 0, err
	}
	items := []models.ImeiBatch{}
	if err := cur.All(context.TODO(), &items); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *imeiValidatorRepository) SetBatchTotal(id primitive.ObjectID, total int) error {
	_, err := imeiBatchesCollection.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{"$set": bson.M{"total": total}})
	return err
}

func (r *imeiValidatorRepository) AddBatchTotal(id primitive.ObjectID, delta int) (*models.ImeiBatch, error) {
	var b models.ImeiBatch
	err := imeiBatchesCollection.FindOneAndUpdate(
		context.TODO(),
		bson.M{"_id": id},
		bson.M{"$inc": bson.M{"total": delta}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&b)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *imeiValidatorRepository) RegisterBatch(id primitive.ObjectID, at time.Time, userID primitive.ObjectID, userName string) error {
	if _, err := imeiRecordsCollection.UpdateMany(context.TODO(),
		bson.M{"batch_id": id, "status": models.ImeiValidatorStatusReady},
		bson.M{"$set": bson.M{
			"status":        models.ImeiValidatorStatusRegistered,
			"registered_at": at,
			"registered_by": userID,
		}},
	); err != nil {
		return err
	}
	_, err := imeiBatchesCollection.UpdateOne(context.TODO(),
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"status":             models.ImeiValidatorStatusRegistered,
			"registered_at":      at,
			"registered_by":      userID,
			"registered_by_name": userName,
		}},
	)
	return err
}

func (r *imeiValidatorRepository) DeleteBatch(id primitive.ObjectID) error {
	if _, err := imeiRecordsCollection.DeleteMany(context.TODO(), bson.M{"batch_id": id}); err != nil {
		return err
	}
	_, err := imeiBatchesCollection.DeleteOne(context.TODO(), bson.M{"_id": id})
	return err
}

func (r *imeiValidatorRepository) FindExisting(imeis []string) (map[string]models.ImeiRecord, error) {
	out := make(map[string]models.ImeiRecord)
	opts := options.Find().SetProjection(bson.M{
		"imei": 1, "batch_id": 1, "status": 1, "created_at": 1, "registered_at": 1,
	})
	for start := 0; start < len(imeis); start += imeiValidatorChunk {
		end := min(start+imeiValidatorChunk, len(imeis))
		cur, err := imeiRecordsCollection.Find(context.TODO(), bson.M{"imei": bson.M{"$in": imeis[start:end]}}, opts)
		if err != nil {
			return nil, err
		}
		var rows []models.ImeiRecord
		if err := cur.All(context.TODO(), &rows); err != nil {
			return nil, err
		}
		for _, row := range rows {
			out[row.IMEI] = row
		}
	}
	return out, nil
}

// InsertRecords inserts unordered in chunks. An IMEI inserted concurrently by
// another upload hits the unique index; it is returned in duplicates instead
// of failing the whole upload.
func (r *imeiValidatorRepository) InsertRecords(records []models.ImeiRecord) (int, []string, error) {
	inserted := 0
	var duplicates []string
	for start := 0; start < len(records); start += imeiValidatorChunk {
		end := min(start+imeiValidatorChunk, len(records))
		chunk := records[start:end]
		docs := make([]interface{}, len(chunk))
		for i := range chunk {
			if chunk[i].ID.IsZero() {
				chunk[i].ID = primitive.NewObjectID()
			}
			docs[i] = chunk[i]
		}
		res, err := imeiRecordsCollection.InsertMany(context.TODO(), docs, options.InsertMany().SetOrdered(false))
		if res != nil {
			inserted += len(res.InsertedIDs)
		}
		if err != nil {
			var bwe mongo.BulkWriteException
			if !errors.As(err, &bwe) || bwe.WriteConcernError != nil {
				return inserted, duplicates, err
			}
			for _, we := range bwe.WriteErrors {
				if we.Code != 11000 { // duplicate key on the unique imei index
					return inserted, duplicates, err
				}
				duplicates = append(duplicates, chunk[we.Index].IMEI)
			}
			if res == nil {
				inserted += len(chunk) - len(bwe.WriteErrors)
			}
		}
	}
	return inserted, duplicates, nil
}

func (r *imeiValidatorRepository) GetRecord(imei string) (*models.ImeiRecord, error) {
	var rec models.ImeiRecord
	err := imeiRecordsCollection.FindOne(context.TODO(), bson.M{"imei": imei}).Decode(&rec)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *imeiValidatorRepository) ListBatchRecords(
	batchID primitive.ObjectID, search string, page, limit int,
) ([]models.ImeiRecord, int64, error) {
	match := bson.M{"batch_id": batchID}
	if search != "" {
		match["imei"] = bson.M{"$regex": "^" + regexp.QuoteMeta(search)}
	}
	total, err := imeiRecordsCollection.CountDocuments(context.TODO(), match)
	if err != nil {
		return nil, 0, err
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: 1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))
	cur, err := imeiRecordsCollection.Find(context.TODO(), match, opts)
	if err != nil {
		return nil, 0, err
	}
	items := []models.ImeiRecord{}
	if err := cur.All(context.TODO(), &items); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *imeiValidatorRepository) AllBatchRecords(batchID primitive.ObjectID) ([]models.ImeiRecord, error) {
	cur, err := imeiRecordsCollection.Find(context.TODO(), bson.M{"batch_id": batchID},
		options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return nil, err
	}
	items := []models.ImeiRecord{}
	if err := cur.All(context.TODO(), &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *imeiValidatorRepository) DeleteRecord(imei string) (bool, error) {
	res, err := imeiRecordsCollection.DeleteOne(context.TODO(), bson.M{"imei": imei})
	if err != nil {
		return false, err
	}
	return res.DeletedCount > 0, nil
}

func (r *imeiValidatorRepository) DeleteDevice(batchID primitive.ObjectID, serial string) ([]string, error) {
	filter := bson.M{"batch_id": batchID, "serial": serial, "status": models.ImeiValidatorStatusReady}
	cur, err := imeiRecordsCollection.Find(context.TODO(), filter, options.Find().SetProjection(bson.M{"imei": 1}))
	if err != nil {
		return nil, err
	}
	var rows []models.ImeiRecord
	if err := cur.All(context.TODO(), &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []string{}, nil
	}
	imeis := make([]string, len(rows))
	for i, row := range rows {
		imeis[i] = row.IMEI
	}
	if _, err := imeiRecordsCollection.DeleteMany(context.TODO(),
		bson.M{"batch_id": batchID, "imei": bson.M{"$in": imeis}, "status": models.ImeiValidatorStatusReady},
	); err != nil {
		return nil, err
	}
	return imeis, nil
}

// BatchIDsWithIMEIPrefix lists the batches holding an IMEI that starts with
// prefix (anchored regex, served by the imei index).
func (r *imeiValidatorRepository) BatchIDsWithIMEIPrefix(prefix string, max int) ([]primitive.ObjectID, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"imei": bson.M{"$regex": "^" + regexp.QuoteMeta(prefix)}}}},
		{{Key: "$group", Value: bson.M{"_id": "$batch_id"}}},
		{{Key: "$limit", Value: max}},
	}
	cur, err := imeiRecordsCollection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}
	var rows []struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	if err := cur.All(context.TODO(), &rows); err != nil {
		return nil, err
	}
	ids := make([]primitive.ObjectID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids, nil
}
