package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User model
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Email        string             `bson:"email"`
	PasswordHash string             `bson:"passwordHash"`
	Name         string             `bson:"name"`
	LastName     string             `bson:"lastName"`
	Phone        string             `bson:"phone"`
	Dni          string             `bson:"dni"`
	CreatedAt    time.Time          `bson:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at,omitempty"`
}
