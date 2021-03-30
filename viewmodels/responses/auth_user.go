package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserResponse response to client
type AuthUser struct {
	ID           primitive.ObjectID   `bson:"_id"`
	Email        string               `bson:"email"`
	Name         string               `bson:"name"`
	LastName     string               `bson:"lastName"`
	PasswordHash string               `bson:"passwordHash"`
	Phone        string               `bson:"phone"`
	Dni          string               `bson:"dni"`
	Profile      primitive.ObjectID   `bson:"profile"`
	Company      primitive.ObjectID   `bson:"company"`
	Brands       []primitive.ObjectID `bson:"brands"`
	Countries    []primitive.ObjectID `bson:"countries"`
	IsInternal   bool                 `bson:"is_internal"`
}
