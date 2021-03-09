package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User model
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Email        string             `bson:"email,omitempty"`
	PasswordHash string             `bson:"passwordHash,omitempty"`
	Name         string             `bson:"name,omitempty"`
	LastName     string             `bson:"lastName,omitempty"`
	Phone        string             `bson:"phone,omitempty"`
	Dni          string             `bson:"dni,omitempty"`
	CreatedAt    time.Time          `bson:"created_at,omitempty"`
	UpdatedAt    time.Time          `bson:"updated_at,omitempty"`
}
