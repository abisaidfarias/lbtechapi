package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Country model
type Country struct {
	ID        primitive.ObjectID `bson:"_id"`
	Name      string             `bson:"name"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty"`
}
