package queries

import (
	"github.com/abisaidfarias/lbtechapi/models"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

func UpdatePrinterRegistry(printer *models.Printer, oid primitive.ObjectID) (primitive.M, primitive.M) {

	filter := primitive.M{
		"_id": oid,
	}
	update := primitive.M{
		"$push": primitive.M{
			"details": printer,
		},
	}
	return filter, update
}
func GetPrinterBySerial(serial string) primitive.M {
	return primitive.M{"serial": serial}
}
