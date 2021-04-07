package responses

import "go.mongodb.org/mongo-driver/bson/primitive"

type TestCase struct {
	Code         string             `json:"code,omitempty" binding:"required,testCaseCode"`
	Name         string             `json:"name" binding:"required"`
	IsActive     bool               `bson:"is_active" json:"is_active"`
	Description  string             `json:"description" binding:"required"`
	Expected     string             `json:"expected" binding:"required"`
	TestCategory primitive.ObjectID `json:"test_category" bson:"test_category"`
}
