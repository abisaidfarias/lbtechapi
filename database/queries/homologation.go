package queries

import (
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

func GetHomologationValidations(deviceId primitive.ObjectID,
	countryId primitive.ObjectID, companyId primitive.ObjectID) primitive.M {
	return primitive.M{"device": deviceId, "company": companyId,
		"country": countryId, "status": enums.HomologationStatus_value["IN_PROGRESS"]}

}
