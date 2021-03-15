package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestCase model
type TestCase struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Code       string             `bson:"code,omitempty"`
	Name       string             `bson:"description,omitempty"`
	CategoryID primitive.ObjectID `bson:"_id,omitempty"`
}
