package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestCase model
type TestCase struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Code         string             `bson:"code,omitempty" json:"code,omitempty"  binding:"testCaseCode" `
	Name         string             `bson:"name,omitempty" json:"name"  `
	TestCategory primitive.ObjectID `bson:"test_category,omitempty" json:"test_category"`
	IsActive     bool               `bson:"is_active" json:"is_active"`
	Description  string             `bson:"description,omitempty" json:"description"`
	Device       string             `bson:"device,omitempty" json:"device"`
	Expected     string             `bson:"expected,omitempty" json:"expected"`
}
