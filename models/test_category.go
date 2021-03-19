package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestCategory model
type TestCategory struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty"`
	Name        string               `bson:"name"`
	Description string               `bson:"description"`
	TestCases   []primitive.ObjectID `bson:"test_cases" json:"test_cases"`
}
