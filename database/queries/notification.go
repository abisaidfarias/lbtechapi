package queries

import (
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

func GetNotifictionByCompany(companyId primitive.ObjectID) primitive.M {
	return primitive.M{"company": companyId}
}
func DeleteNotificationbyCompany(oid primitive.ObjectID) primitive.M {

	return primitive.M{
		"company": oid,
	}
}
