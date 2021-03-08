package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuthResponse response to client
type AuthResponse struct {
	ID    primitive.ObjectID `json:"_id"`
	Email string             `json:"email"`
	Token string             `json:"token"`
}
