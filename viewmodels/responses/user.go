package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserResponse response to client
type User struct {
	ID           primitive.ObjectID   `bson:"_id,omitempty"`
	Email        string               `bson:"email,omitempty"`
	PasswordHash string               `bson:"passwordHash,omitempty"`
	Name         string               `bson:"name,omitempty"`
	LastName     string               `bson:"lastName,omitempty"`
	Phone        string               `bson:"phone,omitempty"`
	Dni          string               `bson:"dni,omitempty"`
	Profile      primitive.ObjectID   `bson:"profile"`
	Company      primitive.ObjectID   `bson:"company"`
	Brands       []primitive.ObjectID `bson:"brands"`
	Countries    []primitive.ObjectID `bson:"countries"`
}
