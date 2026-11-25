package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeviceCatalogValue is one selectable value of a device catalog (device
// type, ROM, RAM, memory capability, USB). Values are ordered and
// deactivated instead of deleted.
type DeviceCatalogValue struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Category string             `bson:"category"`
	Value    string             `bson:"value"`
	// ValueKey is the trimmed, lower-cased value backing the unique
	// (category, value) index.
	ValueKey  string    `bson:"value_key"`
	Unit      string    `bson:"unit,omitempty"`
	Active    bool      `bson:"active"`
	Order     int       `bson:"order"`
	CreatedAt time.Time `bson:"created_at"`
}
