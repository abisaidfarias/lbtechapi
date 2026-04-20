package repositories

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/database/queries"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

type IDeviceTrackingRepository interface {
	Create(*models.DeviceTracking) error
	Get(isInternal bool, companyID primitive.ObjectID,
		brands []primitive.ObjectID, countries []string,
		companies []primitive.ObjectID) ([]*responses.DeviceTrackingExpanded, error)
	AddTrakingLog(trackingLog *models.TrackingLog, deviceTranckingID primitive.ObjectID) error
	GetByCompany(primitive.ObjectID) (*responses.DeviceTracking, error)
	GetByDevice(primitive.ObjectID) (*responses.DeviceTracking, error)
	Delete([]primitive.ObjectID) error
	Update(string, *models.DeviceTracking) error
	AdvancedSearch(*request.SearchOption, primitive.ObjectID, bool, []primitive.ObjectID, []string,
		[]primitive.ObjectID) ([]*responses.DeviceTrackingExpanded, error)
	GetByPerson(primitive.ObjectID) (*responses.DeviceTracking, error)
	GetById(string) (*responses.DeviceTracking, error)
	SetTrackingLogDocumentURLByTrackingID(deviceIDs []primitive.ObjectID, trackingID string, documentURL string) error
}

type deviceTrackingRepository struct {
}

func NewDeviceTrackingRepository() IDeviceTrackingRepository {
	return &deviceTrackingRepository{}
}

var deviceTrackingCollection = database.GetInstance().Collection("device_trackings")

// Create a new tet case
func (r *deviceTrackingRepository) Create(deviceTracking *models.DeviceTracking) error {

	_, err := deviceTrackingCollection.InsertOne(context.TODO(), deviceTracking)

	if err != nil {
		return err
	}
	return nil
}

// Get returns a list of all test cases
func (r *deviceTrackingRepository) Get(isInternal bool, companyID primitive.ObjectID,
	brands []primitive.ObjectID, countries []string,
	companies []primitive.ObjectID) ([]*responses.DeviceTrackingExpanded, error) {
	deviceTrackingQuery := queries.GetDeviceTrackingExpanded(isInternal, companyID, brands, countries, companies)
	cursor, err := deviceTrackingCollection.Aggregate(context.TODO(), deviceTrackingQuery)
	if err != nil {
		return nil, err
	}
	var deviceTrackings []*responses.DeviceTrackingExpanded = []*responses.DeviceTrackingExpanded{}
	if err = cursor.All(context.TODO(), &deviceTrackings); err != nil {
		panic(err)
	}
	cursor.Close(context.TODO())
	return deviceTrackings, nil

}
func (r *deviceTrackingRepository) AddTrakingLog(trackingLog *models.TrackingLog,
	deviceTranckingID primitive.ObjectID) error {

	filter, update := queries.AddTrackingLog(trackingLog, deviceTranckingID)
	_, err := deviceTrackingCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}
	return nil
}
func (r *deviceTrackingRepository) GetByCompany(companyId primitive.ObjectID) (*responses.DeviceTracking, error) {

	var deviceTracking *responses.DeviceTracking
	err := deviceTrackingCollection.FindOne(context.TODO(),
		queries.GetDeviceTrackingByCompany(companyId)).Decode(&deviceTracking)

	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return nil, nil
		default:
			return nil, err
		}
	}
	return deviceTracking, nil
}

func (r *deviceTrackingRepository) GetByDevice(deviceId primitive.ObjectID) (*responses.DeviceTracking, error) {

	var deviceTracking *responses.DeviceTracking
	err := deviceTrackingCollection.FindOne(context.TODO(),
		queries.GetDeviceTrackingByDevice(deviceId)).Decode(&deviceTracking)

	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return nil, nil
		default:
			return nil, err
		}
	}
	return deviceTracking, nil
}
func (r *deviceTrackingRepository) Delete(deviceTrackingsId []primitive.ObjectID) error {

	_, err := deviceTrackingCollection.DeleteMany(context.TODO(),
		queries.DeleteDeviceTracking(deviceTrackingsId))

	if err != nil {
		return err
	}

	return nil
}
func (r *deviceTrackingRepository) Update(id string, deviceTracking *models.DeviceTracking) error {

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter, update := queries.UpdateDeviceTracking(deviceTracking, oid)

	_, err = deviceTrackingCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	return nil
}
func (r *deviceTrackingRepository) AdvancedSearch(searchOption *request.SearchOption,
	companyId primitive.ObjectID, isInternal bool,
	brands []primitive.ObjectID, countries []string,
	companies []primitive.ObjectID) ([]*responses.DeviceTrackingExpanded, error) {

	cursor, err := deviceTrackingCollection.Aggregate(context.TODO(),
		queries.GetDeviceTrackingAdvancedSearch(companyId, isInternal,
			searchOption, brands, countries, companies))
	if err != nil {
		return nil, err
	}
	var deviceTrackings []*responses.DeviceTrackingExpanded = []*responses.DeviceTrackingExpanded{}
	if err = cursor.All(context.TODO(), &deviceTrackings); err != nil {
		panic(err)
	}
	cursor.Close(context.TODO())
	return deviceTrackings, nil
}
func (r *deviceTrackingRepository) GetByPerson(personId primitive.ObjectID) (*responses.DeviceTracking, error) {

	var deviceTracking *responses.DeviceTracking
	err := deviceTrackingCollection.FindOne(context.TODO(),
		queries.GetDeviceTrackingByPerson(personId)).Decode(&deviceTracking)

	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return nil, nil
		default:
			return nil, err
		}
	}
	return deviceTracking, nil
}
func (r *deviceTrackingRepository) GetById(id string) (*responses.DeviceTracking, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}

	var result responses.DeviceTracking

	err = deviceTrackingCollection.FindOne(context.TODO(),
		queries.GetDeviceTrackingById(oid)).Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *deviceTrackingRepository) SetTrackingLogDocumentURLByTrackingID(deviceIDs []primitive.ObjectID,
	trackingID string, documentURL string) error {
	if len(deviceIDs) == 0 || trackingID == "" {
		return nil
	}
	filter := bson.M{"_id": bson.M{"$in": deviceIDs}}
	update := bson.M{"$set": bson.M{"tracking_logs.$[log].document_url": documentURL}}
	opts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{bson.M{"log.tracking_id": trackingID}},
	})
	_, err := deviceTrackingCollection.UpdateMany(context.TODO(), filter, update, opts)
	return err
}
