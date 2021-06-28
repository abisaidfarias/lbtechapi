package repositories

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/database/queries"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

type IPrinterRepository interface {
	Create(*models.Printer) error
	Get() ([]*responses.Printer, error)
}

type printerRepository struct {
}

func NewPrinterRepository() IPrinterRepository {
	return &printerRepository{}
}

var printerCollection = database.GetInstance().Collection("printers")

// Create a new tet case
func (r *printerRepository) Create(printer *models.Printer) error {

	var printerExist *models.Printer

	err := printerCollection.FindOne(context.TODO(),
		queries.GetPrinterBySerial(printer.Serial)).Decode(&printerExist)

	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			_, err := printerCollection.InsertOne(context.TODO(), printer)

			if err != nil {
				return err
			}
		default:
			return err
		}
	}else{
		filter, update := queries.UpdatePrinterRegistry(printerExist, printerExist.ID)

		_, err = deviceCollection.UpdateOne(context.TODO(), filter, update)

		if err != nil {
			return err
		}
	}
	return nil
}

// Get returns a list of all test cases
func (r *printerRepository) Get() ([]*responses.Printer, error) {

	cursor, err := printerCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}
	var countries []*responses.Printer = []*responses.Printer{}
	if err = cursor.All(context.TODO(), &countries); err != nil {
		panic(err)
	}
	cursor.Close(context.TODO())
	return countries, nil
}
