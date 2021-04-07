package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestCase model
type TestResult struct {
	Code             string             `bson:"code" json:"code"`
	Name             string             `bson:"name" json:"name"`
	TestCategory     primitive.ObjectID `bson:"test_category" json:"test_category"`
	IsActive         bool               `bson:"is_active" json:"is_active"`
	Description      string             `bson:"description" json:"description"`
	Expected         string             `bson:"expected" json:"expected"`
	Result           int                `bson:"result" json:"result"`
	IssueDescription string             `bson:"issue_description" json:"issue_description"`
	Comment          string             `bson:"comment" json:"comment"`
}
