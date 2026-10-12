package repositories

import (
	"context"
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ImeiValidationListFilter struct {
	IMEI           string
	Brand          string
	TechnicalModel string
	CommercialName string
	Source         string
	DateFrom       *time.Time
	DateTo         *time.Time
	Page           int
	PageSize       int
}

type IImeiValidationRepository interface {
	GetTacByID(string) (*models.TacCatalog, error)
	UpsertTac(*models.TacCatalog) error
	InsertQuery(*models.ImeiQuery) error
	ListQueries(ImeiValidationListFilter) ([]models.ImeiQuery, int64, error)
	InsertImeiInfoUsage(userID primitive.ObjectID, imei, tac string, at time.Time) error
	CountImeiInfoUsageSince(userID *primitive.ObjectID, since time.Time) (int64, error)
	HasImeiInfoUsageSince(userID primitive.ObjectID, imei string, since time.Time) (bool, error)
}

type imeiValidationRepository struct{}

func NewImeiValidationRepository() IImeiValidationRepository {
	return &imeiValidationRepository{}
}

var (
	tacCatalogCollection    = database.GetInstance().Collection("tac_catalog")
	imeiQueriesCollection   = database.GetInstance().Collection("imei_queries")
	imeiInfoUsageCollection = database.GetInstance().Collection("imei_info_usage")
)

func (r *imeiValidationRepository) GetTacByID(tac string) (*models.TacCatalog, error) {
	var row models.TacCatalog
	err := tacCatalogCollection.FindOne(context.TODO(), bson.M{"_id": tac}).Decode(&row)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *imeiValidationRepository) UpsertTac(row *models.TacCatalog) error {
	filter := bson.M{"_id": row.TAC}
	update := bson.M{
		"$set": bson.M{
			"brand":            row.Brand,
			"technical_model":  row.TechnicalModel,
			"commercial_name":  row.CommercialName,
			"source_imei":      row.SourceIMEI,
			"raw_response":     row.RawResponse,
			"created_at":       row.CreatedAt,
		},
	}
	_, err := tacCatalogCollection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))
	return err
}

func (r *imeiValidationRepository) InsertQuery(row *models.ImeiQuery) error {
	if row.ID.IsZero() {
		row.ID = primitive.NewObjectID()
	}
	_, err := imeiQueriesCollection.InsertOne(context.TODO(), row)
	return err
}

func (r *imeiValidationRepository) ListQueries(filter ImeiValidationListFilter) ([]models.ImeiQuery, int64, error) {
	match := bson.M{}
	if v := strings.TrimSpace(filter.IMEI); v != "" {
		match["imei"] = v
	}
	if v := strings.TrimSpace(filter.Brand); v != "" {
		match["brand"] = bson.M{"$regex": v, "$options": "i"}
	}
	if v := strings.TrimSpace(filter.TechnicalModel); v != "" {
		match["technical_model"] = bson.M{"$regex": v, "$options": "i"}
	}
	if v := strings.TrimSpace(filter.CommercialName); v != "" {
		match["commercial_name"] = bson.M{"$regex": v, "$options": "i"}
	}
	if v := strings.TrimSpace(filter.Source); v != "" {
		match["source"] = v
	}
	if filter.DateFrom != nil || filter.DateTo != nil {
		dateFilter := bson.M{}
		if filter.DateFrom != nil {
			dateFilter["$gte"] = *filter.DateFrom
		}
		if filter.DateTo != nil {
			dateFilter["$lte"] = *filter.DateTo
		}
		match["queried_at"] = dateFilter
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 25
	}
	if pageSize > 200 {
		pageSize = 200
	}

	total, err := imeiQueriesCollection.CountDocuments(context.TODO(), match)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "queried_at", Value: -1}}).
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize))

	cursor, err := imeiQueriesCollection.Find(context.TODO(), match, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(context.TODO())

	rows := []models.ImeiQuery{}
	for cursor.Next(context.TODO()) {
		var row models.ImeiQuery
		if err := cursor.Decode(&row); err != nil {
			return nil, 0, err
		}
		rows = append(rows, row)
	}
	return rows, total, cursor.Err()
}

func (r *imeiValidationRepository) InsertImeiInfoUsage(
	userID primitive.ObjectID,
	imei, tac string,
	at time.Time,
) error {
	_, err := imeiInfoUsageCollection.InsertOne(context.TODO(), models.ImeiInfoUsage{
		ID:     primitive.NewObjectID(),
		UserID: userID,
		IMEI:   imei,
		TAC:    tac,
		At:     at.UTC(),
	})
	return err
}

func (r *imeiValidationRepository) CountImeiInfoUsageSince(
	userID *primitive.ObjectID,
	since time.Time,
) (int64, error) {
	match := bson.M{"at": bson.M{"$gte": since.UTC()}}
	if userID != nil {
		match["user_id"] = *userID
	}
	return imeiInfoUsageCollection.CountDocuments(context.TODO(), match)
}

func (r *imeiValidationRepository) HasImeiInfoUsageSince(
	userID primitive.ObjectID,
	imei string,
	since time.Time,
) (bool, error) {
	n, err := imeiInfoUsageCollection.CountDocuments(context.TODO(), bson.M{
		"user_id": userID,
		"imei":    imei,
		"at":      bson.M{"$gte": since.UTC()},
	})
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
