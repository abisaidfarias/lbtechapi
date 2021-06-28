package queries

import (
	"github.com/abisaidfarias/lbtechapi/models"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

func UpdatePrinterRegistry(detail *models.Detail, oid primitive.ObjectID) (primitive.M, primitive.M) {

	filter := primitive.M{
		"_id": oid,
	}
	update := primitive.M{
		"$push": primitive.M{
			"details": detail,
		},
	}
	return filter, update
}
func GetPrinterBySerial(serial string) primitive.M {
	return primitive.M{"serial": serial}
}
func UpdatePrinter(printer *models.Printer, oid primitive.ObjectID) (primitive.M, primitive.M) {

	filter := primitive.M{
		"_id": oid,
	}
	update := primitive.M{
		"$set": printer,
	}
	return filter, update
}
