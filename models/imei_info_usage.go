package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ImeiInfoUsage records one HTTP call to imei.info (paid lookup). Used for rate limits.
type ImeiInfoUsage struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	UserID primitive.ObjectID `bson:"user_id"`
	IMEI   string             `bson:"imei"`
	TAC    string             `bson:"tac"`
	At     time.Time          `bson:"at"`
}
