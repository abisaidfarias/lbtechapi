package repositories

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/database/queries"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

type IMultibandaRepository interface {
	Create(*models.Multibanda) (string, error)
	GetByInternal([]primitive.ObjectID, []primitive.ObjectID) ([]*responses.MultibandaExpanded, error)
	GetByExternal(primitive.ObjectID, []primitive.ObjectID) ([]*responses.MultibandaExpanded, error)
	PhaseChange(string, *models.Multibanda) error
	GetByIdExpanded(primitive.ObjectID) (*responses.MultibandaExpanded, error)
	ExistsByCompanyDeviceSoftwareOsVersion(primitive.ObjectID, primitive.ObjectID, string, string) (bool, error)
	Delete(primitive.ObjectID) error
	SetRequestDelete(primitive.ObjectID, bool) error
}

type multibandaRepository struct {
}

func NewMultibandaRepository() IMultibandaRepository {
	return &multibandaRepository{}
}

var multibandaCollection = database.GetInstance().Collection("multibandas")

func (r *multibandaRepository) Create(multibanda *models.Multibanda) (string, error) {
	res, err := multibandaCollection.InsertOne(context.TODO(), multibanda)
	if err != nil {
		return "", err
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("unexpected inserted id type")
	}

	multibanda.ID = oid
	return oid.Hex(), nil
}

func (r *multibandaRepository) GetByInternal(
	companies []primitive.ObjectID,
	brands []primitive.ObjectID,
) ([]*responses.MultibandaExpanded, error) {
	return r.list(queries.GetMultibandas(companies, brands, true, primitive.NilObjectID))
}

func (r *multibandaRepository) GetByExternal(
	companyID primitive.ObjectID,
	brands []primitive.ObjectID,
) ([]*responses.MultibandaExpanded, error) {
	return r.list(queries.GetMultibandas([]primitive.ObjectID{}, brands, false, companyID))
}

func (r *multibandaRepository) list(pipeline interface{}) ([]*responses.MultibandaExpanded, error) {
	cursor, err := multibandaCollection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	multibandas := []*responses.MultibandaExpanded{}
	for cursor.Next(context.TODO()) {
		var multibanda responses.MultibandaExpanded
		if err := cursor.Decode(&multibanda); err != nil {
			return nil, err
		}
		functions.EnrichMultibandaExpanded(&multibanda)
		multibandas = append(multibandas, &multibanda)
	}

	return multibandas, nil
}

func (r *multibandaRepository) PhaseChange(id string, multibanda *models.Multibanda) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter, update := queries.UpdateMultibandaPhaseChange(multibanda, oid)

	_, err = multibandaCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (r *multibandaRepository) GetByIdExpanded(multibandaID primitive.ObjectID) (*responses.MultibandaExpanded, error) {
	cursor, err := multibandaCollection.Aggregate(context.TODO(), queries.GetMultibandaExpandedById(multibandaID))
	if err != nil {
		return nil, err
	}

	for cursor.Next(context.TODO()) {
		var multibanda responses.MultibandaExpanded
		if err := cursor.Decode(&multibanda); err != nil {
			return nil, err
		}
		functions.EnrichMultibandaExpanded(&multibanda)
		return &multibanda, nil
	}

	return nil, nil
}

func (r *multibandaRepository) ExistsByCompanyDeviceSoftwareOsVersion(
	companyID primitive.ObjectID,
	deviceID primitive.ObjectID,
	softwareVersion string,
	osVersion string,
) (bool, error) {
	err := multibandaCollection.FindOne(
		context.TODO(),
		queries.GetMultibandaByCompanyDeviceSoftwareOsVersion(companyID, deviceID, softwareVersion, osVersion),
	).Err()
	if err == nil {
		return true, nil
	}
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	return false, err
}

func (r *multibandaRepository) Delete(id primitive.ObjectID) error {
	res, err := multibandaCollection.DeleteOne(context.TODO(), queries.GetMultibandaById(id))
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *multibandaRepository) SetRequestDelete(id primitive.ObjectID, value bool) error {
	filter, update := queries.SetMultibandaRequestDelete(id, value)
	res, err := multibandaCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
