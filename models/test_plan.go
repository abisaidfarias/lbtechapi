package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestCase model
type TestPlan struct {
	ID             primitive.ObjectID   `bson:"_id,omitempty"`
	Name           string               `bson:"name,omitempty" json:"name"  `
	TestCategories []primitive.ObjectID `bson:"test_categories,omitempty" json:"test_category"`
	IsActive       bool                 `bson:"is_active" json:"is_active"`
	Description    string               `bson:"description" json:"description"`
	UserName       string               `bson:"username"`
}
