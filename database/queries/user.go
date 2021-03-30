package queries

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"go.mongodb.org/mongo-driver/bson"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetUsersProfileId(id string) primitive.M {
	val, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		panic(err)
	}
	return primitive.M{"profile": val}
}

func UpdateUser(user *models.User, oid primitive.ObjectID) (primitive.M, primitive.M) {

	filter := primitive.M{
		"_id": oid,
	}
	update := primitive.M{
		"$set": user,
	}
	return filter, update
}
func DeleteUser(oid primitive.ObjectID) primitive.M {

	return primitive.M{
		"_id": oid,
	}
}
func GetUserById(oid primitive.ObjectID) primitive.M {
	return primitive.M{"_id": oid}
}
func GetUserByEmail(email string) primitive.M {
	return primitive.M{"email": email}
}
func GetUserByCompany(companyId primitive.ObjectID) primitive.M {
	return primitive.M{"company": companyId}
}

func GetUserProfileById() []bson.D {
	lookupStage := bson.D{
		primitive.E{Key: "$lookup", Value: bson.D{
			primitive.E{Key: "from", Value: "profiles"},
			primitive.E{Key: "localField", Value: "profile"},
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "profile"},
		}}}
	unwindStage := bson.D{
		primitive.E{Key: "$unwind", Value: bson.D{
			primitive.E{Key: "path", Value: "$profile"},
			primitive.E{Key: "preserveNullAndEmptyArrays", Value: false},
		}}}
	lookupStage2 := bson.D{
		primitive.E{Key: "$lookup", Value: bson.D{
			primitive.E{Key: "from", Value: "companies"},
			primitive.E{Key: "localField", Value: "company"},
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "company"},
		}}}
	unwindStage2 := bson.D{
		primitive.E{Key: "$unwind", Value: bson.D{
			primitive.E{Key: "path", Value: "$company"},
			primitive.E{Key: "preserveNullAndEmptyArrays", Value: false},
		}}}
	// matchStage := bson.D{
	// 	primitive.E{Key: "$match", Value: bson.D{
	// 		primitive.E{Key: "_id", Value: oid},
	// 	}}}
	return mongo.Pipeline{lookupStage, unwindStage,lookupStage2,unwindStage2}
}
