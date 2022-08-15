package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestCategory model
type TestCategory struct {
	ID          primitive.ObjectID   `json:"_id" bson:"_id"`
	Name        string               `json:"name" bson:"name"`
	Description string               `json:"description" bson:"description"`
	TestCases   []primitive.ObjectID `json:"test_cases" bson:"test_cases"`
}
