package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestCase model
type TestCase struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Code        string             `bson:"code,omitempty" json:"code,omitempty"  binding:"testCaseCode" `
	Name        string             `bson:"name,omitempty" json:"name"  `
	CategoryID  string             `bson:"categoryId,omitempty" json:"categoryId"  `
	IsActive    bool               `bson:"isActive,omitempty" json:"isActive"`
	Description string             `bson:"description,omitempty" json:"description"`
	Device      string             `bson:"device,omitempty" json:"device"`
	Expected    string             `bson:"expected,omitempty" json:"expected"`
}
