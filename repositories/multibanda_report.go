package repositories

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
)

type IMultibandaReportRepository interface {
	// GetByMultibanda returns the report for a process, or nil when the
	// engineer has not started one yet.
	GetByMultibanda(primitive.ObjectID) (*models.MultibandaReport, error)
	// Upsert saves the draft/edited report, creating it on first save.
	Upsert(*models.MultibandaReport) error
	// MarkGenerated stores the rendered PDF against the report.
	MarkGenerated(multibandaID primitive.ObjectID, reportURL, generatedBy string, generatedAt time.Time) error
}

type multibandaReportRepository struct{}

func NewMultibandaReportRepository() IMultibandaReportRepository {
	return &multibandaReportRepository{}
}

var multibandaReportCollection = database.GetInstance().Collection("multibanda_reports")

func (r *multibandaReportRepository) GetByMultibanda(multibandaID primitive.ObjectID) (*models.MultibandaReport, error) {
	var report models.MultibandaReport
	err := multibandaReportCollection.
		FindOne(context.TODO(), bson.M{"multibanda": multibandaID}).
		Decode(&report)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *multibandaReportRepository) Upsert(report *models.MultibandaReport) error {
	now := time.Now()
	report.UpdatedDate = now

	filter := bson.M{"multibanda": report.Multibanda}
	update := bson.M{
		"$set": bson.M{
			"status":           report.Status,
			"multibanda_type":  report.MultibandaType,
			"device_info":      report.DeviceInfo,
			"simlock_results":  report.SimlockResults,
			"band_results":     report.BandResults,
			"sae_scenario":     report.SAEScenario,
			"sae_results":      report.SAEResults,
			"evidence":         report.Evidence,
			"carriers_tested":  report.CarriersTested,
			"fm_radio_result":  report.FMRadioResult,
			"fm_radio_comment": report.FMRadioComment,
			"updated_date":     now,
		},
		// created_date and the generated PDF fields survive a draft save:
		// re-saving a report that already produced a PDF must not clear it.
		"$setOnInsert": bson.M{
			"multibanda":   report.Multibanda,
			"created_date": now,
		},
	}

	_, err := multibandaReportCollection.UpdateOne(
		context.TODO(), filter, update, options.Update().SetUpsert(true),
	)
	return err
}

func (r *multibandaReportRepository) MarkGenerated(
	multibandaID primitive.ObjectID,
	reportURL, generatedBy string,
	generatedAt time.Time,
) error {
	filter := bson.M{"multibanda": multibandaID}
	update := bson.M{
		"$set": bson.M{
			"status":       enums.MultibandaReportStatusGenerated,
			"report_url":   reportURL,
			"generated_at": generatedAt,
			"generated_by": generatedBy,
			"updated_date": generatedAt,
		},
	}

	res, err := multibandaReportCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
