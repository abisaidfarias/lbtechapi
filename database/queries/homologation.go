package queries

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"go.mongodb.org/mongo-driver/bson"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetHomologationValidations(deviceId primitive.ObjectID,
	countryId primitive.ObjectID, companyId primitive.ObjectID, isInternal bool) primitive.M {
	return primitive.M{"device": deviceId, "company": companyId,
		"country": countryId, "is_internal_project": isInternal}

}
func GetHomologationById(oid primitive.ObjectID) primitive.M {
	return primitive.M{"_id": oid}
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
	sort := bson.D{
		primitive.E{Key: "$sort", Value: bson.D{
			primitive.E{Key: "completed_date", Value: 1},
			primitive.E{Key: "test_start_date", Value: -1},
			primitive.E{Key: "planning_date", Value: -1},
			primitive.E{Key: "sample_start_date", Value: -1},
		}}}

	var matchStage bson.D
	var objectStage bson.D
	var hasStage bool
	if len(companies) > 0 && isInternal {

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
			lookupStageBrand, unwindStageBrand, matchStage, sort}
	} else {
		pipeline = mongo.Pipeline{lookupStageDevice, unwindStageDevice,
			lookupStageCompany, unwindStageCompany,
			lookupStageCountry, unwindStageCountry,
			lookupStageTestPlan, unwindStageTestPlan,
			lookupStageBrand, unwindStageBrand, sort}
	}

	return pipeline

}
func GetHomologationsGroupedCountryApprovalType() bson.D {

	return bson.D{
		primitive.E{Key: "$group", Value: bson.D{
			primitive.E{Key: "_id", Value: bson.D{
				primitive.E{Key: "year", Value: bson.D{
					primitive.E{Key: "$year", Value: "$test_start_date"},
				}},
				primitive.E{Key: "month", Value: bson.D{
					primitive.E{Key: "$month", Value: "$test_start_date"},
				}},
				primitive.E{Key: "country", Value: "$country.name"},
				primitive.E{Key: "type", Value: "$type"},
			}},
			primitive.E{Key: "count", Value: bson.D{
				primitive.E{Key: "$sum", Value: 1},
			}},
		}}}

}
func GetHomologationsGroupedCountryBrand() bson.D {

	return bson.D{
		primitive.E{Key: "$group", Value: bson.D{
			primitive.E{Key: "_id", Value: bson.D{
				primitive.E{Key: "country", Value: "$country.name"},
				primitive.E{Key: "brand", Value: "$brand.name"},
			}},
			primitive.E{Key: "count", Value: bson.D{
				primitive.E{Key: "$sum", Value: 1},
			}},
		}}}

}
func GetHomologationsGroupedTypeBrand() bson.D {

	return bson.D{
		primitive.E{Key: "$group", Value: bson.D{
			primitive.E{Key: "_id", Value: bson.D{
				primitive.E{Key: "type", Value: "$type"},
				primitive.E{Key: "brand", Value: "$brand.name"},
			}},
			primitive.E{Key: "count", Value: bson.D{
				primitive.E{Key: "$sum", Value: 1},
			}},
		}}}

}
func SortGroupedBrandType() bson.D {

	return bson.D{
		primitive.E{Key: "$sort", Value: bson.D{
			primitive.E{Key: "_id.brand", Value: 1},
			primitive.E{Key: "_id.type", Value: 1},
		}}}

}
func SortGroupedCountryApprovalType() bson.D {

	return bson.D{
		primitive.E{Key: "$sort", Value: bson.D{
			primitive.E{Key: "_id.year", Value: 1},
			primitive.E{Key: "_id.month", Value: 1},
			primitive.E{Key: "_id.country", Value: 1},
			primitive.E{Key: "_id.type", Value: 1},
		}}}

}
func SortGroupedCountryBrand() bson.D {

	return bson.D{
		primitive.E{Key: "$sort", Value: bson.D{
			primitive.E{Key: "_id.country", Value: 1},
			primitive.E{Key: "_id.brand", Value: 1},
		}}}

}
func UpdateTestResult(testResult request.TestResultResume, oid primitive.ObjectID) (primitive.M, primitive.M) {

	filter := primitive.M{
		"_id":               oid,
		"test_results.code": testResult.Code,
	}
	update := primitive.M{
		"$set": primitive.M{
			"test_results.$.overview_issue":     testResult.OverviewIssue,
			"test_results.$.steps_to_reproduce": testResult.StepsToReproduce,
			"test_results.$.actual_result":      testResult.ActualResult,
			"test_results.$.expected_result":    testResult.ExpectedResult,
			"test_results.$.issue_frequency":    testResult.IssueFrequency,
			"test_results.$.issue_severity":     testResult.IssueSeverity,
			"test_results.$.hyperlinks":         testResult.Hyperlinks,
			"test_results.$.images":             testResult.Images,
			"test_results.$.result":             testResult.Result,
			"test_results.$.value":              testResult.Value,
		},
	}
	return filter, update
}
func CreateTestResult(testResult *models.TestResult, oid primitive.ObjectID) (primitive.M, primitive.M) {

	filter := primitive.M{
		"_id": oid,
	}
	update := primitive.M{
		"$push": primitive.M{
			"test_results": testResult,
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
	return filter, update
}
func GetHomologationByTestPlan(oid primitive.ObjectID) primitive.M {
	return primitive.M{"test_plan": oid}
}
func UpdateDocument(documentUrl string, homologationId primitive.ObjectID) (primitive.M, primitive.D) {

	filter := primitive.M{
		"_id": homologationId,
	}
	update := primitive.D{
		{Key: "$set",
			Value: primitive.D{
				{Key: "document_url", Value: documentUrl},
			},
		},
	}
	return filter, update
}
func GetHomologationsGroupedByStatus(companyId primitive.ObjectID) []bson.D {

	groupStage := bson.D{
		primitive.E{Key: "$group", Value: bson.D{
			primitive.E{Key: "_id", Value: "$current_phase"},
			primitive.E{Key: "count", Value: bson.D{
				primitive.E{Key: "$sum", Value: 1},
			}},
		}}}

	return mongo.Pipeline{groupStage}
}
