package queries

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/abisaidfarias/lbtechapi/models"
)

func GetMultibandaByCompanyDeviceSoftwareOsVersion(
	companyID primitive.ObjectID,
	deviceID primitive.ObjectID,
	softwareVersion string,
	osVersion string,
) primitive.M {
	return primitive.M{
		"company":          companyID,
		"device":           deviceID,
		"software_version": softwareVersion,
		"os_version":       osVersion,
	}
}

func GetMultibandaById(oid primitive.ObjectID) primitive.M {
	return primitive.M{"_id": oid}
}

func GetMultibandas(
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
	if len(brands) > 0 {
		preMatch = append(preMatch, primitive.E{
			Key:   "brand",
			Value: bson.D{primitive.E{Key: "$in", Value: brands}},
		})
		hasPreMatch = true
	}

	lookupStageDevice := bson.D{
		primitive.E{Key: "$lookup", Value: bson.D{
			primitive.E{Key: "from", Value: "devices"},
			primitive.E{Key: "localField", Value: "device"},
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "device"},
		}}}
	unwindStageDevice := bson.D{
		primitive.E{Key: "$unwind", Value: bson.D{
			primitive.E{Key: "path", Value: "$device"},
			primitive.E{Key: "preserveNullAndEmptyArrays", Value: true},
		}}}
	lookupStageCompany := bson.D{
		primitive.E{Key: "$lookup", Value: bson.D{
			primitive.E{Key: "from", Value: "companies"},
			primitive.E{Key: "localField", Value: "company"},
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "company"},
		}}}
	unwindStageCompany := bson.D{
		primitive.E{Key: "$unwind", Value: bson.D{
			primitive.E{Key: "path", Value: "$company"},
			primitive.E{Key: "preserveNullAndEmptyArrays", Value: true},
		}}}
	lookupStageBrand := bson.D{
		primitive.E{Key: "$lookup", Value: bson.D{
			primitive.E{Key: "from", Value: "brands"},
			primitive.E{Key: "localField", Value: "brand"},
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "brand"},
		}}}
	unwindStageBrand := bson.D{
		primitive.E{Key: "$unwind", Value: bson.D{
			primitive.E{Key: "path", Value: "$brand"},
			primitive.E{Key: "preserveNullAndEmptyArrays", Value: true},
		}}}
	sort := bson.D{
		primitive.E{Key: "$sort", Value: bson.D{
			primitive.E{Key: "planning_date", Value: -1},
			primitive.E{Key: "created_date", Value: -1},
		}}}

	pipeline := mongo.Pipeline{}
	if hasPreMatch {
		pipeline = append(pipeline, bson.D{primitive.E{Key: "$match", Value: preMatch}})
	}
	pipeline = append(pipeline,
		lookupStageDevice, unwindStageDevice,
		lookupStageCompany, unwindStageCompany,
		lookupStageBrand, unwindStageBrand,
		sort,
	)

	return pipeline
}

func GetMultibandaExpandedById(oid primitive.ObjectID) mongo.Pipeline {
	matchStage := bson.D{
		primitive.E{Key: "$match", Value: bson.D{
			primitive.E{Key: "_id", Value: oid},
		}}}

	lookupStageDevice := bson.D{
		primitive.E{Key: "$lookup", Value: bson.D{
			primitive.E{Key: "from", Value: "devices"},
			primitive.E{Key: "localField", Value: "device"},
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "device"},
		}}}
	unwindStageDevice := bson.D{
		primitive.E{Key: "$unwind", Value: bson.D{
			primitive.E{Key: "path", Value: "$device"},
			primitive.E{Key: "preserveNullAndEmptyArrays", Value: true},
		}}}
	lookupStageCompany := bson.D{
		primitive.E{Key: "$lookup", Value: bson.D{
			primitive.E{Key: "from", Value: "companies"},
			primitive.E{Key: "localField", Value: "company"},
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "company"},
		}}}
	unwindStageCompany := bson.D{
		primitive.E{Key: "$unwind", Value: bson.D{
			primitive.E{Key: "path", Value: "$company"},
			primitive.E{Key: "preserveNullAndEmptyArrays", Value: true},
		}}}
	lookupStageBrand := bson.D{
		primitive.E{Key: "$lookup", Value: bson.D{
			primitive.E{Key: "from", Value: "brands"},
			primitive.E{Key: "localField", Value: "brand"},
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "brand"},
		}}}
	unwindStageBrand := bson.D{
		primitive.E{Key: "$unwind", Value: bson.D{
			primitive.E{Key: "path", Value: "$brand"},
			primitive.E{Key: "preserveNullAndEmptyArrays", Value: true},
		}}}

	return mongo.Pipeline{
		matchStage,
		lookupStageDevice, unwindStageDevice,
		lookupStageCompany, unwindStageCompany,
		lookupStageBrand, unwindStageBrand,
	}
}

func UpdateMultibandaPhaseChange(multibanda *models.Multibanda, oid primitive.ObjectID) (primitive.M, primitive.D) {
	filter := primitive.M{
		"_id": oid,
	}
	update := primitive.D{
		{Key: "$set",
			Value: primitive.D{
				{Key: "current_phase", Value: multibanda.CurrentPhase},
				{Key: "status", Value: multibanda.Status},
				{Key: "software_version", Value: multibanda.SoftwareVersion},
				{Key: "hardware_version", Value: multibanda.HardwareVersion},
				{Key: "os_version", Value: multibanda.OsVersion},
				{Key: "os_version_view", Value: multibanda.OsVersionView},
				{Key: "planning_date", Value: multibanda.PlanningDate},
				{Key: "sample_start_date", Value: multibanda.SampleStartDate},
				{Key: "sample_end_date", Value: multibanda.SampleEndDate},
				{Key: "test_start_date", Value: multibanda.TestStartDate},
				{Key: "test_end_date", Value: multibanda.TestEndDate},
				{Key: "under_start_date", Value: multibanda.UnderStartDate},
				{Key: "under_end_date", Value: multibanda.UnderEndDate},
				{Key: "completed_date", Value: multibanda.CompletedDate},
				{Key: "test_report_url", Value: multibanda.TestReportUrl},
				{Key: "multiband_certificate_url", Value: multibanda.MultibandCertificateUrl},
				{Key: "certificate_number", Value: multibanda.CertificateNumber},
				{Key: "dashboard_phase", Value: multibanda.DashBoardPhase},
			},
		},
	}
	return filter, update
}
