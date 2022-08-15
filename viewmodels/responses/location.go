package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Company model
type Location struct {
	ID   primitive.ObjectID `bson:"_id"`
	Name string             `bson:"name"`
}
