package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Country model
type Country struct {
	ID      primitive.ObjectID `bson:"_id"`
	Name    string             `bson:"name"`
}
