package repositories

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/database/queries"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

type IDeviceRepository interface {
	Create(*models.Device) (*responses.DeviceExpanded, error)
	Get(brands []primitive.ObjectID) ([]*responses.DeviceExpanded, error)
	GetById(string) (*responses.Device, error)
	FindByTechnicalModel(string) ([]*responses.Device, error)
	FindByCommercialModel(string) ([]*responses.Device, error)
	ExistsByTechnicalModelExcludingID(primitive.ObjectID, string) (bool, error)
	ExistsByCommercialModelExcludingID(primitive.ObjectID, string) (bool, error)
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
func (r *deviceRepository) Create(device *models.Device) (*responses.DeviceExpanded, error) {

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
func (r *deviceRepository) Get(brands []primitive.ObjectID) ([]*responses.DeviceExpanded, error) {

	cursor, err := deviceCollection.Aggregate(context.TODO(), queries.GetDevicesExpanded(brands))

	if err != nil {
		return nil, err
	}
	var devices []*responses.DeviceExpanded = []*responses.DeviceExpanded{}
	for cursor.Next(context.TODO()) {
		var device responses.DeviceExpanded
		err := cursor.Decode(&device)
		if err != nil {
			return nil, err
		}
		devices = append(devices, &device)
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

	err = deviceCollection.FindOne(context.TODO(),
		queries.GetDeviceById(oid)).Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *deviceRepository) FindByTechnicalModel(technicalModel string) ([]*responses.Device, error) {
	cursor, err := deviceCollection.Find(context.TODO(), queries.GetDevicesByTechnicalModel(technicalModel))
	if err != nil {
		return nil, err
	}

	devices := []*responses.Device{}
	for cursor.Next(context.TODO()) {
		var device responses.Device
		if err := cursor.Decode(&device); err != nil {
			return nil, err
		}
		devices = append(devices, &device)
	}
	cursor.Close(context.TODO())

	return devices, nil
}

func (r *deviceRepository) FindByCommercialModel(commercialModel string) ([]*responses.Device, error) {
	cursor, err := deviceCollection.Find(context.TODO(), queries.GetDevicesByCommercialModel(commercialModel))
	if err != nil {
		return nil, err
	}

	devices := []*responses.Device{}
	for cursor.Next(context.TODO()) {
		var device responses.Device
		if err := cursor.Decode(&device); err != nil {
			return nil, err
		}
		devices = append(devices, &device)
	}
	cursor.Close(context.TODO())

	return devices, nil
}

func (r *deviceRepository) ExistsByTechnicalModelExcludingID(excludeID primitive.ObjectID, technicalModel string) (bool, error) {
	err := deviceCollection.FindOne(
		context.TODO(),
		queries.GetDeviceByTechnicalModelExcludingID(excludeID, technicalModel),
	).Err()
	if err == nil {
		return true, nil
	}
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	return false, err
}

func (r *deviceRepository) ExistsByCommercialModelExcludingID(excludeID primitive.ObjectID, commercialModel string) (bool, error) {
	err := deviceCollection.FindOne(
		context.TODO(),
		queries.GetDeviceByCommercialModelExcludingID(excludeID, commercialModel),
	).Err()
	if err == nil {
		return true, nil
	}
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	return false, err
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
