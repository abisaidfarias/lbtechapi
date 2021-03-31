package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuthResponse response to client
type AuthResponse struct {
	ID       primitive.ObjectID `json:"_id" bson:"_id"`
	Profile  Profile            `json:"profile" bson:"profile"`
	Email    string             `json:"email" bson:"email"`
	Token    string             `json:"token" bson:"token"`
	Name     string             `json:"name" bson:"name"`
	LastName string             `json:"lastName" bson:"lastName"`
}
