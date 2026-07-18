package repositories



import (

	"context"

	"fmt"
	"time"



	"github.com/abisaidfarias/lbtechapi/database"

	"github.com/abisaidfarias/lbtechapi/database/queries"

	"github.com/abisaidfarias/lbtechapi/models"

	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/mongo"

)



type IShipmentControlRepository interface {

	Create(*models.ShipmentControl) (string, error)

	GetByInternal([]primitive.ObjectID, []primitive.ObjectID) ([]*responses.ShipmentControlExpanded, error)

	GetByExternal(primitive.ObjectID, []primitive.ObjectID) ([]*responses.ShipmentControlExpanded, error)

	GetAvailableMultibandas(primitive.ObjectID, []primitive.ObjectID) ([]*responses.MultibandaExpanded, error)

	GetById(primitive.ObjectID) (*models.ShipmentControl, error)

	Update(string, *models.ShipmentControl) error

	ExistsByMultibandaExcludingID(primitive.ObjectID, primitive.ObjectID) (bool, error)

	PhaseChange(string, *models.ShipmentControl) error

	Delete(primitive.ObjectID) error

	SetRequestDelete(primitive.ObjectID, bool) error

	CountCertificatesByControlPrefix(string) (int64, error)

	UpdateCertificate(primitive.ObjectID, string, string) error

	ClaimCertificateGeneration(primitive.ObjectID, models.OabiCertificateState, time.Time) (bool, error)

	GetCertificateState(primitive.ObjectID) (*models.OabiCertificateState, error)

	MarkCertificateReady(primitive.ObjectID, string, string, time.Time) error

	MarkCertificateFailed(primitive.ObjectID, string) error

}



type shipmentControlRepository struct {

	multibandaRepository IMultibandaRepository

}



func NewShipmentControlRepository(multibandaRepository IMultibandaRepository) IShipmentControlRepository {

	return &shipmentControlRepository{

		multibandaRepository: multibandaRepository,

	}

}



var shipmentControlCollection = database.GetInstance().Collection("shipment_controls")



type shipmentControlListRow struct {

	responses.ShipmentControlExpanded `bson:",inline"`

	MultibandaID                      primitive.ObjectID `bson:"multibanda"`

}



func (r *shipmentControlRepository) Create(shipmentControl *models.ShipmentControl) (string, error) {

	res, err := shipmentControlCollection.InsertOne(context.TODO(), shipmentControl)

	if err != nil {

		return "", err

	}



	oid, ok := res.InsertedID.(primitive.ObjectID)

	if !ok {

		return "", fmt.Errorf("unexpected inserted id type")

	}



	shipmentControl.ID = oid

	return oid.Hex(), nil

}



func (r *shipmentControlRepository) GetByInternal(
	companies []primitive.ObjectID,
	brands []primitive.ObjectID,
) ([]*responses.ShipmentControlExpanded, error) {
	return r.list(queries.GetShipmentControls(companies, brands, true, primitive.NilObjectID))
}

func (r *shipmentControlRepository) GetByExternal(
	companyID primitive.ObjectID,
	brands []primitive.ObjectID,
) ([]*responses.ShipmentControlExpanded, error) {
	return r.list(queries.GetShipmentControls(nil, brands, false, companyID))
}



func (r *shipmentControlRepository) list(pipeline mongo.Pipeline) ([]*responses.ShipmentControlExpanded, error) {

	cursor, err := shipmentControlCollection.Aggregate(context.TODO(), pipeline)

	if err != nil {

		return nil, err

	}



	items := []*responses.ShipmentControlExpanded{}

	for cursor.Next(context.TODO()) {

		var row shipmentControlListRow

		if err := cursor.Decode(&row); err != nil {

			return nil, err

		}



		item := row.ShipmentControlExpanded

		if !row.MultibandaID.IsZero() {

			multibanda, err := r.multibandaRepository.GetByIdExpanded(row.MultibandaID)

			if err != nil {

				return nil, err

			}

			if multibanda != nil {

				item.Multibanda = mapping.ToShipmentControlMultibandaSummary(*multibanda)
				item.Device = mapping.ToShipmentControlDeviceSummary(*multibanda)
				mapping.EnrichShipmentControlSubtelCertificate(&item, multibanda)

			}

		}

		mapping.SanitizeShipmentControlExpandedDates(&item)

		items = append(items, &item)

	}



	return items, nil

}



func (r *shipmentControlRepository) GetAvailableMultibandas(

	companyID primitive.ObjectID,

	brands []primitive.ObjectID,

) ([]*responses.MultibandaExpanded, error) {

	cursor, err := database.GetInstance().

		Collection("multibandas").

		Aggregate(context.TODO(), queries.GetAvailableMultibandas(companyID, brands))

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



func (r *shipmentControlRepository) GetById(id primitive.ObjectID) (*models.ShipmentControl, error) {

	var shipmentControl models.ShipmentControl

	err := shipmentControlCollection.FindOne(context.TODO(), queries.GetShipmentControlById(id)).Decode(&shipmentControl)

	if err == mongo.ErrNoDocuments {

		return nil, nil

	}

	if err != nil {

		return nil, err

	}

	return &shipmentControl, nil

}



func (r *shipmentControlRepository) Update(id string, shipmentControl *models.ShipmentControl) error {

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {

		return err

	}

	filter, update := queries.UpdateShipmentControl(shipmentControl, oid)

	res, err := shipmentControlCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {

		return err

	}

	if res.MatchedCount == 0 {

		return mongo.ErrNoDocuments

	}

	return nil

}



func (r *shipmentControlRepository) ExistsByMultibandaExcludingID(

	excludeID primitive.ObjectID,

	multibandaID primitive.ObjectID,

) (bool, error) {

	err := shipmentControlCollection.FindOne(

		context.TODO(),

		queries.GetShipmentControlByMultibandaExcludingID(excludeID, multibandaID),

	).Err()

	if err == nil {

		return true, nil

	}

	if err == mongo.ErrNoDocuments {

		return false, nil

	}

	return false, err

}



func (r *shipmentControlRepository) PhaseChange(id string, shipmentControl *models.ShipmentControl) error {

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {

		return err

	}



	filter, update := queries.UpdateShipmentControlPhaseChange(shipmentControl, oid)

	_, err = shipmentControlCollection.UpdateOne(context.TODO(), filter, update)

	return err

}

func (r *shipmentControlRepository) Delete(id primitive.ObjectID) error {
	res, err := shipmentControlCollection.DeleteOne(context.TODO(), queries.GetShipmentControlById(id))
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *shipmentControlRepository) SetRequestDelete(id primitive.ObjectID, value bool) error {
	filter, update := queries.SetShipmentControlRequestDelete(id, value)
	res, err := shipmentControlCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *shipmentControlRepository) CountCertificatesByControlPrefix(prefix string) (int64, error) {
	return shipmentControlCollection.CountDocuments(context.TODO(), queries.CountShipmentControlCertificatesByPrefix(prefix))
}

func (r *shipmentControlRepository) UpdateCertificate(id primitive.ObjectID, certificateURL, registroOABI string) error {
	filter, update := queries.UpdateShipmentControlCertificate(id, certificateURL, registroOABI)
	_, err := shipmentControlCollection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (r *shipmentControlRepository) ClaimCertificateGeneration(
	id primitive.ObjectID,
	state models.OabiCertificateState,
	staleBefore time.Time,
) (bool, error) {
	filter, update := queries.ClaimOabiCertificateGeneration(id, staleBefore, state)
	res, err := shipmentControlCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return false, err
	}
	return res.MatchedCount > 0, nil
}

type shipmentControlCertificateStateDoc struct {
	OabiCertificateState *models.OabiCertificateState `bson:"oabi_certificate_state"`
	OabiCertificateUrl   string                       `bson:"oabi_certificate_url"`
}

func (r *shipmentControlRepository) GetCertificateState(id primitive.ObjectID) (*models.OabiCertificateState, error) {
	var doc shipmentControlCertificateStateDoc
	err := shipmentControlCollection.FindOne(
		context.TODO(),
		queries.GetShipmentControlById(id),
	).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return doc.OabiCertificateState, nil
}

func (r *shipmentControlRepository) MarkCertificateReady(
	id primitive.ObjectID,
	certificateURL, registroOABI string,
	generatedAt time.Time,
) error {
	filter := queries.MarkOabiCertificateReadyFilter(id)
	update := queries.MarkOabiCertificateReadyUpdate(certificateURL, registroOABI, generatedAt)
	res, err := shipmentControlCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *shipmentControlRepository) MarkCertificateFailed(id primitive.ObjectID, errorMessage string) error {
	filter := queries.MarkOabiCertificateFailedFilter(id)
	update := queries.MarkOabiCertificateFailedUpdate(errorMessage)
	res, err := shipmentControlCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

