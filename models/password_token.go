package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PasswordTokenPurposeInvite = "invite"
	PasswordTokenPurposeReset  = "reset"
)

type PasswordToken struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id"`
	TokenHash string             `bson:"token_hash"`
	Purpose   string             `bson:"purpose"`
	ExpiresAt time.Time          `bson:"expires_at"`
	UsedAt    *time.Time         `bson:"used_at,omitempty"`
	CreatedAt time.Time          `bson:"created_at"`
}
