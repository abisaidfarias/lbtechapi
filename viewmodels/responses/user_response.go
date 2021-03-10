package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserResponse response to client
type UserResponse struct {
	ID        primitive.ObjectID `json:"_id"`
	Email     string             `json:"email"`
	Name      string             `json:"name"`
	LastName  string             `json:"lastName"`
	Phone     string             `json:"phone"`
	Dni       string             `json:"dni"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at,omitempty"`
}
