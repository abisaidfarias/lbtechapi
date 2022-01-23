package repositories

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/database/queries"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

type IDeviceTrackingRepository interface {
	Create(*models.DeviceTracking) error
	Get(isInternal bool, companyID primitive.ObjectID) ([]*responses.DeviceTracking, error)
	AddTrakingLog(trackingLog *models.TrackingLog, deviceTranckingID primitive.ObjectID) error
	GetByCompany(primitive.ObjectID) (*responses.DeviceTracking, error)
	GetByDevice(primitive.ObjectID) (*responses.DeviceTracking, error)
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
func (r *deviceTrackingRepository) Get(isInternal bool, companyID primitive.ObjectID) ([]*responses.DeviceTracking, error) {

	cursor, err := deviceTrackingCollection.Aggregate(context.TODO(), 
		queries.GetDeviceTrackingExpanded(isInternal,companyID))
	if err != nil {
		return nil, err
	}
	var deviceTrackings []*responses.DeviceTracking = []*responses.DeviceTracking{}
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