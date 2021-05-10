package queries

import (
	"log"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"go.mongodb.org/mongo-driver/bson"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetHomologationValidations(deviceId primitive.ObjectID,
	countryId primitive.ObjectID, companyId primitive.ObjectID) primitive.M {
	return primitive.M{"device": deviceId, "company": companyId,
		"country": countryId, "status": enums.HomologationStatus_value["IN_PROGRESS"]}

}
func GetHomologationById(oid primitive.ObjectID) []bson.D {
	matchStage := bson.D{
		primitive.E{Key: "$match", Value: bson.D{
			primitive.E{Key: "_id", Value: oid},
		}}}
	return mongo.Pipeline{matchStage}
}
func GetHomologations(companies []primitive.ObjectID,
	brands []primitive.ObjectID, countries []primitive.ObjectID,
	isInternal bool, companyID primitive.ObjectID) []bson.D {

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
	lookupStageCountry := bson.D{
		primitive.E{Key: "$lookup", Value: bson.D{
			primitive.E{Key: "from", Value: "countries"},
			primitive.E{Key: "localField", Value: "country"},
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "country"},
		}}}
	unwindStageCountry := bson.D{
		primitive.E{Key: "$unwind", Value: bson.D{
			primitive.E{Key: "path", Value: "$country"},
			primitive.E{Key: "preserveNullAndEmptyArrays", Value: true},
		}}}
	lookupStageTestPlan := bson.D{
		primitive.E{Key: "$lookup", Value: bson.D{
			primitive.E{Key: "from", Value: "test_plans"},
			primitive.E{Key: "localField", Value: "test_plan"},
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "test_plan"},
		}}}
	unwindStageTestPlan := bson.D{
		primitive.E{Key: "$unwind", Value: bson.D{
			primitive.E{Key: "path", Value: "$test_plan"},
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
	var matchStage bson.D
	var objectStage bson.D
	var hasStage bool
	if len(companies) > 0 && isInternal == true {

		var companyStage bson.D
		companyStage = append(companyStage, primitive.E{Key: "$in", Value: companies})
		objectStage = append(objectStage, primitive.E{Key: "company._id", Value: companyStage})
		hasStage = true
	}
	if !isInternal {
		objectStage = append(objectStage, primitive.E{Key: "company._id", Value: companyID})
		hasStage = true
	}
	if len(brands) > 0 {

		var brandStageIn bson.D
		brandStageIn = append(brandStageIn, primitive.E{Key: "$in", Value: brands})
		objectStage = append(objectStage, primitive.E{Key: "brand._id", Value: brandStageIn})
		hasStage = true
	}
	if len(countries) > 0 {

		var countryStage bson.D
		countryStage = append(countryStage, primitive.E{Key: "$in", Value: countries})
		objectStage = append(objectStage, primitive.E{Key: "country._id", Value: countryStage})
		hasStage = true
	}
	matchStage = append(matchStage, primitive.E{Key: "$match", Value: objectStage})

	var pipeline []bson.D
	if hasStage {
		pipeline = mongo.Pipeline{lookupStageDevice, unwindStageDevice,
			lookupStageCompany, unwindStageCompany,
			lookupStageCountry, unwindStageCountry,
			lookupStageTestPlan, unwindStageTestPlan,
			lookupStageBrand, unwindStageBrand, matchStage}
	} else {
		pipeline = mongo.Pipeline{lookupStageDevice, unwindStageDevice,
			lookupStageCompany, unwindStageCompany,
			lookupStageCountry, unwindStageCountry,
			lookupStageTestPlan, unwindStageTestPlan,
			lookupStageBrand, unwindStageBrand}
	}
	log.Println("pipeline", pipeline)

	return pipeline

}
func UpdateTestResult(testResult request.TestResultResume, oid primitive.ObjectID) (primitive.M, primitive.M) {

	filter := primitive.M{
		"_id":               oid,
		"test_results.code": testResult.Code,
	}
	update := primitive.M{
		"$set": primitive.M{
			"test_results.$.issue_description": testResult.IssueDescription,
			"test_results.$.issue_frequency":   testResult.IssueFrequency,
			"test_results.$.issue_severity":    testResult.IssueSeverity,
			"test_results.$.hyperlinks":        testResult.Hyperlinks,
			"test_results.$.images":            testResult.Images,
			"test_results.$.result":            testResult.Result,
			"test_results.$.value":             testResult.Value,
		},
	}
	return filter, update
}
func UpdatePhaseChange(homologation *models.Homologation, oid primitive.ObjectID) (primitive.M, primitive.D) {

	filter := primitive.M{
		"_id": oid,
	}
	update := primitive.D{
		{Key: "$set",
			Value: primitive.D{
				{Key: "current_phase", Value: homologation.CurrentPhase},
				{Key: "status", Value: homologation.Status},
				{Key: "software_version", Value: homologation.SoftwareVersion},
				{Key: "hardware_version", Value: homologation.HardwareVersion},
				{Key: "planning_date", Value: homologation.PlanningDate},
				{Key: "sample_start_date", Value: homologation.SampleStartDate},
				{Key: "sample_end_date", Value: homologation.SampleEndDate},
				{Key: "test_start_date", Value: homologation.TestStartDate},
				{Key: "test_end_date", Value: homologation.TestEndDate},
				{Key: "under_start_date", Value: homologation.UnderStartDate},
				{Key: "under_end_date", Value: homologation.UnderEndDate},
				{Key: "completed_date", Value: homologation.CompletedDate},
			},
		},
	}
	// 	"current_phase":     homologation.CurrentPhase,
	// 	"status":            homologation.Status,
	// 	"software_version":  homologation.SoftwareVersion,
	// 	"hardware_version":  homologation.HardwareVersion,
	// 	"planning_date":     homologation.PlanningDate,
	// 	"sample_start_date": homologation.SampleStartDate,
	// 	"sample_end_date":   homologation.SampleEndDate,
	// 	"test_start_date":   homologation.TestStartDate,
	// 	"test_end_date":     homologation.TestEndDate,
	// 	"under_start_date":  homologation.UnderStartDate,
	// 	"under_end_date":    homologation.UnderEndDate,
	// 	"completed_date":    homologation.CompletedDate,
	// },
	//}
	return filter, update
}
