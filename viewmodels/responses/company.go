package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Company model
type Company struct {
	ID      primitive.ObjectID `bson:"_id"`
	Email   string             `bson:"email"`
	Name    string             `bson:"name"`
	Address string             `bson:"address"`
}
