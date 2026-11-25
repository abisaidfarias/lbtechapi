package repositories

import (
	"context"
	"errors"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IDeviceCatalogRepository interface {
	List() ([]models.DeviceCatalogValue, error)
	ListByCategory(category string) ([]models.DeviceCatalogValue, error)
	Get(primitive.ObjectID) (*models.DeviceCatalogValue, error)
	FindByKey(category, valueKey string, exclude primitive.ObjectID) (*models.DeviceCatalogValue, error)
	Insert(*models.DeviceCatalogValue) error
	UpdateFields(primitive.ObjectID, primitive.M) error
	SetOrder(category string, ids []primitive.ObjectID) error
	Delete(primitive.ObjectID) error
	// UsedCounts maps each value of deviceField to the number of devices
	// using it.
	UsedCounts(deviceField string) (map[string]int, error)
}

type deviceCatalogRepository struct{}

func NewDeviceCatalogRepository() IDeviceCatalogRepository {
	return &deviceCatalogRepository{}
}

var deviceCatalogCollection = database.GetInstance().Collection("device_catalog_values")

var deviceCatalogSort = options.Find().SetSort(bson.D{
	{Key: "category", Value: 1}, {Key: "order", Value: 1}, {Key: "_id", Value: 1},
})

func (r *deviceCatalogRepository) List() ([]models.DeviceCatalogValue, error) {
	return r.find(bson.M{})
}

func (r *deviceCatalogRepository) ListByCategory(category string) ([]models.DeviceCatalogValue, error) {
	return r.find(bson.M{"category": category})
}

func (r *deviceCatalogRepository) find(filter bson.M) ([]models.DeviceCatalogValue, error) {
	cur, err := deviceCatalogCollection.Find(context.TODO(), filter, deviceCatalogSort)
	if err != nil {
		return nil, err
	}
	out := []models.DeviceCatalogValue{}
	if err := cur.All(context.TODO(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *deviceCatalogRepository) Get(id primitive.ObjectID) (*models.DeviceCatalogValue, error) {
	return r.findOne(bson.M{"_id": id})
}

func (r *deviceCatalogRepository) FindByKey(category, valueKey string, exclude primitive.ObjectID) (*models.DeviceCatalogValue, error) {
	filter := bson.M{"category": category, "value_key": valueKey}
	if !exclude.IsZero() {
		filter["_id"] = bson.M{"$ne": exclude}
	}
	return r.findOne(filter)
}

func (r *deviceCatalogRepository) findOne(filter bson.M) (*models.DeviceCatalogValue, error) {
	var v models.DeviceCatalogValue
	err := deviceCatalogCollection.FindOne(context.TODO(), filter).Decode(&v)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *deviceCatalogRepository) Insert(v *models.DeviceCatalogValue) error {
	if v.ID.IsZero() {
		v.ID = primitive.NewObjectID()
	}
	_, err := deviceCatalogCollection.InsertOne(context.TODO(), v)
	return err
}

func (r *deviceCatalogRepository) UpdateFields(id primitive.ObjectID, set primitive.M) error {
	res, err := deviceCatalogCollection.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{"$set": set})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// SetOrder writes order 0..n following ids (all from the same category).
func (r *deviceCatalogRepository) SetOrder(category string, ids []primitive.ObjectID) error {
	models := make([]mongo.WriteModel, 0, len(ids))
	for i, id := range ids {
		models = append(models, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": id, "category": category}).
			SetUpdate(bson.M{"$set": bson.M{"order": i}}))
	}
	if len(models) == 0 {
		return nil
	}
	_, err := deviceCatalogCollection.BulkWrite(context.TODO(), models)
	return err
}

func (r *deviceCatalogRepository) Delete(id primitive.ObjectID) error {
	_, err := deviceCatalogCollection.DeleteOne(context.TODO(), bson.M{"_id": id})
	return err
}

func (r *deviceCatalogRepository) UsedCounts(deviceField string) (map[string]int, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{deviceField: bson.M{"$type": "string", "$ne": ""}}}},
		{{Key: "$group", Value: bson.M{"_id": "$" + deviceField, "count": bson.M{"$sum": 1}}}},
	}
	cur, err := deviceCollection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}
	var rows []struct {
		Value string `bson:"_id"`
		Count int    `bson:"count"`
	}
	if err := cur.All(context.TODO(), &rows); err != nil {
		return nil, err
	}
	out := make(map[string]int, len(rows))
	for _, row := range rows {
		out[row.Value] = row.Count
	}
	return out, nil
}
