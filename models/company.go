package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Company model
type Company struct {
	ID        primitive.ObjectID `bson:"_id"`
	Email     string             `bson:"email"`
	Name      string             `bson:"name"`
	Address   string             `bson:"address"`
	Country   primitive.ObjectID `bson:"country"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty"`
}
