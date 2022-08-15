package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestCategory model
type TestCategoryExpanded struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	TestCases   []TestCase         `json:"test_cases" bson:"test_cases"`
}
