package queries

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"go.mongodb.org/mongo-driver/bson"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetTestCases() []bson.D {
	lookupStage := bson.D{
		primitive.E{Key: "$lookup", Value: bson.D{
			primitive.E{Key: "from", Value: "test_categories"},
			primitive.E{Key: "localField", Value: "test_category"},
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "test_category"},
		}}}
	unwindStage := bson.D{
		primitive.E{Key: "$unwind", Value: bson.D{
			primitive.E{Key: "path", Value: "$test_category"},
			primitive.E{Key: "preserveNullAndEmptyArrays", Value: true},
		}}}
	matchStage := bson.D{
		primitive.E{Key: "$match", Value: bson.D{
			primitive.E{Key: "is_active", Value: true},
		}}}
	return mongo.Pipeline{lookupStage, unwindStage, matchStage}
}
func GeTestCaseById(oid primitive.ObjectID) primitive.M {
	return primitive.M{"_id": oid}
}
func UpdateTestCase(testCase *models.TestCase, oid primitive.ObjectID) (primitive.M, primitive.M) {

	filter := primitive.M{
		"_id": oid,
	}
	update := primitive.M{
		"$set": testCase,
	}
	return filter, update
}
func DeleteTestCase(oid primitive.ObjectID) primitive.M {

	return primitive.M{
		"_id": oid,
	}
}
