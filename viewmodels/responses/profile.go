package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Profile model
type Profile struct {
	ID         primitive.ObjectID `bson:"_id"`
	Name       string             `bson:"name,omitempty"`
	IsInternal bool               `bson:"is_internal,omitempty"`
	Claims     []Claim            `bson:"claims,omitempty"`
}
