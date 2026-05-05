package queries

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"go.mongodb.org/mongo-driver/bson"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AddTrackingLog(trackingLog *models.TrackingLog, oid primitive.ObjectID) (primitive.M, primitive.M) {

	filter := primitive.M{
		"_id": oid,
	}
	update := primitive.M{
		"$push": primitive.M{
			"tracking_logs": trackingLog,
		},
	}
	return filter, update
}
func GetDeviceTrackingExpanded(isInternal bool, companyID primitive.ObjectID,
	brands []primitive.ObjectID, countries []string, companies []primitive.ObjectID) []bson.D {
	lookupCompany := bson.D{
		primitive.E{Key: "$lookup", Value: bson.D{
			primitive.E{Key: "from", Value: "companies"},
			primitive.E{Key: "localField", Value: "company"},
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "company"},
		}}}
	unwindCompany := bson.D{
		primitive.E{Key: "$unwind", Value: bson.D{
			primitive.E{Key: "path", Value: "$company"},
			primitive.E{Key: "preserveNullAndEmptyArrays", Value: true},
		}}}
	lookupDevice := bson.D{
		primitive.E{Key: "$lookup", Value: bson.D{
			primitive.E{Key: "from", Value: "devices"},
			primitive.E{Key: "localField", Value: "device"},
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "device"},
		}}}
	unwindDevice := bson.D{
		primitive.E{Key: "$unwind", Value: bson.D{
			primitive.E{Key: "path", Value: "$device"},
			primitive.E{Key: "preserveNullAndEmptyArrays", Value: true},
		}}}
	lookupBrand := bson.D{
		primitive.E{Key: "$lookup", Value: bson.D{
			primitive.E{Key: "from", Value: "brands"},
			primitive.E{Key: "localField", Value: "device.brand"},
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "device.brand"},
		}}}
	unwindBrand := bson.D{
		primitive.E{Key: "$unwind", Value: bson.D{
			primitive.E{Key: "path", Value: "$device.brand"},
			primitive.E{Key: "preserveNullAndEmptyArrays", Value: true},
		}}}

	var objectStage bson.D
	var matchStage bson.D
	hasStage := false
	if len(brands) > 0 {

		var brandStageIn bson.D
		brandStageIn = append(brandStageIn, primitive.E{Key: "$in", Value: brands})
		objectStage = append(matchStage, primitive.E{Key: "device.brand._id", Value: brandStageIn})
		hasStage = true
	}
	if len(countries) > 0 {
		var countryStageIn bson.D
		countryStageIn = append(countryStageIn, primitive.E{Key: "$in", Value: countries})
		objectStage = append(objectStage, primitive.E{Key: "tracking_logs.country.name",
			Value: countryStageIn})
		hasStage = true
	}
	if len(companies) > 0 {
		var companyStageIn bson.D
		companyStageIn = append(companyStageIn, primitive.E{Key: "$in", Value: companies})
		objectStage = append(objectStage, primitive.E{Key: "company._id",
			Value: companyStageIn})
		hasStage = true
	}

	if hasStage {
		if !isInternal {
			objectStage = append(objectStage, primitive.E{Key: "company._id", Value: companyID})
		}
		matchStage = append(matchStage, primitive.E{Key: "$match", Value: objectStage})
		return mongo.Pipeline{lookupCompany, unwindCompany,
			lookupDevice, unwindDevice, lookupBrand, unwindBrand, matchStage}
	}
	return mongo.Pipeline{lookupCompany, unwindCompany,
		lookupDevice, unwindDevice, lookupBrand, unwindBrand}
}
func GetDeviceTrackingByCompany(companyId primitive.ObjectID) primitive.M {
	return primitive.M{"company": companyId}
}
func GetDeviceTrackingByDevice(deviceId primitive.ObjectID) primitive.M {
	return primitive.M{"device": deviceId}
}
func GetDeviceTrackingByPerson(personId primitive.ObjectID) primitive.M {
	return primitive.M{"device_tracking.tracking_logs.person": personId}
}
func DeleteDeviceTracking(ids []primitive.ObjectID) primitive.M {

	filter := primitive.M{
		"_id": primitive.M{
			"$in": ids,
		},
	}
	return filter
}
func UpdateDeviceTracking(deviceTracking *models.DeviceTracking, oid primitive.ObjectID) (primitive.M, primitive.M) {

	filter := primitive.M{
		"_id": oid,
	}
	update := primitive.M{
		"$set": deviceTracking,
	}
	return filter, update
}
func GetDeviceTrackingAdvancedSearch(companyID primitive.ObjectID,
	isInternal bool, searchOption *request.SearchOption,
	brands []primitive.ObjectID, countries []string, companies []primitive.ObjectID) []bson.D {
	lookupCompany := bson.D{
		primitive.E{Key: "$lookup", Value: bson.D{
			primitive.E{Key: "from", Value: "companies"},
			primitive.E{Key: "localField", Value: "company"},
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "company"},
		}}}
	unwindCompany := bson.D{
		primitive.E{Key: "$unwind", Value: bson.D{
			primitive.E{Key: "path", Value: "$company"},
			primitive.E{Key: "preserveNullAndEmptyArrays", Value: true},
		}}}
	lookupDevice := bson.D{
		primitive.E{Key: "$lookup", Value: bson.D{
			primitive.E{Key: "from", Value: "devices"},
			primitive.E{Key: "localField", Value: "device"},
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "device"},
		}}}
	unwindDevice := bson.D{
		primitive.E{Key: "$unwind", Value: bson.D{
			primitive.E{Key: "path", Value: "$device"},
			primitive.E{Key: "preserveNullAndEmptyArrays", Value: true},
		}}}
	lookupBrand := bson.D{
		primitive.E{Key: "$lookup", Value: bson.D{
			primitive.E{Key: "from", Value: "brands"},
			primitive.E{Key: "localField", Value: "device.brand"},
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "device.brand"},
		}}}
	unwindBrand := bson.D{
		primitive.E{Key: "$unwind", Value: bson.D{
			primitive.E{Key: "path", Value: "$device.brand"},
			primitive.E{Key: "preserveNullAndEmptyArrays", Value: true},
		}}}

	var objectStage bson.D
	var matchStage bson.D
	var hasStage = false
	if len(brands) > 0 {

		var brandStageIn bson.D
		brandStageIn = append(brandStageIn, primitive.E{Key: "$in", Value: brands})
		objectStage = append(matchStage, primitive.E{Key: "device.brand._id", Value: brandStageIn})
		hasStage = true
	}
	if len(countries) > 0 {
		var countryStageIn bson.D
		countryStageIn = append(countryStageIn, primitive.E{Key: "$in", Value: countries})
		objectStage = append(objectStage, primitive.E{Key: "tracking_logs.country.name",
			Value: countryStageIn})
		hasStage = true
	}
	if len(companies) > 0 {
		var companyStageIn bson.D
		companyStageIn = append(companyStageIn, primitive.E{Key: "$in", Value: companies})
		objectStage = append(objectStage, primitive.E{Key: "company._id",
			Value: companyStageIn})
		hasStage = true
	}
	if len(searchOption.Brands) > 0 {
		var brandStageIn bson.D
		brandStageIn = append(brandStageIn, primitive.E{Key: "$in", Value: searchOption.Brands})
		objectStage = append(objectStage, primitive.E{Key: "device.brand.name",
			Value: brandStageIn})
		hasStage = true
	}
	if len(searchOption.CommercialModels) > 0 {
		var modelStageIn bson.D
		modelStageIn = append(modelStageIn, primitive.E{Key: "$in", Value: searchOption.CommercialModels})
		objectStage = append(objectStage, primitive.E{Key: "device.commercial_model",
			Value: modelStageIn})
		hasStage = true
	}
	if len(searchOption.Countries) > 0 {
		var countryStageIn bson.D
		countryStageIn = append(countryStageIn, primitive.E{Key: "$in", Value: searchOption.Countries})
		objectStage = append(objectStage, primitive.E{Key: "tracking_logs.country.name",
			Value: countryStageIn})
		hasStage = true
	}
	if len(searchOption.Locations) > 0 {
		var locationStageIn bson.D
		locationStageIn = append(locationStageIn, primitive.E{Key: "$in", Value: searchOption.Locations})
		objectStage = append(objectStage, primitive.E{Key: "tracking_logs.location.name",
			Value: locationStageIn})
		hasStage = true
	}
	if !isInternal {
		objectStage = append(objectStage, primitive.E{Key: "company._id", Value: companyID})
	}
	if hasStage {
		matchStage = append(matchStage, primitive.E{Key: "$match", Value: objectStage})
		return mongo.Pipeline{lookupCompany, unwindCompany,
			lookupDevice, unwindDevice, lookupBrand, unwindBrand, matchStage}
	}
	return mongo.Pipeline{lookupCompany, unwindCompany,
		lookupDevice, unwindDevice, lookupBrand, unwindBrand}

}
func GetDeviceTrackingById(oid primitive.ObjectID) primitive.M {
	return primitive.M{"_id": oid}
}
