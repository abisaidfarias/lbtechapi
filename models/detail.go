package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User model
type Detail struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	Pages            string             `bson:"pages"`
	Location         string             `bson:"location"`
	MaxTonner        string             `bson:"maxtoner,omitempty"`
	RemToner         string             `bson:"remtoner,omitempty"`
	PercentageTonner string             `bson:"percentage_tonner,omitempty"`
	SNconsumible     string             `bson:"SNconsumible,omitempty"`
	PNconsumible     string             `bson:"PNconsumible,omitempty"`
	CreatedDate      time.Time          `bson:"created_date"`
	Rat              string             `bson:"rat"`
	Level            string             `bson:"level"`
}
