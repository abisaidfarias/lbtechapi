package queries

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"go.mongodb.org/mongo-driver/bson"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

func GetTestPlans() []bson.M {
	lookupStage := []bson.M{
		{
			"$lookup": bson.M{
				"from":         "test_categories",
				"localField":   "test_categories",
				"foreignField": "_id",
				"as":           "test_categories",
			},
		},
	}
	return lookupStage
}

func GeTestPlanById(oid primitive.ObjectID) primitive.M {
	return primitive.M{"_id": oid}
}
func UpdateTestPlan(testPlan *models.TestPlan, oid primitive.ObjectID) (primitive.M, primitive.M) {

	filter := primitive.M{
		"_id": oid,
	}
	update := primitive.M{
		"$set": testPlan,
	}
	return filter, update
}
func DeleteTestPlan(oid primitive.ObjectID) primitive.M {

	return primitive.M{
		"_id": oid,
	}
}
