package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Profile model
type Profile struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Name       string             `bson:"name,omitempty"`
	Claims     []Claim            `bson:"claims,omitempty"`
	IsInternal bool               `bson:"is_internal" json:"is_internal"`
}
