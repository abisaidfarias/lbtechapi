package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestCategory model
type TestCategory struct {
	ID          primitive.ObjectID   `json:"_id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	TestCases   []primitive.ObjectID `bson:"test_cases" json:"test_cases"`
}
