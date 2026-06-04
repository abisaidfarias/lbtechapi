package queries

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
)

func GetShipmentControlById(oid primitive.ObjectID) primitive.M {
	return primitive.M{"_id": oid}
}

func SetShipmentControlRequestDelete(oid primitive.ObjectID, value bool) (primitive.M, primitive.D) {
	filter := primitive.M{"_id": oid}
	if !value {
		filter["request_delete"] = true
	}
	return filter, primitive.D{
		{Key: "$set", Value: primitive.D{
			{Key: "request_delete", Value: value},
		}},
	}
}

func GetAvailableMultibandas(companyID primitive.ObjectID, brands []primitive.ObjectID) mongo.Pipeline {
	preMatch := bson.D{
		{Key: "company", Value: companyID},
		{Key: "status", Value: enums.HomologationStatus_value["APPROVED"]},
		{Key: "request_delete", Value: bson.M{"$ne": true}},
	}
	if len(brands) > 0 {
		preMatch = append(preMatch, primitive.E{
			Key:   "brand",
			Value: bson.D{primitive.E{Key: "$in", Value: brands}},
		})
	}

	lookupStageDevice := bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "devices"},
			{Key: "localField", Value: "device"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "device"},
		}},
	}
	unwindStageDevice := bson.D{
		{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$device"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}},
	}
	lookupStageCompany := bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "companies"},
			{Key: "localField", Value: "company"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "company"},
		}},
	}
	unwindStageCompany := bson.D{
		{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$company"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}},
	}
	lookupStageBrand := bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "brands"},
			{Key: "localField", Value: "brand"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "brand"},
		}},
	}
	unwindStageBrand := bson.D{
		{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$brand"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}},
	}
	sort := bson.D{
		{Key: "$sort", Value: bson.D{
			{Key: "device.commercial_model", Value: 1},
			{Key: "software_version", Value: 1},
		}},
	}

	return mongo.Pipeline{
		bson.D{{Key: "$match", Value: preMatch}},
		lookupStageDevice, unwindStageDevice,
		lookupStageCompany, unwindStageCompany,
		lookupStageBrand, unwindStageBrand,
		sort,
	}
}

func GetShipmentControls(
	companies []primitive.ObjectID,
	brands []primitive.ObjectID,
	isInternal bool,
	companyID primitive.ObjectID,
) mongo.Pipeline {
	var preMatch bson.D
	var hasPreMatch bool

	if len(companies) > 0 && isInternal {
		preMatch = append(preMatch, primitive.E{
			Key:   "company",
			Value: bson.D{primitive.E{Key: "$in", Value: companies}},
		})
		hasPreMatch = true
	}
	if !isInternal {
		preMatch = append(preMatch, primitive.E{Key: "company", Value: companyID})
		hasPreMatch = true
	}
	if !isInternal {
		preMatch = append(preMatch, primitive.E{
			Key:   "request_delete",
			Value: bson.M{"$ne": true},
		})
		hasPreMatch = true
	}

	pipeline := mongo.Pipeline{}
	if hasPreMatch {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: preMatch}})
	}

	if len(brands) > 0 {
		lookupMultibanda := bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "multibandas"},
			{Key: "localField", Value: "multibanda"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "multibanda_ref"},
		}}}
		unwindMultibanda := bson.D{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$multibanda_ref"},
		}}}
		matchBrand := bson.D{{Key: "$match", Value: bson.D{
			{Key: "multibanda_ref.brand", Value: bson.D{{Key: "$in", Value: brands}}},
		}}}
		projectMultibandaRef := bson.D{{Key: "$project", Value: bson.D{
			{Key: "multibanda_ref", Value: 0},
		}}}
		pipeline = append(pipeline, lookupMultibanda, unwindMultibanda, matchBrand, projectMultibandaRef)
	}

	lookupStageCompany := bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "companies"},
			{Key: "localField", Value: "company"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "company"},
		}},
	}
	unwindStageCompany := bson.D{
		{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$company"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}},
	}
	lookupStageCountry := bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "countries"},
			{Key: "localField", Value: "country"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "country"},
		}},
	}
	unwindStageCountry := bson.D{
		{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$country"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}},
	}
	sort := bson.D{
		{Key: "$sort", Value: bson.D{
			{Key: "created_date", Value: -1},
			{Key: "planning_date", Value: -1},
		}},
	}

	pipeline = append(pipeline,
		lookupStageCompany, unwindStageCompany,
		lookupStageCountry, unwindStageCountry,
		sort,
	)
	return pipeline
}

func UpdateShipmentControlPhaseChange(shipmentControl *models.ShipmentControl, oid primitive.ObjectID) (primitive.M, primitive.D) {
	filter := primitive.M{"_id": oid}
	setFields := primitive.D{
		{Key: "current_phase", Value: shipmentControl.CurrentPhase},
		{Key: "status", Value: shipmentControl.Status},
		{Key: "imei_quantity", Value: shipmentControl.ImeiQuantity},
		{Key: "imei_file_url", Value: shipmentControl.ImeiFileUrl},
		{Key: "registered_imei_count", Value: shipmentControl.RegisteredImeiCount},
		{Key: "rework_number", Value: shipmentControl.ReworkNumber},
		{Key: "oabi_certificate", Value: shipmentControl.OabiCertificate},
		{Key: "client", Value: shipmentControl.Client},
		{Key: "subtel_certificate_url", Value: shipmentControl.SubtelCertificateUrl},
		{Key: "subtel_certificate_number", Value: shipmentControl.SubtelCertificateNumber},
		{Key: "oabi_certificate_url", Value: shipmentControl.OabiCertificateUrl},
		{Key: "oabi_certificate_number", Value: shipmentControl.OabiCertificateNumber},
		{Key: "comment", Value: shipmentControl.Comment},
	}
	appendShipmentControlDateField(&setFields, "validation_start_date", shipmentControl.ValidationStartDate)
	appendShipmentControlDateField(&setFields, "validation_end_date", shipmentControl.ValidationEndDate)
	appendShipmentControlDateField(&setFields, "under_revision_start_date", shipmentControl.UnderRevisionStartDate)
	appendShipmentControlDateField(&setFields, "under_revision_end_date", shipmentControl.UnderRevisionEndDate)
	appendShipmentControlDateField(&setFields, "completed_date", shipmentControl.CompletedDate)
	if !shipmentControl.Country.IsZero() {
		setFields = append(setFields, primitive.E{Key: "country", Value: shipmentControl.Country})
	}
	update := primitive.D{
		{Key: "$set", Value: setFields},
	}
	return filter, update
}

func appendShipmentControlDateField(fields *primitive.D, key string, value time.Time) {
	if value.IsZero() {
		return
	}
	*fields = append(*fields, primitive.E{Key: key, Value: value})
}
