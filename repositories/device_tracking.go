package repositories

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/database/queries"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

type IDeviceTrackingRepository interface {
	Create(*models.DeviceTracking) error
	Get(isInternal bool, companyID primitive.ObjectID) ([]*responses.DeviceTracking, error)
	AddTrakingLog(trackingLog *models.TrackingLog, deviceTranckingID primitive.ObjectID) error
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

	if isInternal {
		cursor, err := deviceTrackingCollection.Find(context.TODO(), bson.M{})
		if err != nil {
			panic(err)
		}
		var deviceTrackings []*responses.DeviceTracking = []*responses.DeviceTracking{}
		if err = cursor.All(context.TODO(), &deviceTrackings); err != nil {
			panic(err)
		}
		cursor.Close(context.TODO())
		return deviceTrackings, nil
	}
	cursor, err := deviceCollection.Aggregate(context.TODO(), queries.GetDeviceTracking(companyID))
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
