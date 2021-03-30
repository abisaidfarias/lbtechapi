package repositories

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/database/queries"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

type IDeviceRepository interface {
	Create(*models.Device) (*responses.Device, error)
	Get() ([]*responses.Device, error)
	GetById(string) (*responses.Device, error)
	Update(string, *models.Device) error
	Delete(primitive.ObjectID) error
}

type deviceRepository struct {
}

func NewDeviceRepository() IDeviceRepository {
	return &deviceRepository{}
}

var deviceCollection = database.GetInstance().Collection("devices")

// Create a new tet case
func (r *deviceRepository) Create(device *models.Device) (*responses.Device, error) {

	res, err := deviceCollection.InsertOne(context.TODO(), device)
	if err != nil {
		return nil, err
	}
	id := res.InsertedID.(primitive.ObjectID)
	device.ID = id
	deviceResponses := mapping.DeviceToDeviceResponse(device)
	return deviceResponses, nil
}

// Get returns a list of all test cases
func (r *deviceRepository) Get() ([]*responses.Device, error) {

	cursor, err := deviceCollection.Find(context.TODO(), bson.M{})

	if err != nil {

		panic(err)
	}
	var devices []*responses.Device = []*responses.Device{}
	if err = cursor.All(context.TODO(), &devices); err != nil {
		panic(err)
	}
	cursor.Close(context.TODO())
	return devices, nil
}

func (r *deviceRepository) GetById(id string) (*responses.Device, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}

	var result responses.Device

	err = deviceCollection.FindOne(context.TODO(), queries.GetDeviceById(oid)).Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}
func (r *deviceRepository) Update(id string, device *models.Device) error {

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter, update := queries.UpdateDevice(device, oid)

	_, err = deviceCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	return nil
}
func (r *deviceRepository) Delete(oid primitive.ObjectID) error {

	_, err := deviceCollection.DeleteOne(context.TODO(), queries.DeleteDevice(oid))

	if err != nil {
		return err
	}

	return nil
}
