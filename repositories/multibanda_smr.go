package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// SmrSourceReport is the last generated report of another request whose
// device has the same technical model: the reference an SMR without testing
// is validated against.
type SmrSourceReport struct {
	ReportID     primitive.ObjectID `bson:"report_id"`
	MultibandaID primitive.ObjectID `bson:"multibanda_id"`
	Type         string             `bson:"type"`
	ReportDate   *time.Time         `bson:"report_date"`
	StampCode    string             `bson:"stamp_code"`
}

// ApplySmrDecision atomically sets the decision fields and appends the audit
// entry, only while the validation is still pending (so two people deciding
// at the same time cannot both win). It returns false when it was no longer
// pending.
func (r *multibandaRepository) ApplySmrDecision(
	id primitive.ObjectID, set primitive.M, entry models.SmrValidationEntry,
) (bool, error) {
	res, err := multibandaCollection.UpdateOne(context.TODO(),
		bson.M{"_id": id, "smr_validation.status": models.SmrValidationPending},
		bson.M{"$set": set, "$push": bson.M{"smr_validation.log": entry}},
	)
	if err != nil {
		return false, err
	}
	return res.MatchedCount > 0, nil
}

// SetSmrValidation replaces the validation (nil removes it).
func (r *multibandaRepository) SetSmrValidation(id primitive.ObjectID, v *models.SmrValidation) error {
	update := bson.M{"$unset": bson.M{"smr_validation": ""}}
	if v != nil {
		update = bson.M{"$set": bson.M{"smr_validation": v}}
	}
	_, err := multibandaCollection.UpdateOne(context.TODO(), bson.M{"_id": id}, update)
	return err
}

// LastGeneratedReportForTechnicalModel finds the most recent generated report
// (by the request's report_date, then the report's generated_at) among
// requests of any type whose device has technicalModel, excluding one
// request.
func (r *multibandaRepository) LastGeneratedReportForTechnicalModel(
	technicalModel string, exclude primitive.ObjectID,
) (*SmrSourceReport, error) {
	if technicalModel == "" {
		return nil, nil
	}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"status": "generated", "multibanda": bson.M{"$ne": exclude}}}},
		{{Key: "$lookup", Value: bson.M{
			"from": "multibandas", "localField": "multibanda", "foreignField": "_id", "as": "mb",
		}}},
		{{Key: "$unwind", Value: "$mb"}},
		{{Key: "$lookup", Value: bson.M{
			"from": "devices", "localField": "mb.device", "foreignField": "_id", "as": "dev",
		}}},
		{{Key: "$unwind", Value: "$dev"}},
		{{Key: "$match", Value: bson.M{"dev.technical_model": technicalModel}}},
		{{Key: "$project", Value: bson.M{
			"report_id":     "$_id",
			"multibanda_id": "$mb._id",
			"type":          "$mb.type",
			"report_date":   bson.M{"$ifNull": bson.A{"$mb.report_date", "$generated_at"}},
			"stamp_code":    bson.M{"$ifNull": bson.A{"$device_info.stamp_code", ""}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "report_date", Value: -1}, {Key: "report_id", Value: -1}}}},
		{{Key: "$limit", Value: 1}},
	}
	cur, err := multibandaReportCollection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}
	var rows []SmrSourceReport
	if err := cur.All(context.TODO(), &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

// SetCommunicationRadio updates only the device's FM Radio flag.
func (r *deviceRepository) SetCommunicationRadio(id primitive.ObjectID, value bool) error {
	res, err := deviceCollection.UpdateOne(context.TODO(), bson.M{"_id": id},
		bson.M{"$set": bson.M{"communication_radio": value}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("device not found")
	}
	return nil
}
