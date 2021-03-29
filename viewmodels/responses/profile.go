package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Profile model
type Profile struct {
	ID         primitive.ObjectID `bson:"_id" json:"_id"`
	Name       string             `json:"name"`
	IsInternal bool               `json:"is_internal"`
	Claims     []Claim            `json:"claims"`
}
