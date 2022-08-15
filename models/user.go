package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User model
type User struct {
	ID           primitive.ObjectID   `bson:"_id,omitempty"`
	Email        string               `bson:"email,omitempty"`
	PasswordHash string               `bson:"passwordHash,omitempty"`
	Name         string               `bson:"name,omitempty"`
	LastName     string               `bson:"lastName,omitempty"`
	Phone        string               `bson:"phone,omitempty"`
	IsInternal   bool                 `bson:"is_internal,omitempty"`
	Profile      primitive.ObjectID   `bson:"profile,omitempty"`
	Company      primitive.ObjectID   `bson:"company"`
	Brands       []primitive.ObjectID `bson:"brands"`
	Countries    []primitive.ObjectID `bson:"countries"`
	Clients      []primitive.ObjectID `bson:"clients"`
}
