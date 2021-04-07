package queries

import (
	"go.mongodb.org/mongo-driver/bson"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetCategoriesByIds(categoriesId []primitive.ObjectID) []bson.D {

	lookupStage := bson.D{
		primitive.E{Key: "$lookup", Value: bson.D{
			primitive.E{Key: "from", Value: "test_cases"},
			primitive.E{Key: "localField", Value: "test_cases"},
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "test_cases"},
		}}}
	// matchStage := bson.D{
	// 	primitive.E{Key: "$match", Value: bson.D{
	// 		primitive.E{Key: "_id", Value: categoriesId},
	// 	}}}
	matchStage := bson.D{
		primitive.E{Key: "$match", Value: bson.D{
			primitive.E{Key: "_id", Value: bson.D{
				primitive.E{Key: "$in", Value: categoriesId},
			}}},
		},
	}
	return mongo.Pipeline{lookupStage, matchStage}

}
